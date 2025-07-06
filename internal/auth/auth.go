package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/rusiqe/domainvault/internal/types"
)

// AuthService handles authentication operations
type AuthService struct {
	repo UserRepository
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(user *types.User) error
	GetUserByUsername(username string) (*types.User, error)
	GetUserByID(id string) (*types.User, error)
	UpdateUser(user *types.User) error
	DeleteUser(id string) error
	
	CreateSession(session *types.Session) error
	GetSessionByToken(token string) (*types.Session, error)
	DeleteSession(token string) error
	DeleteExpiredSessions() error
	UpdateLastLogin(userID string) error
}

// NewAuthService creates a new authentication service
func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

// Login authenticates a user and creates a session
func (a *AuthService) Login(username, password string) (*types.LoginResponse, error) {
	// Get user by username
	user, err := a.repo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, fmt.Errorf("account disabled")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create session (24 hours expiry)
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &types.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := a.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := a.repo.UpdateLastLogin(user.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login for user %s: %v\n", user.ID, err)
	}

	return &types.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

// ValidateToken validates a session token and returns the user
func (a *AuthService) ValidateToken(token string) (*types.User, error) {
	// Get session by token
	session, err := a.repo.GetSessionByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		a.repo.DeleteSession(token)
		return nil, fmt.Errorf("token expired")
	}

	// Get user
	user, err := a.repo.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still enabled
	if !user.Enabled {
		return nil, fmt.Errorf("account disabled")
	}

	return user, nil
}

// Logout removes a user session
func (a *AuthService) Logout(token string) error {
	return a.repo.DeleteSession(token)
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CleanupExpiredSessions removes expired sessions
func (a *AuthService) CleanupExpiredSessions() error {
	return a.repo.DeleteExpiredSessions()
}

// generateToken creates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateDefaultAdmin creates the default admin user if it doesn't exist
func (a *AuthService) CreateDefaultAdmin() error {
	// Check if admin user already exists
	_, err := a.repo.GetUserByUsername("admin")
	if err == nil {
		// Admin already exists
		return nil
	}

	// Hash default password
	hashedPassword, err := HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	admin := &types.User{
		Username:     "admin",
		Email:        "admin@domainvault.local",
		PasswordHash: hashedPassword,
		Role:         "admin",
		Enabled:      true,
	}

	return a.repo.CreateUser(admin)
}

// VerifyCurrentUserPassword verifies the current user's password for security operations
func (a *AuthService) VerifyCurrentUserPassword(userID, password string) bool {
	user, err := a.repo.GetUserByID(userID)
	if err != nil {
		return false
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
