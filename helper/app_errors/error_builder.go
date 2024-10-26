package app_errors

import "github.com/pkg/errors"

var (
	ErrNotFound          = errors.New("Not found")
	ErrNoCtxMetaData     = errors.New("No ctx metadata")
	ErrInvalidSessionId  = errors.New("Invalid session id")
	ErrEmailExists       = errors.New("Email already exists")
	ErrAssociationExists = errors.New("Association already exists")
	ErrInvalidRequest    = errors.New("Invalid request")
	ErrInternalError     = errors.New("Internal error")
)
