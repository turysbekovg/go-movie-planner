package main

import (
	"log"
	"net/http"
	"os"

	"github.com/turysbekovg/movie-planner/internal/adapters/tmdb"
	movieHandler "github.com/turysbekovg/movie-planner/internal/handler/http"
	"github.com/turysbekovg/movie-planner/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/turysbekovg/movie-planner/docs" // Пустой импорт для swag

	"time"

	"github.com/turysbekovg/movie-planner/internal/adapters/cache"
)

// @title           Movie Night Planner API
// @version         1.0
// @description     This is a sample server for a movie planner application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
func main() {
	// 1. Конфигурация
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}
	// Читаем API-ключ из переменных окружения.
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		log.Fatal("TMDB_API_KEY environment variable is not set")
	}

	// 2. Сборка зависимостей (Dependency Injection)
	// Создаем адаптер (самый внутренний слой, работает с внешним миром)
	tmdbAdapter := tmdb.NewTMDbAdapter(apiKey)

	// Создаем кэширующий адаптер-декоратор
	// Он оборачивает tmdbAdapter и будет кэшировать его результаты на 5 минут
	cacheAdapter := cache.NewCacheAdapter(tmdbAdapter, 5*time.Minute)

	// Создаем сервис, передавая ему кэш-адаптер
	movieSvc := service.NewMovieService(cacheAdapter)

	// Создаем обработчик, передавая ему сервис
	handler := movieHandler.NewMovieHandler(movieSvc)

	// 3. Настройка роутера и запуск сервера
	r := chi.NewRouter()
	r.Use(middleware.Logger) // Используем логгер для всех запросов

	// Добавляем маршрут для Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // The url pointing to API definition
	))

	// Привязываем метод нашего обработчика к маршруту
	r.Get("/movie/{title}", handler.GetMovie)

	log.Println("Starting server on http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
