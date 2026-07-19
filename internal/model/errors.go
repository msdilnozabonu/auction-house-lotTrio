package model

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInvalid           = errors.New("invalid input")
	ErrForbidden         = errors.New("forbidden")
	ErrOutbid            = errors.New("the bid is lower than the current bid")
	ErrClosed            = errors.New("lot closed")
	ErrLenLogin          = errors.New("login length must be between 3 and 255")
	ErrLenPass           = errors.New("incorrect password")
	ErrIncorrectPassword = errors.New("incorrect password")
)
