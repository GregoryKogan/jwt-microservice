package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/GregoryKogan/jwt-microservice/pkg/auth/authjwt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var ErrInvalidTokenPair = errors.New("invalid token pair")

type AuthRepo interface {
	CacheTokenPair(tokenPair *TokenPair) error
	ExtendTokenPairCacheExpiration(userID uint)
	IsTokenCached(claims *authjwt.JWTClaims) (bool, error)
	DeleteTokenPair(userID uint)
}

type TokenPair struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type AuthRepoImpl struct {
	cache      *redis.Client
	jwtService authjwt.JWTService
}

func NewAuthRepo(cache *redis.Client) AuthRepo {
	return &AuthRepoImpl{
		cache:      cache,
		jwtService: authjwt.NewJWTService(),
	}
}

type tokenUIDPair struct {
	AccessUID  string `json:"access_uid"`
	RefreshUID string `json:"refresh_uid"`
}

func (r *AuthRepoImpl) CacheTokenPair(tokenPair *TokenPair) error {
	accessClaims, err := r.jwtService.ParseToken(tokenPair.Access)
	if err != nil {
		return err
	}

	refreshClaims, err := r.jwtService.ParseToken(tokenPair.Refresh)
	if err != nil {
		return err
	}

	if accessClaims.UserID != refreshClaims.UserID {
		return errors.Join(ErrInvalidTokenPair, errors.New("user IDs do not match"))
	}

	cached := tokenUIDPair{
		AccessUID:  accessClaims.UID,
		RefreshUID: refreshClaims.UID,
	}

	cacheJson, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	userID := accessClaims.UserID
	return r.cache.Set(context.Background(), fmt.Sprintf("token-%d", userID), cacheJson, viper.GetDuration("auth.auto_logout")).Err()
}

func (r *AuthRepoImpl) ExtendTokenPairCacheExpiration(userID uint) {
	go func() {
		r.cache.Expire(context.Background(), fmt.Sprintf("token-%d", userID), viper.GetDuration("auth.auto_logout"))
	}()
}

func (r *AuthRepoImpl) IsTokenCached(claims *authjwt.JWTClaims) (bool, error) {
	cacheJson, err := r.cache.Get(context.Background(), fmt.Sprintf("token-%d", claims.UserID)).Result()
	if err == redis.Nil {
		return false, nil // No such key in Redis cache
	} else if err != nil {
		return false, errors.Join(errors.New("failed to get token from cache"), err)
	}

	var cached tokenUIDPair
	if err := json.Unmarshal([]byte(cacheJson), &cached); err != nil {
		return false, errors.Join(errors.New("failed to unmarshal token from cache"), err)
	}

	var cachedUID string
	switch claims.Type {
	case "access":
		cachedUID = cached.AccessUID
	case "refresh":
		cachedUID = cached.RefreshUID
	default:
		return false, fmt.Errorf("invalid token type: %s", claims.Type)
	}

	return claims.UID == cachedUID, nil
}

func (r *AuthRepoImpl) DeleteTokenPair(userID uint) {
	r.cache.Del(context.Background(), fmt.Sprintf("token-%d", userID))
}
