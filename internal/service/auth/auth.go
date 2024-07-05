package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/logger/sl"
	"sso/storage"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	toketTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

// New создает и возвращает новый экземпляр Auth.
//
// Параметры:
// - log: Логгер для записи журналов.
// - usersaver: Интерфейс для сохранения пользователей.
// - userProvider: Интерфейс для предоставления пользователей.
// - appProvider: Интерфейс для предоставления приложений.
// - tokenTTL: Продолжительность времени действия токена.
//
// Возвращает:
// - *Auth: Новый экземпляр Auth.
func New(
	log *slog.Logger,
	usersaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    usersaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		toketTTL:    tokenTTL,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Login аутентифицирует пользователя на основе предоставленных учетных данных.
//
// Параметры:
// - ctx: Контекст для управления запросом.
// - email: Электронная почта пользователя.
// - password: Пароль пользователя.
// - appID: Идентификатор приложения.
//
// Возвращает:
// - string: Токен аутентификации, если вход успешен.
// - error: Ошибка, если аутентификация не удалась.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)
	log.Info("attempting to login user")

	//Проверяем пользователя
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("failed to find user", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to find user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	//Проверяем пароль
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("failed to compare password", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	//Приложение
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("successfully logged in")

}

// RegisterNewUser регистрирует нового пользователя с указанным адресом электронной почты и паролем.
//
// Параметры:
// - ctx: Контекст для управления запросом.
// - email: Электронная почта пользователя.
// - pass: Пароль пользователя.
//
// Возвращает:
// - int64: Идентификатор нового пользователя.
// - error: Ошибка, если регистрация не удалась.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("Register New User")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash passHash", "error", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Сохранить в базе данных
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)

	if err != nil {
		log.Error("Failed to save user", "error", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Successfully saved user", "id", id)
	return id, nil
}

// IsAdmin check if user is admin
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	panic("implement me")
}
