package auth

import (
	"Llamacommunicator/pkg/config"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
)

type Service struct {
	conf   *config.Specification
	apiKey string
}

func NewAuthService(conf *config.Specification) *Service {
	apiKEy := conf.APIKey
	fmt.Println(apiKEy)
	return &Service{
		conf:   conf,
		apiKey: apiKEy,
	}
}

func (s *Service) ValidateAPIKey(c *fiber.Ctx, key string) (bool, error) {
	hashedAPIKey := sha256.Sum256([]byte(s.apiKey))
	hashedKey := sha256.Sum256([]byte(key))

	if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
		return true, nil
	}
	return false, keyauth.ErrMissingOrMalformedAPIKey
}
