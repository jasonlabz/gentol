package service

import "context"

type HealthCheckService interface {
	DoCheck(ctx context.Context) string
}
