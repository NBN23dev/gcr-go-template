package ports

// Generate service based mocks
//go:generate mockgen -source=./ports.go -destination=../../mocks/service.go -package=mocks github.com/NBN23dev/gcr-go-template

// Services
type Service interface {
	// TODO: Add service functions if needed
}

// Repositories
type Repository interface {
	// TODO: Add repository functions if needed
}
