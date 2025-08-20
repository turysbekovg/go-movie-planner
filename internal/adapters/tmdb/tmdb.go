// Реализация интерфейса MovieProvider. Идет на сайт IMBD и получает данные
package tmdb

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/turysbekovg/movie-planner/internal/errs"
	"github.com/turysbekovg/movie-planner/internal/ports"
)

const (
	// Базовый URL для API и для изображений (постеров)
	apiBaseURL   = "https://api.themoviedb.org/3"
	imageBaseURL = "https://image.tmdb.org/t/p/w500"
)

// Эти две структуры нужны, чтобы разобрать (распарсить) JSON-ответ от TMDb
// Их структура в точности повторяет структуру ответа от API
type tmdbSearchResponse struct {
	Results []tmdbMovie `json:"results"`
}

type tmdbMovie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	PosterPath  string  `json:"poster_path"`
}

// TMDbAdapter -> структура, которая будет реализовывать интерфейс MovieProvider
// это адаптер для технологии "TMDb REST API"
type TMDbAdapter struct {
	apiKey string       // для авторизации
	client *http.Client // для выполнения HTTP calls
}

// NewTMDbAdapter -> это конструктор для нашего адаптера
func NewTMDbAdapter(apiKey string) *TMDbAdapter {
	return &TMDbAdapter{
		apiKey: apiKey,
		client: &http.Client{}, // Создаем стандартный HTTP клиент
	}
}

// ---------------------------------------------------------------------------------------------

// searchForMovie ищет фильм по названию и возвращает полную структуру tmdbMovie
func (a *TMDbAdapter) searchByName(title string) (*tmdbMovie, error) {
	log.Printf("Searching for movie '%s' on TMDb...", title)

	// 1. Собираем правильный URL для запроса
	// url.QueryEscape нужен, чтобы обработать пробелы и спецсимволы в названии, например, The Matrix
	searchURL := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s",
		apiBaseURL, a.apiKey, url.QueryEscape(title))

	// 2. Делаем/отправляем GET-запрос по этому URL
	resp, err := a.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrProviderFailure, err)
	}
	// Закрываем тело ответа, чтобы избежать утечек ресурсов
	defer resp.Body.Close()

	// 3. Проверяем, если вышла ошибка 404 -> возвращаем эту ошибку (ErrorNotFound)
	if resp.StatusCode == http.StatusNotFound {
		return nil, errs.ErrNotFound
	}
	// Также проверяем если запрос не прошел успешно (200 OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: TMDb API returned non-200 status: %s", errs.ErrProviderFailure, resp.Status)
	}

	// 4. Распоковываем (декодируем) JSON ответ в нашу структуру tmdbSearchResponse
	// читает тело ответа и раскладывает JSON данные
	// по полям структур-шаблонов tmdbSearchResponse и tmdbMovie
	var searchResponse tmdbSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("%w: failed to decode TMDb response: %v", errs.ErrProviderFailure, err)
	}

	// 5. Проверяем, что хоть один фильм нашелся
	// Если массив пуст -> не найдено и поиск дальше не нужен (ErrorNotFound)
	if len(searchResponse.Results) == 0 {
		return nil, errs.ErrNotFound
	}

	// 6. Берем самый первый (самый релевантный) фильм из списка результатов
	// Возвращаем указатель на первый найденный фильм
	return &searchResponse.Results[0], nil
}

// fetchRecommendations запрашивает рекомендации по ID фильма
// Возвращает срез названий фильмов, otherwise -> просто вернет пустой срез и запишет ошибку в лог
func (a *TMDbAdapter) fetchRecommendations(movieID int) []string {
	log.Printf("Getting recommendations for movie ID %d...", movieID)

	// 1. Собираем URL, использует movieID
	recsURL := fmt.Sprintf("%s/movie/%d/recommendations?api_key=%s",
		apiBaseURL, movieID, a.apiKey)

	// 2. Отправляем второй GET запрос
	recsResp, err := a.client.Get(recsURL)
	if err != nil {
		log.Printf("Warning: failed to call TMDb recommendations API: %v", err)
		return nil // Возвращаем nil (который превратится в пустой JSON-массив)
	}
	// Закрываем тело
	defer recsResp.Body.Close()

	// 3. Если не вернулся правильный статус то выводим статус и предупреждение
	if recsResp.StatusCode != http.StatusOK {
		log.Printf("Warning: failed to get recommendations, status: %s", recsResp.Status)
		return nil
	}

	// 4. Декодируем JSON ответ
	var recsResponse tmdbSearchResponse
	if err := json.NewDecoder(recsResp.Body).Decode(&recsResponse); err != nil {
		// Если не смогли жекодировать
		log.Printf("Warning: failed to decode recommendations response: %v", err)
		return nil
	}

	// 5. Собираем названия фильмов из рекомендаций в срез
	// Использовал make для предварительного выделения памяти под срез ->
	// так как мы заранее знаем, сколько элементов ожидаем
	recommendations := make([]string, 0, len(recsResponse.Results))
	for _, recMovie := range recsResponse.Results {

		// Добавляем в срез только заголовок каждого рекомендованного фильма
		recommendations = append(recommendations, recMovie.Title)
	}

	// Возвращаем готовый срез с названиями
	return recommendations
}

// SearchMovie -> реализация метода из интерфейса ports.MovieProvider
func (a *TMDbAdapter) SearchMovie(title string) (*ports.Movie, error) {
	// 1. Ищем фильм по названию через searchByName
	foundMovie, err := a.searchByName(title)
	if err != nil {
		// Если фильм не найден или произошла критическая ошибка, сразу возвращаем ошибку
		return nil, err
	}

	// 2. Ищем рекоммендации через fetchRecommendations
	// Эта функция не возвращает ошибок, поэтому не нужно их проверять
	recommendations := a.fetchRecommendations(foundMovie.ID)

	// 3. Собираем финальный результат
	movie := &ports.Movie{
		Title:           foundMovie.Title,
		Overview:        foundMovie.Overview,
		ReleaseDate:     foundMovie.ReleaseDate,
		Rating:          foundMovie.VoteAverage,
		PosterURL:       imageBaseURL + foundMovie.PosterPath,
		Recommendations: recommendations,
	}

	return movie, nil
}
