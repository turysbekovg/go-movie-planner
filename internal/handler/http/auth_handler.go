package http

import (
	"encoding/json"
	"net/http"

	"github.com/turysbekovg/movie-planner/internal/service"
)

type AuthHandler struct {
	userSvc *service.UserService
	authSvc *service.AuthSvc
}

func NewAuthHandler(userSvc *service.UserService, authSvc *service.AuthSvc) *AuthHandler {
	return &AuthHandler{
		userSvc: userSvc,
		authSvc: authSvc,
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user body authRequest true "User Registration Info"
// @Success      201 {object} map[string]interface{}
// @Failure      400 {string} string "Invalid request body"
// @Failure      500 {string} string "Failed to register user"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.userSvc.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"id":      id,
	})
}

// Login godoc
// @Summary      Login a user
// @Description  Logs in a user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body authRequest true "User Credentials"
// @Success      200 {object} map[string]string
// @Failure      400 {string} string "Invalid request body"
// @Failure      401 {string} string "Invalid credentials"
// @Failure      500 {string} string "Failed to generate token"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userSvc.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.authSvc.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
