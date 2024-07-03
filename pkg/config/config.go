package config

import (
	"os"
)

type Specification struct {
	GeminiAPIKey string
	APIKey       string
}

func NewSpecification() *Specification {
	return &Specification{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		APIKey:       os.Getenv("API_KEY"),
	}
}
