package service

import (
	"context"

	"github.com/turysbekovg/movie-planner/internal/ports" // ядро зависит от портов
)

// MovieService -> ядро
type MovieService struct {
	repo ports.MovieRepository
}

func NewMovieService(repo ports.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

// FinalMovieData -> финальный ответ который возвращается пользователю
type FinalMovieData struct {
	ports.Movie
	Advice string `json:"advice" example:"It is a very good choice! A high rated movie, which is recommended to watch."`
}

func (s *MovieService) GetMovieByID(ctx context.Context, id int) (*FinalMovieData, error) {
	movie, err := s.repo.GetMovieByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var advice string
	if movie.Rating >= 7.5 {
		advice = "It is a very good choice! A high rated movie, which is recommended to watch."
	} else if movie.Rating >= 5.0 {
		advice = "A good option for a night, but do not expect something perfect."
	} else {
		advice = "A controversial choice. Not really recommended to watch, but you still can do so."
	}

	// Собираем финальную структуру для ответа
	finalData := &FinalMovieData{
		Movie:  *movie,
		Advice: advice,
	}

	return finalData, nil
}

func (s *MovieService) CreateMovie(ctx context.Context, movie *ports.Movie) (int, error) {
	return s.repo.CreateMovie(ctx, movie)
}

func (s *MovieService) GetAllMovies(ctx context.Context) ([]*ports.Movie, error) {
	return s.repo.GetAllMovies(ctx)
}

func (s *MovieService) UpdateMovie(ctx context.Context, id int, movie *ports.Movie) error {
	return s.repo.UpdateMovie(ctx, id, movie)
}

func (s *MovieService) DeleteMovie(ctx context.Context, id int) error {
	return s.repo.DeleteMovie(ctx, id)
}
