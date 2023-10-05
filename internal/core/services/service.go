package services

import "github.com/NBN23dev/gcr-go-template/internal/core/ports"

type Service struct {
}

// NewService
func NewService(repo ports.Repository) (*Service, error) {
	return &Service{}, nil
}
