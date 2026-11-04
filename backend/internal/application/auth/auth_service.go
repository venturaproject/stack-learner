package auth

import (
	"context"
	"time"

	"github.com/alexedwards/argon2id"

	appshared "boilerplate-learner/internal/application/shared"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/shared/apperrors"
)

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Name     *string `json:"name"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
}

type AuthService struct {
	repo domainauth.Repository
	read ReadRepository
	jwt  *JWTService
	tx   appshared.TxManager
	pub  appshared.EventPublisher
}

func NewAuthService(repo domainauth.Repository, read ReadRepository, jwt *JWTService, tx appshared.TxManager, pub appshared.EventPublisher) *AuthService {
	return &AuthService{repo: repo, read: read, jwt: jwt, tx: tx, pub: pub}
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*UserWithDetails, *TokenPair, error) {
	user, err := s.repo.FindUserByEmailOrUsername(ctx, req.Login)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, apperrors.Unauthorized("Credenciales incorrectas")
	}

	match, err := argon2id.ComparePasswordAndHash(req.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, nil, apperrors.Unauthorized("Credenciales incorrectas")
	}
	if user.Status != "active" {
		return nil, nil, apperrors.Unauthorized("Tu cuenta está desactivada")
	}

	tokens, err := s.jwt.GenerateTokens(user.ID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.repo.CreateRefreshToken(ctx, domainauth.RefreshToken{
		JTI:       tokens.RefreshJTI,
		UserID:    user.ID,
		ExpiresAt: tokens.RefreshExpiresAt,
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, nil, err
	}

	details, err := s.read.LoadUserWithDetails(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	_ = s.pub.Publish(ctx, domainauth.NewUserLoggedIn(user.ID, user.Email))
	return details, tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshJTI string) error {
	return s.repo.RevokeRefreshToken(ctx, refreshJTI)
}

func (s *AuthService) Refresh(ctx context.Context, refreshJTI string) (*TokenPair, error) {
	rt, err := s.repo.FindRefreshTokenByJTI(ctx, refreshJTI)
	if err != nil {
		return nil, err
	}
	if rt == nil || rt.Revoked {
		return nil, apperrors.Unauthorized("Token inválido o revocado")
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, apperrors.Unauthorized("Token expirado")
	}
	tokens, err := s.jwt.GenerateTokens(rt.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.tx.WithinTx(ctx, func(txctx context.Context) error {
		user, err := s.repo.LockUser(txctx, rt.UserID)
		if err != nil {
			return err
		}
		if user == nil || user.Status != "active" {
			return apperrors.Unauthorized("Tu cuenta está desactivada")
		}
		consumed, err := s.repo.ConsumeRefreshToken(txctx, refreshJTI, time.Now())
		if err != nil {
			return err
		}
		if !consumed {
			return apperrors.Unauthorized("Token inválido o ya utilizado")
		}
		return s.repo.CreateRefreshToken(txctx, domainauth.RefreshToken{
			JTI:       tokens.RefreshJTI,
			UserID:    rt.UserID,
			ExpiresAt: tokens.RefreshExpiresAt,
			CreatedAt: time.Now(),
		})
	}); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID int32) (*UserWithDetails, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.NotFound("Usuario no encontrado")
	}
	return s.read.LoadUserWithDetails(ctx, user.ID)
}

func (s *AuthService) UpdateMe(ctx context.Context, userID int32, req UpdateProfileRequest) (*UserWithDetails, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperrors.NotFound("Usuario no encontrado")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Username != nil {
		user.Username = req.Username
	}
	if req.Email != nil {
		existing, _ := s.repo.FindUserByEmail(ctx, *req.Email)
		if existing != nil && existing.ID != userID {
			return nil, apperrors.Conflict("Este email ya está en uso")
		}
		user.Email = *req.Email
	}
	user.UpdatedAt = time.Now()

	if _, err := s.repo.UpdateUser(ctx, *user); err != nil {
		return nil, err
	}
	return s.read.LoadUserWithDetails(ctx, userID)
}

func (s *AuthService) UploadAvatar(ctx context.Context, userID int32, avatarURL string) (*UserWithDetails, error) {
	if err := s.repo.UpsertUserSettings(ctx, domainauth.UserSettings{UserID: userID, Avatar: &avatarURL}); err != nil {
		return nil, err
	}
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		return nil, apperrors.NotFound("Usuario no encontrado")
	}
	return s.read.LoadUserWithDetails(ctx, userID)
}
