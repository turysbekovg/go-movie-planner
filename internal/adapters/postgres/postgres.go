package postgres

import (
	"context"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turysbekovg/movie-planner/internal/errs"
	"github.com/turysbekovg/movie-planner/internal/ports"

	"github.com/jackc/pgx/v5"
)

// Стракт, которая  реализовывает интерфейс MovieRepository
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{pool: pool}
}

func (a *PostgresAdapter) CreateMovie(ctx context.Context, movie *ports.Movie) (int, error) {
	var id int

	// $1, $2, -> это плейсхолдеры для сейф вставки переменных в запрос (защита от SQL-инъекций)
	query := `INSERT INTO movies (title, overview, release_date, rating, poster_url, recommendations) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := a.pool.QueryRow(ctx, query,
		movie.Title,
		movie.Overview,
		movie.ReleaseDate.Time,
		movie.Rating,
		movie.PosterURL,
		strings.Join(movie.Recommendations, ","),
	).Scan(&id) // Для чтения и записи id

	if err != nil {
		log.Printf("Error creating movie: %v", err)
		return 0, err
	}

	return id, nil
}

func (a *PostgresAdapter) GetMovieByID(ctx context.Context, id int) (*ports.Movie, error) {
	var m ports.Movie
	var recommendations string

	query := `SELECT id, title, overview, release_date, rating, poster_url, recommendations 
              FROM movies WHERE id = $1`

	err := a.pool.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.Title,
		&m.Overview,
		&m.ReleaseDate,
		&m.Rating,
		&m.PosterURL,
		&recommendations, // Сначала читаем в строку
	)

	if err != nil {
		// pgx.ErrNoRows -> ошибка, которую pgx возвращает, если SELECT не нашел ни одной строки
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		log.Printf("Error getting movie by ID: %v", err)
		return nil, err
	}

	m.Recommendations = strings.Split(recommendations, ",")

	return &m, nil
}

func (a *PostgresAdapter) UpdateMovie(ctx context.Context, id int, movie *ports.Movie) error {
	query := `UPDATE movies SET 
                  title = $1, 
                  overview = $2, 
                  release_date = $3, 
                  rating = $4, 
                  poster_url = $5, 
                  recommendations = $6,
                  updated_at = CURRENT_TIMESTAMP
              WHERE id = $7`

	// a.pool.Exec -> выполняет запрос, который не возвращает строк (как UPDATE и тп)
	_, err := a.pool.Exec(ctx, query,
		movie.Title,
		movie.Overview,
		movie.ReleaseDate.Time,
		movie.Rating,
		movie.PosterURL,
		strings.Join(movie.Recommendations, ","),
		id,
	)

	if err != nil {
		log.Printf("Error updating movie: %v", err)
		return err
	}

	return nil
}

func (a *PostgresAdapter) DeleteMovie(ctx context.Context, id int) error {
	query := `DELETE FROM movies WHERE id = $1`

	_, err := a.pool.Exec(ctx, query, id)

	if err != nil {
		log.Printf("Error deleting movie: %v", err)
		return err
	}

	return nil
}

func (a *PostgresAdapter) GetAllMovies(ctx context.Context) ([]*ports.Movie, error) {

	movies := make([]*ports.Movie, 0)

	query := `SELECT id, title, overview, release_date, rating, poster_url, recommendations FROM movies`

	rows, err := a.pool.Query(ctx, query)
	if err != nil {
		log.Printf("Error querying all movies: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m ports.Movie
		var recommendations string

		err := rows.Scan(
			&m.ID,
			&m.Title,
			&m.Overview,
			&m.ReleaseDate,
			&m.Rating,
			&m.PosterURL,
			&recommendations,
		)
		if err != nil {
			log.Printf("Error scanning movie row: %v", err)
			return nil, err
		}

		m.Recommendations = strings.Split(recommendations, ",")

		movies = append(movies, &m)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating movie rows: %v", err)
		return nil, err
	}

	return movies, nil
}

func (a *PostgresAdapter) CreateUser(ctx context.Context, user *ports.User) (int, error) {
	var id int
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	err := a.pool.QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return 0, err
	}

	return id, nil
}

func (a *PostgresAdapter) GetUserByEmail(ctx context.Context, email string) (*ports.User, error) {
	var u ports.User
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`

	err := a.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		log.Printf("Error getting user by email: %v", err)
		return nil, err
	}

	return &u, nil
}
