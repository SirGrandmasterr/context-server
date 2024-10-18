package handler

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func Login(r *storage.StorageReader) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.FormValue("user")
		pass := c.FormValue("pass")

		// Throws Unauthorized error

		player, err := r.ReadPlayer(user, context.Background())
		if err != nil {
			print("Throwing unautzorized")
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		if !CheckPasswordHash(pass, player.Password) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// Create the Claims
		claims := jwt.MapClaims{
			"name":  player.Username,
			"admin": true,
			"exp":   time.Now().Add(time.Hour * 72).Unix(),
		}

		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("secret"))
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{"token": t})
	}
}

func CreateUser(r *storage.StorageReader, w *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.FormValue("user")
		pass := c.FormValue("pass")

		_, err := r.ReadPlayer(user, context.Background())
		if err == nil {
			//If no error is thrown a player was found that uses that username.
			c.Status(fiber.StatusForbidden)
			return c.Send([]byte("Username already taken. "))
		}
		hash, err := HashPassword(pass)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.Send([]byte("Something happened while creating the account."))
		}
		player := entities.Player{
			ID:       primitive.NewObjectID().String(),
			Username: user,
			Password: hash,
			History:  "",
		}
		err = w.SavePlayers(player, context.Background())
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.Send([]byte("Something happened while creating the account."))
		}
		return c.SendStatus(fiber.StatusCreated)

	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
