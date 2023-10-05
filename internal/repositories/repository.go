package repositories

type Repository struct {
}

// NewRepository
func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}
