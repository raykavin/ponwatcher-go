package repository

import "pon_watcher/internal/usecase"

// InMemoryOLTRepository provides an in-memory implementation of OLTRepository
type InMemoryOLTRepository struct {
	repository map[int]string
}

// Ensure the struct implements all interface methods
var _ usecase.OLTRepository = (*InMemoryOLTRepository)(nil)

// NewInMemoryOLTRepository creates a new in-memory OLT repository
func NewInMemoryOLTRepository() *InMemoryOLTRepository {
	return &InMemoryOLTRepository{
		repository: map[int]string{
			1: "10.100.2.1", 
			2: "10.100.2.2", 
			3: "10.100.2.3", 
			4: "10.100.2.4", 
			5: "10.100.2.5",  
			6: "10.100.2.6",  
		},
	}
}

// FirstByID retrieves the OLT IP address by OLT ID
func (r *InMemoryOLTRepository) FirstByID(oltID int) (string, bool) {
	ip, exists := r.repository[oltID]
	return ip, exists
}
