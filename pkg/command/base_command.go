package command

import (
	"Llamacommunicator/pkg/config"

	"github.com/gofiber/storage/mongodb/v2"

	"go.uber.org/zap"
)

// BaseCommand holds common command properties
type BaseCommand struct {
	config *config.Specification
	log    *zap.SugaredLogger
}

// NewBaseCommand creates a structure with common shared properties of the commands
func NewBaseCommand(config *config.Specification, logger *zap.SugaredLogger) BaseCommand {
	return BaseCommand{
		config: config,
		log:    logger,
	}
}

func (cmd *BaseCommand) NewDatabaseConnection() *mongodb.Storage {

	store := mongodb.New(mongodb.Config{
		ConnectionURI: cmd.config.DBConnLink,
		Database:      "llamadrama",
		Collection:    "llama_storage",
		Reset:         false,
	})

	return store

}
