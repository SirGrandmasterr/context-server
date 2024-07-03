package router

import (
	"Llamacommunicator/api/handler"
	"Llamacommunicator/pkg/services/assistant"

	"github.com/gofiber/fiber/v2"
)

func AssistantRouter(app fiber.Router, service *assistant.Service) {
	app.Post("/reaction", handler.RequestReaction(service))
}
