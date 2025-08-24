package usecase

import "context"

// OLTRepository defines the interface for OLT repository operations
type OLTRepository interface {
	FirstByID(oltID int) (string, bool)
}

type FetchOLTIPByID interface {
	Execute(ctx context.Context, oltID int) (string, bool)
}

type fetchOLTIPByID struct {
	repository OLTRepository
}

func NewFetchOLTIPByID(repository OLTRepository) FetchOLTIPByID {
	return &fetchOLTIPByID{
		repository: repository,
	}
}

// Execute implements FetchOLTIPByID.
func (uc *fetchOLTIPByID) Execute(ctx context.Context, oltID int) (string, bool) {
	return uc.repository.FirstByID(oltID)
}
