package handler

import (
	"github.com/gofiber/fiber/v2"
)

func Pong() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Status(200)

		return c.JSON(fiber.Map{"msg": "Pong."})
	}
}
