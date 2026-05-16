package grpc

import (
	"github.com/relationskat/auth-service/internal/service"
	"go.uber.org/zap"
)

type GRPC struct {
	service service.Service
	log     *zap.Logger
}
