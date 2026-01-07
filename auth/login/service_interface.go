package login

import (
	"context"
)

// ServiceInterface defines the interface for login service operations
type ServiceInterface interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}

