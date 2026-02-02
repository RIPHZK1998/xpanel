// Package service contains business logic implementations.
package service

import (
	"context"
	"errors"
	"time"

	"xpanel/internal/models"
	"xpanel/internal/repository"
	"xpanel/pkg/jwt"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// Auth errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrTokenBlacklisted   = errors.New("token has been revoked")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// AuthService handles authentication-related business logic.
type AuthService struct {
	userRepo    *repository.UserRepository
	subRepo     *repository.SubscriptionRepository
	planRepo    *repository.PlanRepository
	jwtManager  *jwt.Manager
	redisClient *redis.Client
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	userRepo *repository.UserRepository,
	subRepo *repository.SubscriptionRepository,
	planRepo *repository.PlanRepository,
	jwtManager *jwt.Manager,
	redisClient *redis.Client,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		subRepo:     subRepo,
		planRepo:    planRepo,
		jwtManager:  jwtManager,
		redisClient: redisClient,
	}
}

// RegisterRequest contains user registration data.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest contains user login credentials.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token TTL in seconds
}

// Register creates a new user account with a free subscription.
func (s *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	// Check if email is already taken
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Find free plan
	freePlan, err := s.planRepo.GetByName("free")
	if err != nil {
		// Log error but don't fail hard if free plan is missing, create user without sub
		// (Or handle graceful fallback)
		// For now, let's just proceed without subscription or return valid user
		return user, nil
	}

	// Create free subscription for new user
	subscription := &models.UserSubscription{
		UserID:    user.ID,
		PlanID:    freePlan.ID,
		Status:    models.SubscriptionActive,
		StartDate: time.Now(),
		AutoRenew: false,
	}

	if err := s.subRepo.Create(subscription); err != nil {
		// Rollback user creation on subscription failure
		_ = s.userRepo.Delete(user.ID)
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns token pair.
func (s *AuthService) Login(req *LoginRequest) (*TokenPair, *models.User, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil, nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

// RefreshToken generates a new access token using a valid refresh token.
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check token type
	if claims.TokenType != jwt.RefreshToken {
		return nil, ErrInvalidToken
	}

	// Check if token is blacklisted
	if s.isTokenBlacklisted(refreshToken) {
		return nil, ErrTokenBlacklisted
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if user.Status != models.UserStatusActive {
		return nil, ErrInvalidCredentials
	}

	// Blacklist old refresh token
	_ = s.blacklistToken(refreshToken, s.jwtManager.GetRefreshTokenTTL())

	// Generate new token pair
	return s.generateTokenPair(user)
}

// Logout invalidates the user's tokens.
func (s *AuthService) Logout(accessToken, refreshToken string) error {
	ctx := context.Background()

	// Blacklist access token
	if accessToken != "" {
		_ = s.redisClient.Set(ctx, "blacklist:"+accessToken, "1", s.jwtManager.GetAccessTokenTTL()).Err()
	}

	// Blacklist refresh token
	if refreshToken != "" {
		_ = s.redisClient.Set(ctx, "blacklist:"+refreshToken, "1", s.jwtManager.GetRefreshTokenTTL()).Err()
	}

	return nil
}

// ValidateAccessToken validates an access token and returns the claims.
func (s *AuthService) ValidateAccessToken(token string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != jwt.AccessToken {
		return nil, ErrInvalidToken
	}

	if s.isTokenBlacklisted(token) {
		return nil, ErrTokenBlacklisted
	}

	return claims, nil
}

// generateTokenPair creates a new access/refresh token pair.
func (s *AuthService) generateTokenPair(user *models.User) (*TokenPair, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenTTL().Seconds()),
	}, nil
}

// isTokenBlacklisted checks if a token is in the blacklist.
func (s *AuthService) isTokenBlacklisted(token string) bool {
	ctx := context.Background()
	result, err := s.redisClient.Exists(ctx, "blacklist:"+token).Result()
	if err != nil {
		return false // Fail open in case of Redis error
	}
	return result > 0
}

// blacklistToken adds a token to the blacklist.
func (s *AuthService) blacklistToken(token string, ttl time.Duration) error {
	ctx := context.Background()
	return s.redisClient.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}
