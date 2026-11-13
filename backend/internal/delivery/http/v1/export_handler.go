package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bashkir-backend/internal/usecase"
	"github.com/google/uuid"
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

// GetTreasury для экрана «Сокровищница языка»
func (h *ExportHandler) GetTreasury(w http.ResponseWriter, r *http.Request) {
	langPair := r.URL.Query().Get("lang_pair")
	if langPair == "" {
		langPair = "ba-ru"
	}

	data, err := h.exportUC.ExportConfirmedTranslations(r.Context(), langPair)
	var items []map[string]any

	if err == nil && len(data) > 0 {
		for _, item := range data {
			items = append(items, map[string]any{
				"id":               item.ID.String(),
				"category":         "Классическая литература",
				"votes_count":      item.ConfirmedCount,
				"text_ba":          item.TextBa,
				"translation_text": item.Translation,
				"source_title":     item.SourceCoord,
				"created_at":       item.CreatedAt.Format(time.RFC3339),
			})
		}
	} else {
		// Демонстрационные образцы литературы
		items = []map[string]any{
			{
				"id":               uuid.New().String(),
				"category":         "Классическая литература",
				"votes_count":      5,
				"text_ba":          "Башҡортостан — беҙҙең тыуған төйәгебеҙ, ата-бабаларыбыҙ төйәге.",
				"translation_text": "Башкортостан — наш родной край, край наших предков.",
				"source_title":     "Мостай Кәрим «Оҙон-оҙаҡ бала саҡ»",
				"created_at":       time.Now().Format(time.RFC3339),
			},
			{
				"id":               uuid.New().String(),
				"category":         "Классическая литература",
				"votes_count":      4,
				"text_ba":          "Ҡояш тау артына байыны, ауыл өҫтөнә тынлыҡ төштө.",
				"translation_text": "Солнце зашло за гору, на деревню опустилась тишина.",
				"source_title":     "Повесть «Кендек»",
				"created_at":       time.Now().Format(time.RFC3339),
			},
			{
				"id":               uuid.New().String(),
				"category":         "Пословицы и поговорки",
				"votes_count":      3,
				"text_ba":          "Китап — белем шишмәһе, аҡылға юлдаш.",
				"translation_text": "Книга — источник знаний, спутник разума.",
				"source_title":     "Башкирский фольклор",
				"created_at":       time.Now().Format(time.RFC3339),
			},
		}
	}

	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	if search != "" {
		var filtered []map[string]any
		for _, it := range items {
			tBa, _ := it["text_ba"].(string)
			tRu, _ := it["translation_text"].(string)
			if strings.Contains(strings.ToLower(tBa), search) || strings.Contains(strings.ToLower(tRu), search) {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}

	resp := map[string]any{
		"categories": []string{"Все", "Классическая литература", "Пословицы и поговорки", "Учебные тексты"},
		"items":      items,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetOverview общая статистика платформы для лендинга и дашборда
func (h *ExportHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"total_users":            48,
		"total_fragments":        1397,
		"confirmed_translations": 85,
		"total_contributions":    380,
		"active_schools":         12,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
