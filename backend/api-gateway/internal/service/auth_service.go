// Package service provides business logic for API Gateway
package service

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/auth"
	ldapauth "github.com/wangjialin/myops/pkg/auth/ldap"
	"github.com/wangjialin/myops/pkg/auth/jwt"
	"github.com/wangjialin/myops/pkg/auth/redis"
	"github.com/wangjialin/myops/pkg/db"
	apperrors "github.com/wangjialin/myops/pkg/errors"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	User *UserResponse `json:"user"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int          `json:"expiresIn"`
	User         *UserResponse `json:"user"`
}

// LDAPLoginRequest represents an LDAP login request
type LDAPLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// UserResponse represents user information in responses
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// AuthService handles authentication operations
type AuthService struct {
	userRepo   *db.UserRepository
	jwtManager *jwt.Manager
	tokenRepo  *redis.RefreshTokenRepository
	ldapClient *ldapauth.Client
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo *db.UserRepository, jwtManager *jwt.Manager, tokenRepo *redis.RefreshTokenRepository, ldapClient *ldapauth.Client) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		tokenRepo:  tokenRepo,
		ldapClient: ldapClient,
	}
}

// Register handles user registration
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1. Validate input
	if err := auth.UsernameFormat(req.Username); err != nil {
		return nil, apperrors.Wrap("INVALID_USERNAME", err.Error(), err)
	}
	if err := auth.EmailFormat(req.Email); err != nil {
		return nil, apperrors.Wrap("INVALID_EMAIL", err.Error(), err)
	}
	if err := auth.PasswordStrength(req.Password); err != nil {
		return nil, apperrors.Wrap("WEAK_PASSWORD", err.Error(), err)
	}

	// 2. Check uniqueness
	exists, err := s.userRepo.ExistsUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrUsernameExists
	}

	exists, err = s.userRepo.ExistsEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrEmailExists
	}

	// 3. Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 4. Create user
	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		UserType:     model.UserTypeLocal,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5. Return response (without password)
	return &RegisterResponse{
		User: &UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// Login handles user login
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 1. Find user
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, err
	}

	// 2. Verify password (return same error whether user not found or password wrong)
	if user == nil || !auth.ComparePassword(user.PasswordHash, req.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	// 3. Generate Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username)
	if err != nil {
		return nil, err
	}

	// 4. Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	// 5. Validate Refresh Token to get tokenID
	tokenID, userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 6. Store Refresh Token in Redis (30 days)
	err = s.tokenRepo.Store(ctx, userID, tokenID, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	// 7. Return response
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		User: &UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// LDAPLogin handles LDAP user authentication and login
func (s *AuthService) LDAPLogin(ctx context.Context, req *LDAPLoginRequest) (*LoginResponse, error) {
	// 1. Check LDAP client is available
	if s.ldapClient == nil {
		return nil, apperrors.NewError("LDAP_NOT_CONFIGURED", "LDAP authentication is not configured")
	}

	// 2. Authenticate with LDAP
	ldapEntry, err := s.ldapClient.Authenticate(req.Username, req.Password)
	if err != nil {
		return nil, apperrors.Wrap("LDAP_AUTH_FAILED", "LDAP authentication failed: "+err.Error(), err)
	}

	// 3. Extract user information from LDAP
	ldapUsername := ldapauth.GetUserAttribute(ldapEntry, "uid")
	ldapEmail := ldapauth.GetUserAttribute(ldapEntry, "mail")
	ldapDisplayName := ldapauth.GetUserAttribute(ldapEntry, "cn")

	// Use provided username as fallback
	if ldapUsername == "" {
		ldapUsername = req.Username
	}
	if ldapEmail == "" {
		ldapEmail = req.Username + "@ldap.local"
	}

	// 4. Check if user exists in local database
	user, err := s.userRepo.FindByUsername(ctx, ldapUsername)
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 5. Create or update user in local database
	if user == nil {
		// Create new LDAP user
		user = &model.User{
			ID:       uuid.New(),
			Username: ldapUsername,
			Email:    ldapEmail,
			UserType: model.UserTypeLDAP,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	} else {
		// Update existing LDAP user email if changed
		if user.Email != ldapEmail {
			user.Email = ldapEmail
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, err
			}
		}
	}

	// 6. Generate Access Token
	displayName := ldapDisplayName
	if displayName == "" {
		displayName = ldapUsername
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), displayName)
	if err != nil {
		return nil, err
	}

	// 7. Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	// 8. Validate Refresh Token to get tokenID
	tokenID, userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 9. Store Refresh Token in Redis (30 days)
	err = s.tokenRepo.Store(ctx, userID, tokenID, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	// 10. Return response
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		User: &UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// RefreshToken handles token refresh
func (s *AuthService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// 1. Validate Refresh Token
	tokenID, userID, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	// 2. Check Redis storage
	valid, err := s.tokenRepo.Verify(ctx, userID, tokenID)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, apperrors.ErrUnauthorized
	}

	// 3. Get user info
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// 4. Generate new Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username)
	if err != nil {
		return nil, err
	}

	// 5. Return response (keeping same refresh token)
	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		ExpiresIn:    3600,
	}, nil
}

// FindByUsername finds a user by username
func (s *AuthService) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return user, nil
}
