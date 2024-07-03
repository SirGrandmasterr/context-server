package command

import (
	"Llamacommunicator/pkg/config"
	"log"

	"github.com/boltdb/bolt"
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

func (cmd *BaseCommand) NewDatabaseConnection() (*bolt.DB, error) {

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//create Database if doesn't exist
	return db, nil

}
