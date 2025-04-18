package errors

import "errors"

var (
	UserNotExist          = errors.New("пользователя с данным Email не существует")
	PasswordWrong         = errors.New("неверный пароль")
	EmailAlreadyExist     = errors.New("данный email уже существует")
	InvalidEmailFormat    = errors.New("некорректный формат email")
	InvalidPasswordFormat = errors.New("некорректный формат пароля. Пароль должен содержать не менее 8 символов, содержать номер, букву и специальный символ")
)

var (
	ErrRefreshTokenRevoked = errors.New("данный refresh token был отозван")
	ErrRefreshTokenExpired = errors.New("срок действия данного refresh token истёк")
)
