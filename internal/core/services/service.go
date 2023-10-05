package services

import "github.com/NBN23dev/go-service-template/internal/core/ports"

type Service struct {
}

// NewService
func NewService(repo ports.Repository) (*Service, error) {
	return &Service{}, nil
}
