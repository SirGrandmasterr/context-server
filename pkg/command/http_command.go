package command

import (
	"Llamacommunicator/api/router"
	"Llamacommunicator/pkg/auth"
	"Llamacommunicator/pkg/services/assistant"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/keyauth/v2"
	"github.com/urfave/cli"
)

type HttpCommand struct {
	BaseCommand
}

func NewHttpCommand(baseCommand BaseCommand) *HttpCommand {
	return &HttpCommand{
		BaseCommand: baseCommand,
	}
}

func (cmd *HttpCommand) Run(clictx *cli.Context) {
	db := cmd.BaseCommand.NewDatabaseConnection()
	defer db.Client().Disconnect(context.Background())
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	authService := auth.NewAuthService(cmd.Config)

	app.Use(keyauth.New(keyauth.Config{
		Validator: authService.ValidateAPIKey,
	}))

	api := app.Group("/api")

	validator := validator.New()

	//Services
	assistantService := assistant.NewAssistantService(cmd.Log, validator)

	router.AssistantRouter(api, assistantService)

	cmd.BaseCommand.Log.Fatal(app.Listen(":8079"))
}
