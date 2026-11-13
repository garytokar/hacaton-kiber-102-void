package v1

import (
	"encoding/json"
	"net/http"

	"bashkir-backend/internal/usecase"
)

type ExportHandler struct {
	exportUC *usecase.ExportUseCase
}

func NewExportHandler(exportUC *usecase.ExportUseCase) *ExportHandler {
	return &ExportHandler{exportUC: exportUC}
}

// ExportTranslations эндпоинт для выгрузки подтвержденной базы переводов по разделу 5 ТЗ
func (h *ExportHandler) ExportTranslations(w http.ResponseWriter, r *http.Request) {
	langPair := r.URL.Query().Get("lang_pair")
	if langPair == "" {
		langPair = "ba-ru"
	}

	data, err := h.exportUC.ExportConfirmedTranslations(r.Context(), langPair)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
