package ports

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// @swaggertype string
// @format date
// @example "2023-10-27"
// Чтобы записать дату правильно
type CustomDate struct {
	time.Time
}

// UnmarshalJSON -> метод для декодирования JSON
func (cd *CustomDate) UnmarshalJSON(b []byte) (err error) {

	log.Println("--- Custom UnmarshalJSON CALLED! ---")
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	cd.Time, err = time.Parse("2006-01-02", s)
	return err
}

// Чтобы pgx смог положить в кастомную дату
func (cd *CustomDate) Scan(src interface{}) error {
	// Чекнем, что из базы пришли данные типа time.Time
	if t, ok := src.(time.Time); ok {
		*cd = CustomDate{Time: t}
		return nil
	}

	return fmt.Errorf("cannot scan %T into CustomDate", src)
}

// Для того чтобы выдать удобный releaseDate в GET (2006-01-02 к примеру)
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	formatted := cd.Time.Format("2006-01-02")
	return []byte(`"` + formatted + `"`), nil
}

type Movie struct {
	ID              int        `json:"id" example:"1"`
	Title           string     `json:"title" example:"Inception"`
	Overview        string     `json:"overview" example:"A thief who steals corporate secrets..."`
	ReleaseDate     CustomDate `json:"release_date"`
	Rating          float64    `json:"rating" example:"8.8"`
	PosterURL       string     `json:"poster_url" example:"https://image.tmdb.org/..."`
	Recommendations []string   `json:"recommendations" example:"The Matrix,Shutter Island"`
}

// Мы не добавляем json тег для password_hash, чтобы случайно не отдать его клиенту
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type MovieRepository interface {
	CreateMovie(ctx context.Context, movie *Movie) (int, error)
	GetMovieByID(ctx context.Context, id int) (*Movie, error)
	GetAllMovies(ctx context.Context) ([]*Movie, error)
	UpdateMovie(ctx context.Context, id int, movie *Movie) error
	DeleteMovie(ctx context.Context, id int) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}
