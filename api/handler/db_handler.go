package handler

import (
	"Llamacommunicator/api/presenter"
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func CreateAction(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		var requestBody entities.Action
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}
		err = r.SaveActionOptionEntity2(requestBody, ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}

		c.Status(200)
		return c.JSON(fiber.Map{"created": "true"})
	}
}

func UpdateAction(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Throws Unauthorized error

		return c.JSON(fiber.Map{"updated": "true"})
	}
}

func CreateLocation(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		var requestBody entities.Location
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}
		err = r.SaveLocations(requestBody, ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}

		c.Status(200)
		return c.JSON(fiber.Map{"created": "true"})
	}
}
func UpdateLocation(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Throws Unauthorized error

		return c.JSON(fiber.Map{"updated": "true"})
	}
}

func CreateObject(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		var requestBody entities.RelevantObject
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}
		err = r.SaveObject(requestBody, ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}

		c.Status(200)
		return c.JSON(fiber.Map{"created": "true"})
	}
}
func UpdateObject(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Throws Unauthorized error

		return c.JSON(fiber.Map{"updated": "true"})
	}
}
func CreateBaseprompt(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		var requestBody entities.BasePrompt
		err := c.BodyParser(&requestBody)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}
		err = r.SaveBasePrompt(requestBody, ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NewActionErrorResponse(err))
		}

		c.Status(200)
		return c.JSON(fiber.Map{"created": "true"})
	}
}
func UpdateBaseprompt(r *storage.StorageWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Throws Unauthorized error

		return c.JSON(fiber.Map{"updated": "true"})
	}
}
