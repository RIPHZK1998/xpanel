package main

import (
	"log"

	"xpanel/config"
	"xpanel/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Migration script to add subscription plan tables and migrate existing data
func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database using application config
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Starting migration...")

	// Step 1: Create new tables
	log.Println("Creating new tables...")
	if err := db.AutoMigrate(
		&models.SubscriptionPlan{},
		&models.UserSubscription{},
	); err != nil {
		log.Fatal("Failed to create new tables:", err)
	}

	// Step 2: Create default subscription plans
	log.Println("Creating default subscription plans...")
	plans := []models.SubscriptionPlan{
		{
			Name:        "free",
			DisplayName: "Free Plan",
			Duration:    models.DurationMonthly,
			Price:       0,
			DataLimitGB: 5,
			MaxDevices:  2,
			Status:      models.PlanStatusActive,
			Description: "Basic free plan with limited data",
			Features:    "5GB monthly data, 2 devices, Basic support",
		},
		{
			Name:        "weekly_basic",
			DisplayName: "Weekly Basic",
			Duration:    models.DurationWeekly,
			Price:       5,
			DataLimitGB: 50,
			MaxDevices:  3,
			Status:      models.PlanStatusActive,
			Description: "Perfect for short-term needs",
			Features:    "50GB weekly data, 3 devices, Priority support",
		},
		{
			Name:        "monthly_standard",
			DisplayName: "Monthly Standard",
			Duration:    models.DurationMonthly,
			Price:       15,
			DataLimitGB: 200,
			MaxDevices:  5,
			Status:      models.PlanStatusActive,
			Description: "Most popular plan for regular users",
			Features:    "200GB monthly data, 5 devices, Priority support, All nodes",
		},
		{
			Name:        "quarterly_pro",
			DisplayName: "Quarterly Pro",
			Duration:    models.DurationQuarterly,
			Price:       40,
			DataLimitGB: 600,
			MaxDevices:  10,
			Status:      models.PlanStatusActive,
			Description: "Best value for power users",
			Features:    "600GB quarterly data, 10 devices, Premium support, All nodes, Priority routing",
		},
		{
			Name:        "annual_premium",
			DisplayName: "Annual Premium",
			Duration:    models.DurationAnnual,
			Price:       150,
			DataLimitGB: 0, // Unlimited
			MaxDevices:  20,
			Status:      models.PlanStatusActive,
			Description: "Ultimate plan with unlimited data",
			Features:    "Unlimited data, 20 devices, 24/7 Premium support, All nodes, Priority routing, Dedicated IP option",
		},
	}

	for _, plan := range plans {
		var existing models.SubscriptionPlan
		if err := db.Where("name = ?", plan.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&plan).Error; err != nil {
				log.Printf("Failed to create plan %s: %v", plan.Name, err)
			} else {
				log.Printf("Created plan: %s", plan.DisplayName)
			}
		} else {
			log.Printf("Plan %s already exists, skipping", plan.Name)
		}
	}

	// Step 3: Migrate existing subscriptions to user_subscriptions
	log.Println("Migrating existing subscriptions...")

	var oldSubscriptions []models.Subscription
	if err := db.Find(&oldSubscriptions).Error; err != nil {
		log.Printf("Warning: Could not load old subscriptions: %v", err)
	} else {
		log.Printf("Found %d existing subscriptions to migrate", len(oldSubscriptions))

		for _, oldSub := range oldSubscriptions {
			// Map old plan type to new plan
			var planName string
			switch oldSub.Plan {
			case models.PlanFree:
				planName = "free"
			case models.PlanMonthly:
				planName = "monthly_standard"
			case models.PlanYearly:
				planName = "annual_premium"
			default:
				planName = "free"
			}

			// Find the corresponding plan
			var plan models.SubscriptionPlan
			if err := db.Where("name = ?", planName).First(&plan).Error; err != nil {
				log.Printf("Warning: Could not find plan %s for user %d", planName, oldSub.UserID)
				continue
			}

			// Check if user subscription already exists
			var existing models.UserSubscription
			if err := db.Where("user_id = ?", oldSub.UserID).First(&existing).Error; err == gorm.ErrRecordNotFound {
				// Create new user subscription
				userSub := models.UserSubscription{
					UserID:        oldSub.UserID,
					PlanID:        plan.ID,
					Status:        oldSub.Status,
					DataUsedBytes: oldSub.DataUsedBytes,
					StartDate:     oldSub.StartDate,
					ExpiresAt:     oldSub.ExpiresAt,
					AutoRenew:     false,
					CreatedAt:     oldSub.CreatedAt,
					UpdatedAt:     oldSub.UpdatedAt,
				}

				if err := db.Create(&userSub).Error; err != nil {
					log.Printf("Failed to migrate subscription for user %d: %v", oldSub.UserID, err)
				} else {
					log.Printf("Migrated subscription for user %d to plan %s", oldSub.UserID, plan.DisplayName)
				}
			} else {
				log.Printf("User %d already has a subscription, skipping", oldSub.UserID)
			}
		}
	}

	log.Println("Migration completed successfully!")
	log.Println("\nNext steps:")
	log.Println("1. Assign nodes to plans via admin interface")
	log.Println("2. Test the new subscription system")
	log.Println("3. Once verified, you can drop the old 'subscriptions' table")
}
