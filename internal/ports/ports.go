package ports

// Создаем объект/модель Movie
// Поля -> title, overview, releasedate, rating, posterURL
type Movie struct {
	Title           string   `json:"title" example:"Inception"`
	Overview        string   `json:"overview" example:"A thief who steals corporate secrets..."`
	ReleaseDate     string   `json:"release_date" example:"2010-07-15"`
	Rating          float64  `json:"rating" example:"8.8"`
	PosterURL       string   `json:"poster_url" example:"https://image.tmdb.org/..."`
	Recommendations []string `json:"recommendations" example:"The Matrix,Shutter Island"`
}

// Создаем интерфейс MovieProvider
// Обязан иметь SearchMovie (возвращает ->  *Movie и ошибку error)
type MovieProvider interface {
	SearchMovie(title string) (*Movie, error)
}
