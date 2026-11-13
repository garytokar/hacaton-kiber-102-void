package usecase

import (
	"context"
	"fmt"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/pkg/hasher"
	"bashkir-backend/internal/pkg/jwt"
	"bashkir-backend/internal/pkg/tghash"

	"github.com/google/uuid"
)

type AuthUseCase struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

func NewAuthUseCase(userRepo domain.UserRepository, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo, cfg: cfg}
}

// RegisterEmail регистрация по email/паролю для учителей и экспертов
func (uc *AuthUseCase) RegisterEmail(ctx context.Context, fullName, email, password string, role domain.UserRole, schoolID, classID *uuid.UUID) (*domain.User, error) {
	existing, _ := uc.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hash, err := hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:             uuid.New(),
		Email:          &email,
		HashedPassword: &hash,
		FullName:       fullName,
		Role:           role,
		SchoolID:       schoolID,
		ClassID:        classID,
		CurrentStreak:  0,
		MaxStreak:      0,
		TotalXP:        0,
		CreatedAt:      time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// LoginEmail вход по email/паролю
func (uc *AuthUseCase) LoginEmail(ctx context.Context, email, password string) (*domain.AuthTokenResponse, *domain.User, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if user.HashedPassword == nil || !hasher.CheckPassword(password, *user.HashedPassword) {
		return nil, nil, domain.ErrInvalidCredentials
	}

	token, exp, err := jwt.GenerateToken(user.ID, string(user.Role), uc.cfg.JWTSecret, uc.cfg.JWTTTL)
	if err != nil {
		return nil, nil, err
	}

	return &domain.AuthTokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresAt:   exp,
	}, user, nil
}

// AuthTelegramWebApp бесшовная авторизация из Telegram Mini App
func (uc *AuthUseCase) AuthTelegramWebApp(ctx context.Context, initDataRaw string) (*domain.AuthTokenResponse, *domain.User, error) {
	tgUser, err := tghash.ValidateInitData(initDataRaw, uc.cfg.TelegramBotToken)
	if err != nil {
		return nil, nil, domain.ErrInvalidTelegramHash
	}

	user, err := uc.GetOrCreateTgUser(ctx, tgUser.ID, tgUser.Username, fmt.Sprintf("%s %s", tgUser.FirstName, tgUser.LastName))
	if err != nil {
		return nil, nil, err
	}

	token, exp, err := jwt.GenerateToken(user.ID, string(user.Role), uc.cfg.JWTSecret, uc.cfg.JWTTTL)
	if err != nil {
		return nil, nil, err
	}

	return &domain.AuthTokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresAt:   exp,
	}, user, nil
}

// GetOrCreateTgUser находит или создает пользователя по Telegram ID
func (uc *AuthUseCase) GetOrCreateTgUser(ctx context.Context, tgID int64, username, fullName string) (*domain.User, error) {
	user, err := uc.userRepo.FindByTgID(ctx, tgID)
	if err == nil && user != nil {
		return user, nil
	}

	name := fullName
	if name == "" || name == " " {
		name = username
	}
	if name == "" {
		name = "Ученик"
	}

	var uname *string
	if username != "" {
		uname = &username
	}

	newUser := &domain.User{
		ID:            uuid.New(),
		TgID:          &tgID,
		TgUsername:    uname,
		FullName:      name,
		Role:          domain.RoleStudent,
		CurrentStreak: 0,
		MaxStreak:     0,
		TotalXP:       0,
		CreatedAt:     time.Now(),
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}
