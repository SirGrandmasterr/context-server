package config

import (
	"os"
)

type Specification struct {
	GeminiAPIKey string
	APIKey       string
	DBConnLink   string
}

func NewSpecification() *Specification {
	return &Specification{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		APIKey:       os.Getenv("API_KEY"),
		DBConnLink:   os.Getenv("DB_CONN_LINK"),
	}
}
