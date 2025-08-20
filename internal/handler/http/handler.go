package http

import (
	"encoding/json"
	"net/http"

	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/turysbekovg/movie-planner/internal/errs"
	"github.com/turysbekovg/movie-planner/internal/service" // адаптер зависит от ядра
)

// MovieHandler -> это структура обработчика. Является адаптером
// Она содержит ссылку на сервис, чтобы иметь возможность вызывать его методы
type MovieHandler struct {
	service *service.MovieService
}

// NewMovieHandler -> конструктор, который создает MovieHanlder получая на вход экземпляр сервиса
func NewMovieHandler(s *service.MovieService) *MovieHandler {
	return &MovieHandler{
		service: s, // Сервис
	}
}

// GetMovie -> метод, который будет привязан к URL. Это логика Адаптера

// GetMovie godoc
// @Summary      Get movie details by title
// @Description  get details of a movie by its title from TMDb
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        title path string true "Movie Title"
// @Success      200 {object} service.FinalMovieData
// @Failure      404 {object} string "Not Found"
// @Failure      503 {object} string "Service Unavailable"
// @Failure      500 {object} string "Internal Server Error"
// @Router       /movie/{title} [get]
func (h *MovieHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	// 1. Извлекаем название фильма из URL
	title := chi.URLParam(r, "title")

	// 2. Вызываем метод сервиса для получения данных о фильме
	movieData, err := h.service.GetMovie(title)
	if err != nil {
		// Используем errors.Is() для проверки типа ошибки
		if errors.Is(err, errs.ErrNotFound) {
			// Если ошибка ErrorNotFound, возвращаем статус 404
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if errors.Is(err, errs.ErrProviderFailure) {
			// Если ошибка связана с внешним провайдером, возвращаем 503
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		} else {
			// Для всех остальных неожиданных ошибок возвращаем 500
			http.Error(w, "An internal server error occurred", http.StatusInternalServerError)
		}
		return
	}

	// 3. Если все прошло успешно, устанавливаем заголовок Content-Type,
	// чтобы браузер понимал, что отправляю JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 4. Превращаем данные (структуру `FinalMovieData`) в JSON
	// и отправляем их в качестве тела ответа
	err = json.NewEncoder(w).Encode(movieData)
	if err != nil {
		// Если возникла ошибка при кодировании в JSON, сообщаем об этом
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
