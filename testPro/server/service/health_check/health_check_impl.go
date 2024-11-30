package health_check

import (
	"context"
	"sync"

	"testPro/server/service"
)

var svc *Service
var once sync.Once

func GetService() service.HealthCheckService {
	if svc != nil {
		return svc
	}
	once.Do(func() {
		svc = &Service{}
	})

	return svc
}

type Service struct {
}

func (s Service) DoCheck(ctx context.Context) string {
	return "success"
}
