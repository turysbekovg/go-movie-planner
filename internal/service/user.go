package service

import (
	"context"
	"fmt"

	"github.com/turysbekovg/movie-planner/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(ctx context.Context, email, password string) (int, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &ports.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	// Сохраняем пользователя в базу через репозиторий
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) LoginUser(ctx context.Context, email, password string) (*ports.User, error) {
	// Ищем юзера
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Сравниваем предоставленный пароль с хэшем в базе
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Если пароли не совпали, bcrypt вернет ошибку.
		return nil, fmt.Errorf("invalid credentials")
	}

	// Если все в порядке, возвращаем пользователя
	return user, nil
}
