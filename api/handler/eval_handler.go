package handler

import (
	"Llamacommunicator/pkg/services/evaluation"
	"Llamacommunicator/pkg/storage"

	"github.com/gofiber/fiber/v2"
)

func TestActionSelectionPrecisionMusic(r *storage.StorageReader, w *storage.StorageWriter, eserv *evaluation.EvalService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		str := eserv.TestActionSelectionPrecision()
		return c.JSON(fiber.Map{"answer": str})

	}
}
func TestActionSelectionPrecisionFollow(r *storage.StorageReader, w *storage.StorageWriter, eserv *evaluation.EvalService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		str := eserv.TestActionSelectionPrecisionFollowPlayer()
		return c.JSON(fiber.Map{"answer": str})

	}
}
func TestActionSelectionPrecisionFollowNoWalk(r *storage.StorageReader, w *storage.StorageWriter, eserv *evaluation.EvalService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		str := eserv.TestActionSelectionPrecisionFollowPlayerNoWalk()
		return c.JSON(fiber.Map{"answer": str})

	}
}

func TestartInformationSpeech(r *storage.StorageReader, w *storage.StorageWriter, eserv *evaluation.EvalService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		str := eserv.CreateArtInformationNeedleHaystackPrompt(r, w)
		return c.JSON(fiber.Map{"answer": str})
	}
}
