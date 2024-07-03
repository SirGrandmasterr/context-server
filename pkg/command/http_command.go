package command

import (
	"Llamacommunicator/api/router"
	"Llamacommunicator/pkg/auth"
	"Llamacommunicator/pkg/services/assistant"

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
	db, err := cmd.BaseCommand.NewDatabaseConnection()
	if err != nil {
		cmd.BaseCommand.log.Errorln("Error in DB Connection")
	}
	defer db.Close()
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	authService := auth.NewAuthService(cmd.config)

	app.Use(keyauth.New(keyauth.Config{
		Validator: authService.ValidateAPIKey,
	}))

	api := app.Group("/api")

	validator := validator.New()

	//Services
	assistantService := assistant.NewAssistantService(cmd.log, validator)

	router.AssistantRouter(api, assistantService)

	cmd.BaseCommand.log.Fatal(app.Listen(":8079"))
}
