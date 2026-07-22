package platform

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go-crud-db-p2/config"
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
	emailSender    ports.IEmailSender
	otpLength      int
	otpExpiresIn   time.Duration
}

func NewAuthService(
	userRepository ports.IUserRepository,
	cfg config.Config,
	idGenerator ports.IUserIDGenerator,
	passwords ports.IPasswordHasher,
	tokens ports.ITokenProvider,
	google ports.IGoogleIdentityProvider,
	states ports.IAuthStateStore,
	clock ports.IClock,
	emailSender ports.IEmailSender,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		contextTimeout: cfg.Context.Timeout,
		idGenerator:    idGenerator,
		passwords:      passwords,
		tokens:         tokens,
		google:         google,
		states:         states,
		clock:          clock,
		emailSender:    emailSender,
		otpLength:      cfg.OTP.Length,
		otpExpiresIn:   cfg.OTP.ExpiresIn,
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
	if user.IsAccountLocked(s.clock.Now()) {
		return nil, domain.ErrAccountLocked
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

func (s *AuthService) ForgotPassword(ctx context.Context, request domain.ForgotPasswordRequest) error {
	ctx, cancel := s.context(ctx)
	defer cancel()

	email, err := domain.NormalizeEmail(request.Email)
	if err != nil {
		return err
	}

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil
		}
		return err
	}

	otp, err := generateOTP(s.otpLength)
	if err != nil {
		return err
	}

	now := s.clock.Now()
	otpHash := hashOTP(otp)
	otpExpiresAt := now.Add(s.otpExpiresIn)

	user.RequestPasswordReset(otpHash, otpExpiresAt, now)
	if err := s.userRepository.Save(ctx, user); err != nil {
		return err
	}

	if err := s.emailSender.SendOTP(user.Email, otp); err != nil {
		user.ResetOTPHash = ""
		user.ResetOTPExpiresAt = nil
		user.AccountLockedUntil = nil
		user.UpdatedAt = s.clock.Now().UTC()
		_ = s.userRepository.Save(ctx, user)
		return err
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, request domain.ResetPasswordRequest) error {
	ctx, cancel := s.context(ctx)
	defer cancel()

	email, err := domain.NormalizeEmail(request.Email)
	if err != nil {
		return err
	}
	if err := domain.ValidatePassword(request.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		return domain.ErrInvalidOTP
	}

	if request.OTP == "" {
		return domain.ErrInvalidOTP
	}

	now := s.clock.Now()

	if user.ResetOTPHash == "" || user.ResetOTPExpiresAt == nil {
		return domain.ErrInvalidOTP
	}

	otpHash := hashOTP(request.OTP)
	if user.ResetOTPHash != otpHash {
		return domain.ErrInvalidOTP
	}

	if now.After(*user.ResetOTPExpiresAt) {
		return domain.ErrOTPExpired
	}

	passwordHash, err := s.passwords.Hash(request.NewPassword)
	if err != nil {
		return err
	}

	user.ResetPassword(passwordHash, now)
	if err := s.userRepository.Save(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID domain.UserID, request domain.UpdateProfileRequest) (*domain.User, error) {
	ctx, cancel := s.context(ctx)
	defer cancel()

	if err := request.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.UpdateProfile(request, s.clock.Now())

	if err := s.userRepository.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID domain.UserID, request domain.ChangePasswordRequest) error {
	ctx, cancel := s.context(ctx)
	defer cancel()

	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.HasPassword() {
		return domain.ErrInvalidPassword
	}
	if err := s.passwords.Compare(user.PasswordHash, request.CurrentPassword); err != nil {
		return domain.ErrInvalidPassword
	}
	if err := domain.ValidatePassword(request.NewPassword); err != nil {
		return err
	}

	passwordHash, err := s.passwords.Hash(request.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = s.clock.Now().UTC()
	if err := s.userRepository.Save(ctx, user); err != nil {
		return err
	}

	return nil
}

func generateOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	max := big.NewInt(int64(10))
	max.Exp(max, big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

func hashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
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
