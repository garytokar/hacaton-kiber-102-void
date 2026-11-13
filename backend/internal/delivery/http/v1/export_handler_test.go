package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type mockContentRepoForExport struct {
	confirmed []domain.ConfirmedTranslation
}

func (m *mockContentRepoForExport) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	return nil, nil
}
func (m *mockContentRepoForExport) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForExport) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	return nil
}
func (m *mockContentRepoForExport) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	return nil, nil
}
func (m *mockContentRepoForExport) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	return m.confirmed, nil
}
func (m *mockContentRepoForExport) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return nil, nil
}
func (m *mockContentRepoForExport) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForExport) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	return nil
}


func TestExportHandler_ExportTranslations(t *testing.T) {
	fragID := uuid.New()
	confirmedList := []domain.ConfirmedTranslation{
		{
			ID:             uuid.New(),
			FragmentID:     fragID,
			TextBa:         "Ағиҙел йылғаһы — Башҡортостандың иң ҙур һыу юлы.",
			LangPair:       "ba-ru",
			Translation:    "Река Белая — крупнейшая водная артерия Башкортостана.",
			SourceCoord:    "География Башкортостана, стр. 42",
			ConfirmedCount: 5,
			Status:         "confirmed",
		},
	}

	contentRepo := &mockContentRepoForExport{confirmed: confirmedList}
	exportUC := usecase.NewExportUseCase(contentRepo)
	handler := v1.NewExportHandler(exportUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/translations?lang_pair=ba-ru", nil)
	rec := httptest.NewRecorder()

	handler.ExportTranslations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var resp []domain.ConfirmedTranslation
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("ожидалась 1 запись, получено %d", len(resp))
	}
	// Проверка соответствия разделу 5 исходного ТЗ
	entry := resp[0]
	if entry.FragmentID != fragID {
		t.Errorf("некорректный fragment_id")
	}
	if entry.LangPair != "ba-ru" {
		t.Errorf("некорректная языковая пара")
	}
	if entry.ConfirmedCount != 5 {
		t.Errorf("неверное количество подтверждений")
	}
	if entry.Status != "confirmed" {
		t.Errorf("статус должен быть confirmed")
	}
}


func (m *mockContentRepoForExport) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	return nil, nil
}
