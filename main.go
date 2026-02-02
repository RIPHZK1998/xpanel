// xpanel - VPN user management backend service
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xpanel/config"
	"xpanel/internal/handler"
	"xpanel/internal/middleware"
	"xpanel/internal/models"
	"xpanel/internal/repository"
	"xpanel/internal/service"
	"xpanel/internal/xray"
	"xpanel/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting xpanel service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	logger.Info("Database connected successfully")

	// Initialize Redis
	redisClient := initRedis(cfg, logger)
	logger.Info("Redis connected successfully")

	// Initialize or load JWT secret from database
	jwtSecret, err := initJWTSecret(db, cfg.JWT.SecretKey, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize JWT secret: %v", err)
	}

	// Initialize or load Node API key from database
	if err := initNodeAPIKey(db, logger); err != nil {
		logger.Fatalf("Failed to initialize Node API key: %v", err)
	}

	// Initialize JWT manager with database-stored secret
	jwtManager := jwt.NewManager(
		jwtSecret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	planRepo := repository.NewPlanRepository(db)
	userSubRepo := repository.NewUserSubscriptionRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	trafficRepo := repository.NewTrafficRepository(db)
	systemConfigRepo := repository.NewSystemConfigRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	// Initialize services
	// Initialize services
	authService := service.NewAuthService(userRepo, subRepo, planRepo, jwtManager, redisClient)
	subscriptionService := service.NewSubscriptionService(subRepo, userRepo, planRepo)
	planService := service.NewPlanService(planRepo, userSubRepo)
	userService := service.NewUserService(userRepo, deviceRepo, subscriptionService)
	nodeService := service.NewNodeService(nodeRepo)
	trafficService := service.NewTrafficService(trafficRepo, subRepo, activityRepo)
	adminService := service.NewAdminService(userRepo, subRepo, userSubRepo, nodeRepo, trafficRepo)
	nodeAgentService := service.NewNodeAgentService(nodeRepo, userRepo, userSubRepo, trafficRepo)
	systemConfigService := service.NewSystemConfigService(systemConfigRepo, jwtSecret)
	activityService := service.NewActivityService(activityRepo, userRepo)

	// Initialize xray manager
	xrayManager := xray.NewManager()

	// Load and register nodes with xray manager
	nodes, err := nodeRepo.GetAll()
	if err != nil {
		logger.Warnf("Failed to load nodes: %v", err)
	} else {
		for _, node := range nodes {
			xrayManager.RegisterNode(&node)
		}
		logger.Infof("Registered %d nodes with xray manager", len(nodes))
	}

	// Create default admin user if none exists
	if err := createDefaultAdmin(userRepo, subRepo, logger); err != nil {
		logger.Warnf("Failed to create default admin: %v", err)
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, trafficService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	configHandler := handler.NewConfigHandler(userService, nodeService, xrayManager)

	// Initialize admin handlers
	adminUserHandler := handler.NewAdminUserHandler(adminService, userService)
	adminUserSubscriptionHandler := handler.NewAdminUserSubscriptionHandler(adminService, userService)
	adminNodeHandler := handler.NewAdminNodeHandler(nodeService, userService, xrayManager)
	adminSubscriptionHandler := handler.NewAdminSubscriptionHandler(subscriptionService)
	adminStatsHandler := handler.NewAdminStatsHandler(adminService)
	adminSystemHandler := handler.NewAdminSystemHandler(systemConfigService)
	adminAdministratorHandler := handler.NewAdminAdministratorHandler(adminService)
	planHandler := handler.NewPlanHandler(planService)

	// Initialize node agent handler
	nodeAgentHandler := handler.NewNodeAgentHandler(nodeAgentService)
	nodeConfigHandler := handler.NewNodeConfigHandler(nodeService)
	activityHandler := handler.NewActivityHandler(activityService)

	// Setup Gin router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Global middleware
	router.Use(middleware.CORS()) // Must be first to handle OPTIONS
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	// Rate limiter (100 requests per minute)
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, 60)
	router.Use(rateLimiter.Middleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Serve static files for frontend
	router.Static("/web", "./web")

	// Redirect root to login
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/login.html")
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// Protected routes
		authMiddleware := middleware.AuthMiddleware(authService)

		user := v1.Group("/user").Use(authMiddleware)
		{
			user.GET("/profile", userHandler.GetProfile)
			user.GET("/devices", userHandler.GetDevices)
			user.DELETE("/devices/:id", userHandler.DeactivateDevice)
			user.GET("/subscription", subscriptionHandler.GetSubscription)
			user.GET("/config", configHandler.GetConfig)
		}

		subscription := v1.Group("/subscription").Use(authMiddleware)
		{
			subscription.POST("/renew", subscriptionHandler.Renew)
		}

		nodes := v1.Group("/nodes").Use(authMiddleware)
		{
			nodes.GET("", nodeHandler.GetNodes)
		}

		// Admin routes (requires both auth and admin role)
		adminMiddleware := middleware.AdminMiddleware(userService)
		admin := v1.Group("/admin").Use(authMiddleware, adminMiddleware)
		{
			// User management
			admin.GET("/users", adminUserHandler.ListUsers)
			admin.POST("/users", adminUserHandler.CreateUser)
			admin.GET("/users/:id", adminUserHandler.GetUser)
			admin.PUT("/users/:id", adminUserHandler.UpdateUser)
			admin.DELETE("/users/:id", adminUserHandler.DeleteUser)
			admin.POST("/users/:id/suspend", adminUserHandler.SuspendUser)
			admin.POST("/users/:id/activate", adminUserHandler.ActivateUser)
			admin.GET("/users/:id/links", adminUserHandler.GetUserLinks)
			admin.PUT("/users/:id/subscription", adminUserSubscriptionHandler.UpdateUserSubscription)

			// Node management
			admin.GET("/nodes", adminNodeHandler.ListNodes)
			admin.GET("/nodes/:id", adminNodeHandler.GetNode)
			admin.POST("/nodes", adminNodeHandler.CreateNode)
			admin.PUT("/nodes/:id", adminNodeHandler.UpdateNode)
			admin.DELETE("/nodes/:id", adminNodeHandler.DeleteNode)
			admin.POST("/nodes/:id/sync", adminNodeHandler.SyncNode)
			admin.GET("/nodes/:id/stats", adminNodeHandler.GetNodeStats)

			// Subscription management
			admin.GET("/subscriptions", adminSubscriptionHandler.ListSubscriptions)
			admin.POST("/subscriptions/:id/extend", adminSubscriptionHandler.ExtendSubscription)
			admin.POST("/subscriptions/:id/reset-data", adminSubscriptionHandler.ResetDataUsage)

			// Plan management (new)
			admin.GET("/plans", planHandler.ListPlans)
			admin.GET("/plans/:id", planHandler.GetPlan)
			admin.POST("/plans", planHandler.CreatePlan)
			admin.PUT("/plans/:id", planHandler.UpdatePlan)
			admin.DELETE("/plans/:id", planHandler.DeletePlan)
			admin.PUT("/plans/:id/nodes", planHandler.AssignNodes)
			admin.GET("/plans/:id/nodes", planHandler.GetPlanNodes)
			admin.GET("/plans/:id/users", planHandler.GetPlanUsers)
			admin.PUT("/users/:id/plan", planHandler.AssignPlanToUser)

			// Statistics
			admin.GET("/stats/overview", adminStatsHandler.GetOverview)

			// Activity monitoring
			admin.GET("/activity/online", activityHandler.GetOnlineUsers)
			admin.GET("/activity/stats", activityHandler.GetStats)

			// System configuration
			admin.GET("/system/config", adminSystemHandler.GetConfig)
			admin.POST("/system/config/reload", adminSystemHandler.ReloadCache)
			admin.GET("/system/config/:key/reveal", adminSystemHandler.RevealConfig)
			admin.PUT("/system/config/:key", adminSystemHandler.UpdateConfig)

			// Administrator management
			admin.GET("/administrators", adminAdministratorHandler.ListAdministrators)
			admin.POST("/administrators", adminAdministratorHandler.CreateAdministrator)
			admin.PUT("/administrators/:id/password", adminAdministratorHandler.ChangePassword)
			admin.DELETE("/administrators/:id", adminAdministratorHandler.DeleteAdministrator)
		}

		// Logout (needs auth)
		v1.POST("/auth/logout", authMiddleware, authHandler.Logout)

		// Node agent routes (for node servers to communicate with panel)
		// Protected with API key authentication
		nodeAuthMiddleware := middleware.NodeAuth(systemConfigService)
		nodeAgent := v1.Group("/node-agent", nodeAuthMiddleware)
		{
			nodeAgent.POST("/heartbeat", nodeAgentHandler.Heartbeat)
			nodeAgent.GET("/:node_id/sync", nodeAgentHandler.SyncUsers)
			nodeAgent.GET("/:node_id/config", nodeConfigHandler.GetNodeConfig)
			nodeAgent.POST("/traffic", nodeAgentHandler.ReportTraffic)
			nodeAgent.POST("/activity", activityHandler.ReportActivity)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Server starting on %s", cfg.Server.Addr())
		// if err := srv.ListenAndServeTLS("./cert.pem", "./key.pem"); err != nil && err != http.ErrServerClosed {
		// 	logger.Fatalf("Failed to start server: %v", err)
		// }
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

// initDatabase initializes the PostgreSQL database connection.
func initDatabase(cfg *config.Config, logger *logrus.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Subscription{},
		&models.SubscriptionPlan{},
		&models.UserSubscription{},
		&models.Node{},
		&models.Device{},
		&models.TrafficLog{},
		&models.UserActivity{},
		&models.SystemConfig{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Migrate legacy subscriptions to new structure
	if err := migrateLegacySubscriptions(db, logger); err != nil {
		logger.Warnf("Failed to migrate legacy subscriptions: %v", err)
	}

	logger.Info("Database migrations completed")

	return db, nil
}

func migrateLegacySubscriptions(db *gorm.DB, logger *logrus.Logger) error {
	// Check if legacy subscriptions exist
	var legacySubs []models.Subscription
	// Use raw SQL to avoid model conflicts if model struct is removed later,
	// but here we know the table exists.
	if err := db.Find(&legacySubs).Error; err != nil {
		return nil // Table might be empty or not exist
	}

	if len(legacySubs) == 0 {
		return nil
	}

	logger.Infof("Found %d legacy subscriptions to migrate", len(legacySubs))

	// Get plans map
	var plans []models.SubscriptionPlan
	if err := db.Find(&plans).Error; err != nil {
		return err
	}

	planMap := make(map[string]uint)
	for _, p := range plans {
		// Map simple names to IDs.
		// Assuming plans "free", "monthly", "yearly" might exist or variants.
		// If using the dump the user imported, names are "free", "monthly_standard", "annual_premium" etc.
		// Legacy plan types were "free", "monthly", "yearly".

		// Map legacy enum -> DB plan name
		// free -> free
		// monthly -> monthly_standard
		// yearly -> annual_premium

		planMap[p.Name] = p.ID
	}

	for _, legacy := range legacySubs {
		// Check if user already has a new subscription
		var count int64
		db.Model(&models.UserSubscription{}).Where("user_id = ?", legacy.UserID).Count(&count)
		if count > 0 {
			continue // Already migrated
		}

		// determine plan ID
		var planID uint
		switch legacy.Plan {
		case models.PlanFree:
			planID = planMap["free"]
		case models.PlanMonthly:
			planID = planMap["monthly_standard"]
			if planID == 0 {
				planID = planMap["monthly"]
			}
		case models.PlanYearly:
			planID = planMap["annual_premium"]
			if planID == 0 {
				planID = planMap["yearly"]
			}
		}

		if planID == 0 {
			// Fallback to free if plan not found
			planID = planMap["free"]
		}

		if planID == 0 {
			logger.Warnf("Skipping migration for user %d: no matching plan found", legacy.UserID)
			continue
		}

		newSub := models.UserSubscription{
			UserID:        legacy.UserID,
			PlanID:        planID,
			Status:        legacy.Status,
			DataUsedBytes: legacy.DataUsedBytes,
			StartDate:     legacy.StartDate,
			ExpiresAt:     legacy.ExpiresAt,
			CreatedAt:     legacy.CreatedAt,
			UpdatedAt:     legacy.UpdatedAt,
		}

		if err := db.Create(&newSub).Error; err != nil {
			logger.Errorf("Failed to migrate subscription for user %d: %v", legacy.UserID, err)
		}
	}

	return nil
}

// initRedis initializes the Redis client.
func initRedis(cfg *config.Config, logger *logrus.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}

// createDefaultAdmin creates a default admin user if none exists.
func createDefaultAdmin(userRepo *repository.UserRepository, subRepo *repository.SubscriptionRepository, logger *logrus.Logger) error {
	// Check if any admin exists
	users, err := userRepo.GetActiveUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Role == models.UserRoleAdmin {
			logger.Info("Admin user already exists")
			return nil
		}
	}

	// Create default admin (no subscription needed for admins)
	password := "admin123" // In production, generate a random password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		Email:        "admin@xpanel.local",
		PasswordHash: string(hashedPassword),
		Role:         models.UserRoleAdmin,
		Status:       models.UserStatusActive,
	}

	if err := userRepo.Create(admin); err != nil {
		return err
	}

	logger.Infof("Created default admin user - Email: %s, Password: %s", admin.Email, password)
	logger.Warn("IMPORTANT: Change the default admin password immediately!")

	return nil
}

// initJWTSecret initializes the JWT secret from database or generates a new one on first startup.
// This ensures the JWT secret is consistent across restarts without requiring .env configuration.
func initJWTSecret(db *gorm.DB, fallbackSecret string, logger *logrus.Logger) (string, error) {
	const jwtSecretKey = "jwt_secret"

	// Check if JWT secret exists in database
	var config models.SystemConfig
	err := db.Where("key = ?", jwtSecretKey).First(&config).Error

	if err == nil {
		// Found existing secret in database
		logger.Info("JWT secret loaded from database")
		return config.Value, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("failed to query JWT secret: %w", err)
	}

	// First startup: generate a new secure JWT secret
	secret, err := generateSecureSecret(64)
	if err != nil {
		// Fallback to env variable if random generation fails
		if fallbackSecret != "" && fallbackSecret != "your-super-secret-key-change-in-production" {
			logger.Warn("Failed to generate random secret, using fallback from environment")
			secret = fallbackSecret
		} else {
			return "", fmt.Errorf("failed to generate JWT secret: %w", err)
		}
	}

	// Store in database
	config = models.SystemConfig{
		Key:         jwtSecretKey,
		Value:       secret,
		Encrypted:   false, // Stored as plain text (database should be secured)
		Description: "JWT signing secret - auto-generated on first startup",
	}

	if err := db.Create(&config).Error; err != nil {
		return "", fmt.Errorf("failed to store JWT secret: %w", err)
	}

	logger.Info("Generated and stored new JWT secret in database")
	return secret, nil
}

// generateSecureSecret generates a cryptographically secure random string of specified length.
func generateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// initNodeAPIKey initializes the Node API key from database or generates a new one on first startup.
// This key is used by agents to authenticate with the panel and can be changed from the Settings page.
func initNodeAPIKey(db *gorm.DB, logger *logrus.Logger) error {
	const nodeAPIKeyKey = "node_api_key"

	// Check if Node API key exists in database
	var config models.SystemConfig
	err := db.Where("key = ?", nodeAPIKeyKey).First(&config).Error

	if err == nil {
		// Found existing key in database
		logger.Info("Node API key loaded from database")
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to query Node API key: %w", err)
	}

	// First startup: generate a new secure Node API key
	apiKey, err := generateSecureSecret(32)
	if err != nil {
		return fmt.Errorf("failed to generate Node API key: %w", err)
	}

	// Store in database (note: will be encrypted by SystemConfigService when updated via UI)
	config = models.SystemConfig{
		Key:         nodeAPIKeyKey,
		Value:       apiKey,
		Encrypted:   false,
		Description: "API key for agent nodes to authenticate with the panel. Change from Settings page.",
	}

	if err := db.Create(&config).Error; err != nil {
		return fmt.Errorf("failed to store Node API key: %w", err)
	}

	logger.Infof("Generated new Node API key (view/change in Settings): %s...", apiKey[:8])
	return nil
}
