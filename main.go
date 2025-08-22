package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/turysbekovg/movie-planner/internal/adapters/cache"
	"github.com/turysbekovg/movie-planner/internal/adapters/postgres" // Наш новый адаптер
	handler "github.com/turysbekovg/movie-planner/internal/handler/http"
	"github.com/turysbekovg/movie-planner/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/turysbekovg/movie-planner/docs" // Пустой импорт для swag
)

func connectToDB() *pgxpool.Pool {
	// Собираем строку подключения из переменных окружения.
	// Если их нет, используем значения по умолчанию из нашего docker-compose.yml
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "myuser"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "mypassword"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5433"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "movie_planner"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Проверяем, что соединение действительно установлено
	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	log.Println("Successfully connected to the database!")
	return dbpool
}

func connectToRedis() *redis.Client {
	// В нашем docker-compose.yml Redis доступен по адресу "redis:6379",
	// но мы пробросили порт 6379 на наш localhost, поэтому можем использовать его.
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // нет пароля
		DB:       0,  // использовать БД по умолчанию
	})

	// Проверяем, что соединение установлено.
	// Ping должен вернуть "PONG" в случае успеха.
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis!")
	return rdb
}

// @title           Movie Night Planner API
// @version         1.0
// @description     This is a sample server for a movie planner application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a JWT token.

// @host      localhost:8080
// @BasePath  /
func main() {
	// 1. Конфигурация
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	dbPool := connectToDB()
	defer dbPool.Close()

	redisClient := connectToRedis()
	defer redisClient.Close()

	// 2. Сборка зависимостей (Dependency Injection)

	// Адаптер для БД
	dbAdapter := postgres.NewPostgresAdapter(dbPool)

	// Кэш адаптер
	cacheAdapter := cache.NewRedisCacheAdapter(dbAdapter, redisClient, 5*time.Minute)

	// Сервис для фильмов
	movieSvc := service.NewMovieService(cacheAdapter)

	// Обработчик для фильмов
	movieHandler := handler.NewMovieHandler(movieSvc) // <<< ИЗМЕНЕНИЕ 2: Используем новый псевдоним

	// Сервис для пользователей
	userSvc := service.NewUserService(dbAdapter)

	// Сервис для JWT
	jwtSecretKey := "my_super_secret_key"
	jwtTTL := 24 * time.Hour
	authSvc := service.NewAuthSvc(jwtSecretKey, jwtTTL)

	// Обработчик для аутентификации
	authHandler := handler.NewAuthHandler(userSvc, authSvc) // <<< ИЗМЕНЕНИЕ 3: Используем новый псевдоним

	// 3. Настройка роутера и запуск сервера
	r := chi.NewRouter()
	r.Use(middleware.Logger) // Используем логгер для всех запросов

	// Добавляем маршрут для Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // The url pointing to API definition
	))

	// Роуты для аутентификации (публичные)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register) // POST /auth/register
		r.Post("/login", authHandler.Login)       // POST /auth/login
	})

	// Группа ПУБЛИЧНЫХ роутов для фильмов (только чтение)
	r.Route("/movies", func(r chi.Router) {
		r.Get("/", movieHandler.GetAllMovies)     // GET /movies
		r.Get("/{id}", movieHandler.GetMovieByID) // GET /movies/123
	})

	// Группа ЗАЩИЩЕННЫХ роутов для фильмов (создание, изменение, удаление)
	r.Group(func(r chi.Router) {
		// Применяем наше AuthMiddleware ко всем роутам внутри этой группы.
		// Мы передаем в него authSvc, чтобы middleware мог проверять токены.
		r.Use(handler.AuthMiddleware(authSvc))

		// Роуты, которые теперь требуют валидный JWT.
		r.Post("/movies", movieHandler.CreateMovie)        // POST /movies
		r.Put("/movies/{id}", movieHandler.UpdateMovie)    // PUT /movies/123
		r.Delete("/movies/{id}", movieHandler.DeleteMovie) // DELETE /movies/123
	})

	log.Println("Starting server on http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Теперь получается вот так:
// HTTP Запрос -> Handler -> Service -> RedisCacheAdapter -> PostgresAdapter -> База Данных
