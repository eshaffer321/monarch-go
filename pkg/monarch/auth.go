package monarch

import (
	"context"

	"github.com/eshaffer321/monarch-go/internal/auth"
	internalTypes "github.com/eshaffer321/monarch-go/internal/types"
)

// authService implements the AuthService interface
type authService struct {
	client  *Client
	service *auth.Service
}

// newAuthService creates a new auth service
func newAuthService(client *Client) *authService {
	return &authService{
		client: client,
		service: auth.NewService(
			client.baseURL,
			client.httpClient,
			client.options.Logger,
			client.options.UserAgent,
		),
	}
}

// convertSession converts internal types.Session to monarch.Session
func (a *authService) convertSession(s *internalTypes.Session) *Session {
	if s == nil {
		return nil
	}
	return &Session{
		Token:      s.Token,
		UserID:     s.UserID,
		Email:      s.Email,
		ExpiresAt:  s.ExpiresAt,
		DeviceUUID: s.DeviceUUID,
	}
}

// Login performs authentication
func (a *authService) Login(ctx context.Context, email, password string) error {
	if err := a.service.Login(ctx, email, password); err != nil {
		return err
	}

	// Get session and update client
	session, err := a.service.GetSession()
	if err != nil {
		return err
	}

	a.client.session = a.convertSession(session)
	a.client.transport.SetSession(session)

	// Save session if configured
	if a.client.options.SessionFile != "" {
		_ = a.service.SaveSession(a.client.options.SessionFile)
	}

	return nil
}

// LoginWithMFA performs login with MFA
func (a *authService) LoginWithMFA(ctx context.Context, email, password, mfaCode string) error {
	if err := a.service.LoginWithMFA(ctx, email, password, mfaCode); err != nil {
		return err
	}

	// Get session and update client
	session, err := a.service.GetSession()
	if err != nil {
		return err
	}

	a.client.session = a.convertSession(session)
	a.client.transport.SetSession(session)

	// Save session if configured
	if a.client.options.SessionFile != "" {
		_ = a.service.SaveSession(a.client.options.SessionFile)
	}

	return nil
}

// LoginWithEmailOTP performs login with email OTP code
func (a *authService) LoginWithEmailOTP(ctx context.Context, email, password, otpCode string) error {
	if err := a.service.LoginWithEmailOTP(ctx, email, password, otpCode); err != nil {
		return err
	}

	// Get session and update client
	session, err := a.service.GetSession()
	if err != nil {
		return err
	}

	a.client.session = a.convertSession(session)
	a.client.transport.SetSession(session)

	// Save session if configured
	if a.client.options.SessionFile != "" {
		_ = a.service.SaveSession(a.client.options.SessionFile)
	}

	return nil
}

// LoginWithTOTP performs login with TOTP secret
func (a *authService) LoginWithTOTP(ctx context.Context, email, password, totpSecret string) error {
	if err := a.service.LoginWithTOTP(ctx, email, password, totpSecret); err != nil {
		return err
	}

	// Get session and update client
	session, err := a.service.GetSession()
	if err != nil {
		return err
	}

	a.client.session = a.convertSession(session)
	a.client.transport.SetSession(session)

	// Save session if configured
	if a.client.options.SessionFile != "" {
		_ = a.service.SaveSession(a.client.options.SessionFile)
	}

	return nil
}

// LoginInteractive performs interactive login with prompts for MFA/OTP
func (a *authService) LoginInteractive(ctx context.Context, email, password string) error {
	if err := a.service.LoginInteractive(ctx, email, password); err != nil {
		return err
	}

	// Get session and update client
	session, err := a.service.GetSession()
	if err != nil {
		return err
	}

	a.client.session = a.convertSession(session)
	a.client.transport.SetSession(session)

	// Save session if configured
	if a.client.options.SessionFile != "" {
		_ = a.service.SaveSession(a.client.options.SessionFile)
	}

	return nil
}

// GetSession returns the current session
func (a *authService) GetSession() (*Session, error) {
	session, err := a.service.GetSession()
	if err != nil {
		return nil, err
	}
	return a.convertSession(session), nil
}

// SaveSession saves session to file
func (a *authService) SaveSession(path string) error {
	return a.service.SaveSession(path)
}

// LoadSession loads session from file
func (a *authService) LoadSession(path string) error {
	if err := a.service.LoadSession(path); err != nil {
		return err
	}

	// Get session and update client
	session, err := a.service.GetSession()
	if err != nil {
		return err
	}

	a.client.session = a.convertSession(session)
	a.client.transport.SetSession(session)

	return nil
}
