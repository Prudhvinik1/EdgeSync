package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/prudhvinik1/edgesync/internal/models"
	"github.com/prudhvinik1/edgesync/internal/repositories"
	"github.com/prudhvinik1/edgesync/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	accountRepo repositories.AccountRepository
	deviceRepo  repositories.DeviceRepository
	sessionRepo repositories.SessionRepository
	jwtSecret   string
	jwtExpiry   time.Duration
}

type LoginRequest struct {
	Email      string
	Password   string
	DeviceID   *uuid.UUID // Optional - nil means create new device
	DeviceName string
	DeviceType string
}

type LoginResponse struct {
	Token     string
	ExpiresAt time.Time
	DeviceID  uuid.UUID
	AccountID uuid.UUID
}

type TokenClaims struct {
	AccountID uuid.UUID
	DeviceID  uuid.UUID
	SessionID string
}

func NewAuthService(
	accountRepo repositories.AccountRepository,
	deviceRepo repositories.DeviceRepository,
	sessionRepo repositories.SessionRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		accountRepo: accountRepo,
		deviceRepo:  deviceRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	// Check if email already exists
	existing, err := s.accountRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return ErrEmailExists
	}
	if err != nil && err != repositories.ErrNotFound {
		return fmt.Errorf("failed to check email: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create account
	account := &models.Account{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	err = s.accountRepo.Create(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Validate credentials
	account, err := s.accountRepo.GetByEmail(ctx, req.Email)
	if err == repositories.ErrNotFound {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if !utils.CheckPassword(account.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	// Handle device
	var device *models.Device
	if req.DeviceID != nil {
		// Use existing device
		device, err = s.deviceRepo.GetByID(ctx, *req.DeviceID)
		if err == repositories.ErrNotFound {
			return nil, errors.New("device not found")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get device: %w", err)
		}
		if device.AccountID != account.ID {
			return nil, errors.New("device does not belong to account")
		}
	} else {
		// Create new device
		device = &models.Device{
			AccountID:  account.ID,
			Name:       req.DeviceName,
			DeviceType: req.DeviceType,
		}
		err = s.deviceRepo.Create(ctx, device)
		if err != nil {
			return nil, fmt.Errorf("failed to create device: %w", err)
		}
	}

	// Create session
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(s.jwtExpiry)
	session := &models.Session{
		ID:        sessionID,
		AccountID: account.ID,
		DeviceID:  device.ID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate token
	token, err := s.generateToken(account.ID, device.ID, sessionID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		AccountID: account.ID,
		DeviceID:  device.ID,
	}, nil
}

func (s *AuthService) generateToken(accountID, deviceID uuid.UUID, sessionID string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":       accountID.String(),
		"device_id": deviceID.String(),
		"jti":       sessionID,
		"exp":       expiresAt.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) VerifyToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract account ID
	accountIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Extract device ID
	deviceIDStr, ok := claims["device_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Extract session ID
	sessionID, ok := claims["jti"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &TokenClaims{
		AccountID: accountID,
		DeviceID:  deviceID,
		SessionID: sessionID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.VerifyToken(tokenString)
	if err != nil {
		return err
	}

	// Delete session using session ID from token
	err = s.sessionRepo.Delete(ctx, claims.SessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, tokenString string) error {
	claims, err := s.VerifyToken(tokenString)
	if err != nil {
		return err
	}

	err = s.sessionRepo.DeleteAllForAccount(ctx, claims.AccountID)
	if err != nil {
		return fmt.Errorf("failed to logout all sessions: %w", err)
	}

	return nil
}
