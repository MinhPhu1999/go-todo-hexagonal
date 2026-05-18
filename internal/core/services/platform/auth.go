package platform

import (
	"context"
	"errors"
	"time"

	domain "go-crud-db-p2/internal/core/domain/platform"
	ports "go-crud-db-p2/internal/core/ports/platform"
)

type AuthService struct {
	userRepository ports.IUserRepository
	contextTimeout time.Duration
	idGenerator    ports.IUserIDGenerator
	passwords      ports.IPasswordHasher
	tokens         ports.ITokenProvider
	google         ports.IGoogleIdentityProvider
	states         ports.IAuthStateStore
	clock          ports.IClock
}

func NewAuthService(
	userRepository ports.IUserRepository,
	timeout time.Duration,
	idGenerator ports.IUserIDGenerator,
	passwords ports.IPasswordHasher,
	tokens ports.ITokenProvider,
	google ports.IGoogleIdentityProvider,
	states ports.IAuthStateStore,
	clock ports.IClock,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		contextTimeout: timeout,
		idGenerator:    idGenerator,
		passwords:      passwords,
		tokens:         tokens,
		google:         google,
		states:         states,
		clock:          clock,
	}
}

func (s *AuthService) Register(ctx context.Context, request domain.RegisterRequest) (*domain.AuthResponse, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	email, err := domain.NormalizeEmail(request.Email)
	if err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(request.Password); err != nil {
		return nil, err
	}

	if _, err := s.userRepository.FindByEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailAlreadyExists
	} else if !errors.Is(err, domain.ErrUnauthorized) {
		return nil, err
	}

	passwordHash, err := s.passwords.Hash(request.Password)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	user, err := domain.NewEmailUser(s.idGenerator.NewUserID(), email, request.Name, passwordHash, now)
	if err != nil {
		return nil, err
	}
	user.MarkLoggedIn(now)

	if err := s.userRepository.Save(ctx, user); err != nil {
		return nil, err
	}

	return s.newAuthResponse(user)
}

func (s *AuthService) Login(ctx context.Context, request domain.LoginRequest) (*domain.AuthResponse, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	email, err := domain.NormalizeEmail(request.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if !user.HasPassword() {
		return nil, domain.ErrInvalidCredentials
	}
	if err := s.passwords.Compare(user.PasswordHash, request.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	user.MarkLoggedIn(s.clock.Now())
	if err := s.userRepository.Save(ctx, user); err != nil {
		return nil, err
	}

	return s.newAuthResponse(user)
}

func (s *AuthService) GoogleLoginURL(ctx context.Context) (string, error) {
	_, cancel := s.context(ctx)
	defer cancel()

	state, err := s.states.Generate()
	if err != nil {
		return "", err
	}
	return s.google.AuthCodeURL(state)
}

func (s *AuthService) GoogleCallback(ctx context.Context, state string, code string) (*domain.AuthResponse, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	if !s.states.Verify(state) {
		return nil, domain.ErrGoogleInvalidOAuthState
	}

	profile, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	if !profile.EmailVerified {
		return nil, domain.ErrGoogleEmailNotVerified
	}

	return s.signInWithGoogleProfile(ctx, profile)
}

func (s *AuthService) signInWithGoogleProfile(ctx context.Context, profile domain.GoogleProfile) (*domain.AuthResponse, error) {
	user, err := s.userRepository.FindByGoogleID(ctx, profile.Subject)
	if err == nil {
		now := s.clock.Now()
		if err := user.AttachGoogleProfile(profile, now); err != nil {
			return nil, err
		}
		user.MarkLoggedIn(now)
		if err := s.userRepository.Save(ctx, user); err != nil {
			return nil, err
		}
		return s.newAuthResponse(user)
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		return nil, err
	}

	user, err = s.userRepository.FindByEmail(ctx, profile.Email)
	if err == nil {
		now := s.clock.Now()
		if err := user.AttachGoogleProfile(profile, now); err != nil {
			return nil, err
		}
		user.MarkLoggedIn(now)
		if err := s.userRepository.Save(ctx, user); err != nil {
			return nil, err
		}
		return s.newAuthResponse(user)
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		return nil, err
	}

	now := s.clock.Now()
	user, err = domain.NewGoogleUser(s.idGenerator.NewUserID(), profile, now)
	if err != nil {
		return nil, err
	}
	user.MarkLoggedIn(now)
	if err := s.userRepository.Save(ctx, user); err != nil {
		return nil, err
	}

	return s.newAuthResponse(user)
}

func (s *AuthService) CurrentUser(ctx context.Context, id string) (*domain.User, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	userID, err := domain.ParseUserID(id)
	if err != nil {
		return nil, err
	}

	return s.userRepository.FindByID(ctx, userID)
}

func (s *AuthService) newAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	token, err := s.tokens.IssueToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		TokenType:   "Bearer",
		AccessToken: token.Token,
		ExpiresAt:   token.ExpiresAt,
		User:        user,
	}, nil
}

func (s *AuthService) context(parent context.Context) (context.Context, context.CancelFunc) {
	if s.contextTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, s.contextTimeout)
}
