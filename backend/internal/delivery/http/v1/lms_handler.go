package v1

import (
	"encoding/json"
	"net/http"

	"bashkir-backend/internal/delivery/http/middleware"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LMSHandler struct {
	lmsUC *usecase.LMSUseCase
}

func NewLMSHandler(lmsUC *usecase.LMSUseCase) *LMSHandler {
	return &LMSHandler{lmsUC: lmsUC}
}

func (h *LMSHandler) GetClassStats(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "class_id")
	if classIDStr == "" {
		classIDStr = chi.URLParam(r, "id")
	}
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid class_id"}`, http.StatusBadRequest)
		return
	}

	students, err := h.lmsUC.GetClassStudents(r.Context(), classID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(students)
}

func (h *LMSHandler) GetSchoolsOptions(w http.ResponseWriter, r *http.Request) {
	schools, err := h.lmsUC.GetSchoolsLeaderboard(r.Context(), 50)
	var result []map[string]any
	if err == nil && len(schools) > 0 {
		for i, s := range schools {
			result = append(result, map[string]any{
				"id":   i + 1,
				"name": s.SchoolName,
				"city": s.City,
				"classes": []map[string]any{
					{"id": i*10 + 1, "name": "9 А"},
					{"id": i*10 + 2, "name": "10 Б"},
					{"id": i*10 + 3, "name": "11 А"},
				},
			})
		}
	} else {
		result = []map[string]any{
			{
				"id":   1,
				"name": "Башкирская республиканская гимназия-интернат № 1 им. Р. Гарипова",
				"city": "г. Уфа",
				"classes": []map[string]any{
					{"id": 1, "name": "9 А"},
					{"id": 2, "name": "10 Б"},
				},
			},
			{
				"id":   2,
				"name": "Гимназия № 3 им. М. Горького",
				"city": "г. Уфа",
				"classes": []map[string]any{
					{"id": 3, "name": "8 В"},
					{"id": 4, "name": "11 А"},
				},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

type createClassReq struct {
	SchoolID    uuid.UUID `json:"school_id"`
	GradeNumber int       `json:"grade_number"`
	Letter      string    `json:"letter"`
}

func (h *LMSHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	var req createClassReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.SchoolID == uuid.Nil {
		http.Error(w, `{"error":"school_id is required"}`, http.StatusBadRequest)
		return
	}

	class, err := h.lmsUC.CreateClass(r.Context(), req.SchoolID, req.GradeNumber, req.Letter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(class)
}

func (h *LMSHandler) GenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "id")
	if classIDStr == "" {
		classIDStr = chi.URLParam(r, "class_id")
	}
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid class id"}`, http.StatusBadRequest)
		return
	}

	code, err := h.lmsUC.GenerateInviteCode(r.Context(), classID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"class_id":    classID,
		"invite_code": code,
		"code":        code,
	})
}

type joinClassReq struct {
	InviteCode string `json:"invite_code"`
	Code       string `json:"code"`
}

func (h *LMSHandler) JoinClass(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req joinClassReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	code := req.InviteCode
	if code == "" {
		code = req.Code
	}

	class, err := h.lmsUC.JoinClassByCode(r.Context(), userID, code)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":   true,
		"class_id":  class.ID,
		"school_id": class.SchoolID,
		"message":   "Вы успешно присоединились к классу!",
		"class":     class,
	})
}

type addStudentReq struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Password string    `json:"password"`
}

func (h *LMSHandler) AddStudent(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "id")
	if classIDStr == "" {
		classIDStr = chi.URLParam(r, "class_id")
	}
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid class id"}`, http.StatusBadRequest)
		return
	}

	var req addStudentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == uuid.Nil && req.Email == "" {
		http.Error(w, `{"error":"user_id or email is required"}`, http.StatusBadRequest)
		return
	}

	var class *domain.Class
	var studentUser *domain.User

	if req.UserID != uuid.Nil {
		class, err = h.lmsUC.AssignStudentToClass(r.Context(), req.UserID, classID)
	} else {
		studentUser, class, err = h.lmsUC.RegisterStudentByTeacher(r.Context(), classID, req.FullName, req.Email, req.Password)
	}

	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	resp := map[string]any{
		"success":   true,
		"class_id":  class.ID,
		"school_id": class.SchoolID,
		"message":   "Ученик успешно добавлен в класс",
	}
	if studentUser != nil {
		resp["student_id"] = studentUser.ID
		resp["email"] = studentUser.Email
		resp["full_name"] = studentUser.FullName
		resp["user"] = studentUser
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

