package router

import (
	"Llamacommunicator/api/handler"
	"Llamacommunicator/pkg/services/evaluation"
	"Llamacommunicator/pkg/storage"

	"github.com/gofiber/fiber/v2"
)

func EvalRouter(app fiber.Router, w *storage.StorageWriter, r *storage.StorageReader, srv *evaluation.EvalService) {
	//app.Post("/create", handler.RequestReaction(service))
	app.Post("/actionselectionprecisionmusic", handler.TestActionSelectionPrecisionMusic(r, w, srv))
	app.Post("/actionselectionprecisionfollow", handler.TestActionSelectionPrecisionFollow(r, w, srv))
	app.Post("/actionselectionprecisionfollownowalk", handler.TestActionSelectionPrecisionFollowNoWalk(r, w, srv))
	app.Post("/provideArtInformationSpeech", handler.TestartInformationSpeech(r, w, srv))
	app.Post("/action", handler.CreateAction(w))
	app.Post("/baseprompt", handler.CreateAction(w))
}
