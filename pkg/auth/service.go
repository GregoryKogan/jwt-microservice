package auth

import (
	"errors"

	"github.com/GregoryKogan/jwt-microservice/pkg/auth/authjwt"
)

var ErrInvalidToken = errors.New("invalid token")

type AuthService interface {
	Authenticate(accessToken string) (*authjwt.JWTClaims, error)
	Login(userID uint) (*TokenPair, error)
	Refresh(refreshToken string) (*TokenPair, error)
	Logout(accessToken string) error
}

type AuthServiceImpl struct {
	repo       AuthRepo
	jwtService authjwt.JWTService
}

func NewAuthService(repo AuthRepo) AuthService {
	return &AuthServiceImpl{
		repo:       repo,
		jwtService: authjwt.NewJWTService(),
	}
}

func (s *AuthServiceImpl) Authenticate(accessToken string) (*authjwt.JWTClaims, error) {
	claims, err := s.jwtService.ParseToken(accessToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, errors.Join(ErrInvalidToken, errors.New("invalid token type"))
	}

	cached, err := s.repo.IsTokenCached(claims)
	if err != nil {
		return nil, err
	}

	if !cached {
		return nil, errors.Join(ErrInvalidToken, errors.New("token not found"))
	}

	s.repo.ExtendTokenPairCacheExpiration(claims.UserID)

	return claims, nil
}

func (s *AuthServiceImpl) Login(userID uint) (*TokenPair, error) {
	accessToken, err := s.jwtService.NewAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.NewRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	tokenPair := &TokenPair{
		Access:  accessToken,
		Refresh: refreshToken,
	}

	err = s.repo.CacheTokenPair(tokenPair)
	if err != nil {
		return nil, errors.Join(errors.New("failed to cache token pair"), err)
	}

	return tokenPair, nil
}

func (s *AuthServiceImpl) Refresh(refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtService.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, errors.Join(ErrInvalidToken, errors.New("invalid token type"))
	}

	ok, err := s.repo.IsTokenCached(claims)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.Join(ErrInvalidToken, errors.New("token not found"))
	}

	return s.Login(claims.UserID)
}

func (s *AuthServiceImpl) Logout(accessToken string) error {
	claims, err := s.jwtService.ParseToken(accessToken)
	if err != nil {
		return err
	}

	s.repo.DeleteTokenPair(claims.UserID)

	return nil
}
