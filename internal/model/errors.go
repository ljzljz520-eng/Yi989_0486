package model

import "errors"

var (
	ErrNotFound       = errors.New("entity not found")
	ErrConflict       = errors.New("version conflict")
	ErrNoStock        = errors.New("batch has no stock")
	ErrInvalidState   = errors.New("invalid state transition")
	ErrUnauthorized   = errors.New("permission denied")
	ErrAlreadyExists  = errors.New("entity already exists")
	ErrInvalidRequest = errors.New("invalid request")
)
