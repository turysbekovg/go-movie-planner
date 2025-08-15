package errs

import "errors"

// Объявляем наши кастомные типы ошибок как переменные

// ErrNotFound будет возвращаться, когда фильм не найден
var ErrNotFound = errors.New("the requested resource was not found")

// ErrProviderFailure будет возвращаться, когда внешний сервис (провайдер)
// не отвечает или возвращает ошибку, не связанную с "не найдено"
var ErrProviderFailure = errors.New("the external provider failed to respond")
