package handler

import (
	"Llamacommunicator/api/presenter"
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/services/assistant"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func RegisterPlayerAction(service *assistant.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		var requestBody entities.WebSocketMessage
		err := c.BodyParser(&requestBody)
		if err != nil {
			service.Log.Errorln(err)
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NewAssistantErrorResponse(err))
		}
		err = service.Val.Struct(requestBody)
		if err != nil {
			service.Log.Errorln(err)
			c.Status(http.StatusBadRequest)
			return c.JSON(presenter.NewAssistantErrorResponse(err))
		}
		chosenAction, err := service.AskAssistant(ctx, requestBody)
		if err != nil {
			service.Log.Errorln(err)
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.NewAssistantErrorResponse(err))
		}

		c.Status(200)
		return c.JSON(chosenAction)
	}
}
