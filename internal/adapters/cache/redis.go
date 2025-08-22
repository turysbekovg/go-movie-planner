package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turysbekovg/movie-planner/internal/ports"
)

type RedisCacheAdapter struct {
	next   ports.MovieRepository // след репозиторий
	client *redis.Client
	ttl    time.Duration
}

// Конструктор для нашего кэширующего адаптера
func NewRedisCacheAdapter(next ports.MovieRepository, client *redis.Client, ttl time.Duration) *RedisCacheAdapter {
	return &RedisCacheAdapter{
		next:   next,
		client: client,
		ttl:    ttl,
	}
}

func (a *RedisCacheAdapter) GetMovieByID(ctx context.Context, id int) (*ports.Movie, error) {
	// Формируем ключ для Redis
	key := fmt.Sprintf("movie:%d", id)

	// Пытаемся получить данные по ключу
	cachedData, err := a.client.Get(ctx, key).Result()
	if err == nil {
		log.Printf("Cache HIT for movie ID: %d", id)
		var movie ports.Movie
		if err := json.Unmarshal([]byte(cachedData), &movie); err == nil {
			return &movie, nil
		}
	}

	// Если err != nil -> Cache MISS, идем к следующему репозиторию (в БД)
	log.Printf("Cache MISS for movie ID: %d. Fetching from next repository.", id)
	movie, err := a.next.GetMovieByID(ctx, id)
	if err != nil {
		// Если в базе фильма нет, то и в кэш ничего не кладем
		return nil, err
	}

	// Сериализуем полученную структуру в JSON для сохранения в кэше
	jsonData, err := json.Marshal(movie)
	if err != nil {
		log.Printf("Warning: failed to marshal movie for cache: %v", err)
		return movie, nil // Возвращаем фильм, но не кэшируем в случае ошибки
	}

	// Сохраняем JSON в Redis
	err = a.client.Set(ctx, key, jsonData, a.ttl).Err()
	if err != nil {
		log.Printf("Warning: failed to set cache for movie ID %d: %v", id, err)
	}

	return movie, nil
}

func (a *RedisCacheAdapter) UpdateMovie(ctx context.Context, id int, movie *ports.Movie) error {
	err := a.next.UpdateMovie(ctx, id, movie)
	if err != nil {
		return err
	}

	// Если обновление в базе прошло успешно -> инвалидируем кэш
	key := fmt.Sprintf("movie:%d", id)
	if err := a.client.Del(ctx, key).Err(); err != nil {
		log.Printf("Warning: failed to invalidate cache for movie ID %d: %v", id, err)
	} else {
		log.Printf("Cache invalidated for movie ID: %d", id)
	}
	return nil
}

func (a *RedisCacheAdapter) DeleteMovie(ctx context.Context, id int) error {
	err := a.next.DeleteMovie(ctx, id)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("movie:%d", id)

	// Если усмешно -> инвалидируем кэш
	if err := a.client.Del(ctx, key).Err(); err != nil {
		log.Printf("Warning: failed to invalidate cache for movie ID %d: %v", id, err)
	} else {
		log.Printf("Cache invalidated for movie ID: %d", id)
	}
	return nil
}

// Для этих методов мы кидаем вызов дальше, не добавляя логику кэширования
func (a *RedisCacheAdapter) CreateMovie(ctx context.Context, movie *ports.Movie) (int, error) {
	return a.next.CreateMovie(ctx, movie)
}

func (a *RedisCacheAdapter) GetAllMovies(ctx context.Context) ([]*ports.Movie, error) {
	return a.next.GetAllMovies(ctx)
}
