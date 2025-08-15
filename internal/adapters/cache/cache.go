package cache

import (
	"log"
	"sync"
	"time"

	"github.com/turysbekovg/movie-planner/internal/ports"
)

// Структура для хранения одной записи в кэше (cacheEntry)
type cacheEntry struct {
	movie     *ports.Movie // Данные о фильме, которые мы кэшируем
	createdAt time.Time    // Время, когда запись была добавлена в кэш
}

// CacheAdapter -> адаптер-декоратор
type CacheAdapter struct {
	// next - это настоящий провайдер (например, TMDbAdapter),
	// к которому мы будем обращаться, если данных нет в кэше
	next ports.MovieProvider

	// cache - хранилище. Ключ - название фильма (string),
	// значение - запись в кэше (cacheEntry)
	cache map[string]cacheEntry

	// ttl (Time To Live) - время жизни записи в кэше
	ttl time.Duration

	// mu - мьютекс для безопасного доступа к кэшу из нескольких горутин
	mu sync.Mutex
}

// NewCacheAdapter - конструктор для кэширующего адаптера
func NewCacheAdapter(next ports.MovieProvider, ttl time.Duration) *CacheAdapter {
	return &CacheAdapter{
		next:  next,
		ttl:   ttl,
		cache: make(map[string]cacheEntry),
	}
}

// --------------------------------------------------------------------------------------

// SearchMovie реализует интерфейс MovieProvider, добавляя логику кэширования
func (a *CacheAdapter) SearchMovie(title string) (*ports.Movie, error) {
	// 1. Проверка кэша

	// Закрываем мьютекс, чтобы безопасно работать с map
	// defer a.mu.Unlock() гарантирует, что мьютекс будет освобожден при выходе из функции
	a.mu.Lock()
	entry, found := a.cache[title]
	a.mu.Unlock() // Освобождаем мьютекс сразу после чтения

	// Проверяем, есть ли запись в кэше и не устарела ли она
	if found && time.Since(entry.createdAt) < a.ttl {
		log.Printf("Cache HIT for movie: %s", title)
		// 2. Если в кэше есть запись -> возвращаем
		return entry.movie, nil
	}

	// 3. Если в кэше нет записи -> идем к следующему провайдеру
	log.Printf("Cache MISS for movie: %s. Fetching from next provider...", title)
	movie, err := a.next.SearchMovie(title)
	if err != nil {
		// Если настоящий провайдер вернул ошибку -> ее не кэшируем, а возвращаем
		return nil, err
	}

	// 4. Сохранение в кэш

	// Снова закрываем мьютекс, но теперь для записи в map
	a.mu.Lock()
	a.cache[title] = cacheEntry{
		movie:     movie,
		createdAt: time.Now(),
	}
	a.mu.Unlock()

	log.Printf("Stored movie '%s' in cache", title)

	return movie, nil
}
