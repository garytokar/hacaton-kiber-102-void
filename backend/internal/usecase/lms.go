package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bashkir-backend/internal/domain"

	"github.com/google/uuid"
)

type LMSUseCase struct {
	lmsRepo  domain.LMSRepository
	userRepo domain.UserRepository
}

func NewLMSUseCase(lmsRepo domain.LMSRepository, userRepo ...domain.UserRepository) *LMSUseCase {
	var uRepo domain.UserRepository
	if len(userRepo) > 0 {
		uRepo = userRepo[0]
	}
	return &LMSUseCase{lmsRepo: lmsRepo, userRepo: uRepo}
}

func (uc *LMSUseCase) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.lmsRepo.GetSchoolsLeaderboard(ctx, limit)
}

func (uc *LMSUseCase) GetClassesLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardClassEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.lmsRepo.GetClassesLeaderboard(ctx, limit)
}

func (uc *LMSUseCase) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	return uc.lmsRepo.GetClassStudents(ctx, classID)
}

// CreateClass создает новый класс учителем и генерирует уникальный код приглашения
func (uc *LMSUseCase) CreateClass(ctx context.Context, schoolID uuid.UUID, gradeNumber int, letter string) (*domain.Class, error) {
	if gradeNumber < 1 || gradeNumber > 11 {
		return nil, errors.New("номер класса должен быть от 1 до 11")
	}
	letter = strings.TrimSpace(letter)
	if letter == "" {
		letter = "А"
	}

	classID := uuid.New()
	code := fmt.Sprintf("INV-%d%s-%s", gradeNumber, letter, strings.ToUpper(uuid.New().String()[:6]))

	class := &domain.Class{
		ID:          classID,
		SchoolID:    schoolID,
		GradeNumber: gradeNumber,
		Letter:      letter,
		InviteCode:  &code,
		CreatedAt:   time.Now(),
	}

	if err := uc.lmsRepo.CreateClass(ctx, class); err != nil {
		return nil, err
	}
	return class, nil
}

// GenerateInviteCode генерирует новый инвайт-код для класса
func (uc *LMSUseCase) GenerateInviteCode(ctx context.Context, classID uuid.UUID) (string, error) {
	class, err := uc.lmsRepo.GetClassByID(ctx, classID)
	if err != nil {
		return "", err
	}

	code := fmt.Sprintf("INV-%d%s-%s", class.GradeNumber, class.Letter, strings.ToUpper(uuid.New().String()[:6]))
	if err := uc.lmsRepo.UpdateClassInviteCode(ctx, classID, code); err != nil {
		return "", err
	}
	return code, nil
}

// JoinClassByCode присоединяет ученика к классу по коду приглашения
func (uc *LMSUseCase) JoinClassByCode(ctx context.Context, userID uuid.UUID, inviteCode string) (*domain.Class, error) {
	cleanCode := strings.TrimSpace(inviteCode)
	if cleanCode == "" {
		return nil, errors.New("код приглашения не может быть пустым")
	}

	class, err := uc.lmsRepo.GetClassByInviteCode(ctx, cleanCode)
	if err != nil || class == nil {
		return nil, errors.New("класс с таким кодом приглашения не найден")
	}

	if err := uc.lmsRepo.AssignUserToClass(ctx, userID, class.ID, class.SchoolID); err != nil {
		return nil, err
	}

	return class, nil
}

// AssignStudentToClass прикрепляет ученика к классу
func (uc *LMSUseCase) AssignStudentToClass(ctx context.Context, userID, classID uuid.UUID) (*domain.Class, error) {
	class, err := uc.lmsRepo.GetClassByID(ctx, classID)
	if err != nil {
		return nil, err
	}

	if err := uc.lmsRepo.AssignUserToClass(ctx, userID, class.ID, class.SchoolID); err != nil {
		return nil, err
	}
	return class, nil
}

// RegisterStudentByTeacher регистрирует нового ученика учителем и прикрепляет к классу
func (uc *LMSUseCase) RegisterStudentByTeacher(ctx context.Context, classID uuid.UUID, fullName, email, password string) (*domain.User, *domain.Class, error) {
	class, err := uc.lmsRepo.GetClassByID(ctx, classID)
	if err != nil {
		return nil, nil, err
	}

	cleanEmail := strings.TrimSpace(email)
	if cleanEmail == "" {
		return nil, nil, errors.New("email обязателен для регистрации ученика")
	}

	cleanName := strings.TrimSpace(fullName)
	if cleanName == "" {
		cleanName = "Ученик"
	}

	if uc.userRepo == nil {
		return nil, class, nil
	}

	existing, _ := uc.userRepo.FindByEmail(ctx, cleanEmail)
	if existing != nil {
		if err := uc.lmsRepo.AssignUserToClass(ctx, existing.ID, class.ID, class.SchoolID); err != nil {
			return nil, nil, err
		}
		existing.ClassID = &class.ID
		existing.SchoolID = &class.SchoolID
		return existing, class, nil
	}

	if password == "" {
		password = "StudentPass123!"
	}

	user := &domain.User{
		ID:                uuid.New(),
		Email:             &cleanEmail,
		FullName:          cleanName,
		Role:              domain.RoleStudent,
		PreferredLangPair: "ba-ru",
		SchoolID:          &class.SchoolID,
		ClassID:           &class.ID,
		CurrentStreak:     0,
		MaxStreak:         0,
		TotalXP:           0,
		CreatedAt:         time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	return user, class, nil
}

