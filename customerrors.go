package main

import "errors"

var (
	NotFoundError = errors.New("Not found")
)

// ErrorResponse a wrapper for error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Error   error  `json:"error"`
	Message string `json:"message"`
}
