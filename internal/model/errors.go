package model

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInvalid           = errors.New("invalid input")
	ErrForbidden         = errors.New("forbidden")
	ErrOutbid            = errors.New("the bid is lower than the current bid or the lot is closed")
	ErrClosed            = errors.New("lot closed")
	ErrLenLogin          = errors.New("login length must be between 3 and 255")
	ErrLenPass           = errors.New("password must be at least 8 characters")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidAuth       = errors.New("invalid login or password")
	ErrUserNotFound      = errors.New("user not found")
	ErrDatabase          = errors.New("database error")
	ErrStartPrice        = errors.New("invalid start price")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrStatusNotChanged  = errors.New("status not changed")
	ErrFileTooLarge      = errors.New("file too large")
	ErrPhotoNotFound     = errors.New("photo not found")
	ErrInvalidFileType   = errors.New("invalid file type")

	ErrInvalidID       = errors.New("invalid id")
	ErrUserIDNotFound  = errors.New("user_id not found")
	ErrInvalidUserType = errors.New("user_id has invalid type")

	// Tokens
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrSessionNotFound = errors.New("session not found")
)
