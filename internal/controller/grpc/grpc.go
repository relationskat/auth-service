package grpc

import (
	"github.com/relationskat/auth-service/internal/service"
	authv1 "github.com/relationskat/auth-service/pkg/gen"
	"go.uber.org/zap"
)

type GRPC struct {
	authv1.UnimplementedAuthServer
	service service.Service
	log     *zap.Logger
}

func New(log *zap.Logger, svc service.Service) *GRPC {
	return &GRPC{
		service: svc,
		log:     log.Named("grpc"),
	}
}
