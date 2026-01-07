package hydra

import (
	"context"
)

// MockClient is a mock implementation of Hydra client for testing
type MockClient struct {
	AcceptLoginRequestFunc func(ctx context.Context, challenge string, req *AcceptLoginRequest) (*AcceptLoginResponse, error)
	GetLoginRequestFunc     func(ctx context.Context, challenge string) (*GetLoginRequest, error)
	RejectLoginRequestFunc  func(ctx context.Context, challenge string, errorCode string, errorDescription string) error
}

// AcceptLoginRequest accepts a login request (mock implementation)
func (m *MockClient) AcceptLoginRequest(ctx context.Context, challenge string, req *AcceptLoginRequest) (*AcceptLoginResponse, error) {
	if m.AcceptLoginRequestFunc != nil {
		return m.AcceptLoginRequestFunc(ctx, challenge, req)
	}
	// Default mock behavior - return success
	return &AcceptLoginResponse{
		RedirectTo: "http://example.com/callback?code=mock_code",
	}, nil
}

// GetLoginRequest retrieves login request (mock implementation)
func (m *MockClient) GetLoginRequest(ctx context.Context, challenge string) (*GetLoginRequest, error) {
	if m.GetLoginRequestFunc != nil {
		return m.GetLoginRequestFunc(ctx, challenge)
	}
	// Default mock behavior
	return &GetLoginRequest{
		Challenge: challenge,
		Subject:   "mock_user",
	}, nil
}

// RejectLoginRequest rejects a login request (mock implementation)
func (m *MockClient) RejectLoginRequest(ctx context.Context, challenge string, errorCode string, errorDescription string) error {
	if m.RejectLoginRequestFunc != nil {
		return m.RejectLoginRequestFunc(ctx, challenge, errorCode, errorDescription)
	}
	return nil
}

