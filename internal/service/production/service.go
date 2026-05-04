package production

import (
	"log/slog"

	"github.com/relationskat/auth-service/internal/repository"
)

type Service struct {
	log        *slog.Logger
	repository repository.Repository
}

func New(log *slog.Logger) (*Service, error) {
	service := &Service{
		log: log,
	}

	return service, nil
}
