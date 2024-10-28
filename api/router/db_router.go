package router

import (
	"Llamacommunicator/api/handler"
	"Llamacommunicator/pkg/storage"

	"github.com/gofiber/fiber/v2"
)

func DBRouter(app fiber.Router, w *storage.StorageWriter) {
	//app.Post("/create", handler.RequestReaction(service))
	app.Post("/object", handler.CreateAction(w))
	app.Post("/location", handler.CreateAction(w))
	app.Post("/action", handler.CreateAction(w))
	app.Post("/baseprompt", handler.CreateAction(w))
}
