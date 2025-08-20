package service

import (
	"github.com/turysbekovg/movie-planner/internal/ports" // ядро зависит от портов
)

// MovieService -> ядро, структура сервиса
// Она содержит ссылку на наш порт
// сервис зависит от интерфейса, а не от конкретной реализации
type MovieService struct {
	provider ports.MovieProvider
}

// NewMovieService -> конструктор для ядра/сервиса
// Он принимает на вход provider и сохраняет его внутри создаваемого сервиса
// Получает того кто реализовал MovieProvider
func NewMovieService(provider ports.MovieProvider) *MovieService {
	return &MovieService{provider: provider} // Внутрь кладем адаптер (cacheAdapter)
}

// FinalMovieData -> финальный ответ который возвращается пользователю
// включает все поля из ports.Movie и добавляет новое поле Advice
type FinalMovieData struct {
	ports.Movie
	Advice string `json:"advice" example:"It is a very good choice! A high rated movie, which is recommended to watch."`
}

// GetMovie -> главный метод сервиса (driving port/входящий порт)
func (s *MovieService) GetMovie(title string) (*FinalMovieData, error) {
	// 1. Вызываем метод SearchMovie у своего s.provider, чтобы получить базовые данные о фильме
	movie, err := s.provider.SearchMovie(title)
	if err != nil {
		// Если провайдер вернул ошибку, мы просто передаем ее дальше
		return nil, err
	}

	// 2. Наша бизнес-логика: генерируем совет на основе рейтинга
	var advice string
	if movie.Rating >= 7.5 {
		advice = "It is a very good choice! A high rated movie, which is recommended to watch."
	} else if movie.Rating >= 5.0 {
		advice = "A good option for a night, but do not expect something perfect."
	} else {
		advice = "A controversial choice. Not really recommended to watch, but you still can do so."
	}

	// 3. Собираем финальную структуру для ответа
	finalData := &FinalMovieData{
		Movie:  *movie, // Копируем все поля из полученной структуры movie
		Advice: advice,
	}

	return finalData, nil
}
