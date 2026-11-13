package usecase

import (
	"context"

	"bashkir-backend/internal/domain"
)

type ExportUseCase struct {
	contentRepo domain.ContentRepository
}

func NewExportUseCase(contentRepo domain.ContentRepository) *ExportUseCase {
	return &ExportUseCase{contentRepo: contentRepo}
}

// ExportConfirmedTranslations выгрузка согласованной базы парных переводов по разделу 5 ТЗ
func (uc *ExportUseCase) ExportConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	if langPair == "" {
		langPair = "ba-ru"
	}
	return uc.contentRepo.GetConfirmedTranslations(ctx, langPair)
}
