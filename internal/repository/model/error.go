package model

import "errors"

var (
	ErrRecordNotFound   = errors.New("record not found") // requested record is not found
	ErrDuplicateService = errors.New("duplicate service")
)
