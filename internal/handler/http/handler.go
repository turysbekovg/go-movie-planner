package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/turysbekovg/movie-planner/internal/errs"
	"github.com/turysbekovg/movie-planner/internal/ports"
	"github.com/turysbekovg/movie-planner/internal/service"
)

// Нужна для генерации правильной документации в Swagger
type SwaggerMovieRequest struct {
	Title           string   `json:"title" example:"Inception"`
	Overview        string   `json:"overview" example:"A thief who steals corporate secrets..."`
	ReleaseDate     string   `json:"release_date" example:"2010-07-16"`
	Rating          float64  `json:"rating" example:"8.8"`
	PosterURL       string   `json:"poster_url" example:"https://image.tmdb.org/..."`
	Recommendations []string `json:"recommendations" example:"The Matrix,Shutter Island"`
}

type MovieHandler struct {
	service *service.MovieService
}

// NewMovieHandler -> конструктор, который создает MovieHanlder получая на вход экземпляр сервиса
func NewMovieHandler(s *service.MovieService) *MovieHandler {
	return &MovieHandler{
		service: s, // Сервис
	}
}

// CreateMovie godoc
// @Summary      Create a new movie
// @Description  Adds a new movie to the database. Requires authentication.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param movie body http.SwaggerMovieRequest true "Movie data to create"
// @Success      201 {object} map[string]int
// @Failure      400 {string} string "Invalid request body"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Failed to create movie"
// @Security     BearerAuth
// @Router       /movies [post]
func (h *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie ports.Movie
	// Читаем JSON из тела запроса и декодируем его в нашу структуру
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Вызываем метод сервиса для создания фильма
	id, err := h.service.CreateMovie(r.Context(), &movie)
	if err != nil {
		http.Error(w, "Failed to create movie", http.StatusInternalServerError)
		return
	}

	// Отвечаем клиенту, что успешно создан (201)
	// и возвращаем ID созданного фильма
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

// GetMovieByID godoc
// @Summary      Get a movie by ID
// @Description  Retrieves movie details for a given ID. This endpoint is public.
// @Tags         movies
// @Produce      json
// @Param        id path int true "Movie ID"
// @Success      200 {object} service.FinalMovieData
// @Failure      400 {string} string "Invalid movie ID"
// @Failure      404 {string} string "the requested resource was not found"
// @Router       /movies/{id} [get]
func (h *MovieHandler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	movieData, err := h.service.GetMovieByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("Internal error: %v", err)
			http.Error(w, "An internal server error occurred", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movieData)
}

// UpdateMovie godoc
// @Summary      Update a movie
// @Description  Updates an existing movie's details. Requires authentication.
// @Tags         movies
// @Accept       json
// @Param        id path int true "Movie ID"
// @Param movie body http.SwaggerMovieRequest true "Movie data to update"
// @Success      204 "No Content"
// @Failure      400 {string} string "Invalid movie ID or request body"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Failed to update movie"
// @Security     BearerAuth
// @Router       /movies/{id} [put]
func (h *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var movie ports.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.UpdateMovie(r.Context(), id, &movie)
	if err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}

	// Отвечаем, что все прошло успешно
	w.WriteHeader(http.StatusNoContent)
}

// DeleteMovie godoc
// @Summary      Delete a movie
// @Description  Deletes a movie from the database. Requires authentication.
// @Tags         movies
// @Param        id path int true "Movie ID"
// @Success      204 "No Content"
// @Failure      400 {string} string "Invalid movie ID"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Failed to delete movie"
// @Security     BearerAuth
// @Router       /movies/{id} [delete]
func (h *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteMovie(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllMovies godoc
// @Summary      Get all movies
// @Description  Retrieves a list of all movies. This endpoint is public.
// @Tags         movies
// @Produce      json
// @Success      200 {array} ports.Movie
// @Failure      500 {string} string "Failed to get movies"
// @Router       /movies [get]
func (h *MovieHandler) GetAllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.service.GetAllMovies(r.Context())
	if err != nil {
		http.Error(w, "Failed to get movies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}
