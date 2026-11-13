package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user with this email or telegram id already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("forbidden")
	ErrFragmentNotFound    = errors.New("fragment not found")
	ErrAlreadySubmitted       = errors.New("task fragment already submitted by this user")
	ErrInvalidTelegramHash    = errors.New("invalid telegram initData signature")
	ErrProfanityDetected      = errors.New("обнаружена ненормативная лексика")
	ErrInvalidTranslationText = errors.New("перевод слишком короткий или некорректный")
)

