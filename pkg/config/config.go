package config

import (
	"os"
)

type Specification struct {
	LlmBig     string
	LlmSmall   string
	DBConnLink string
	Secret     string
}

func NewSpecification() *Specification {
	return &Specification{
		LlmBig:     os.Getenv("LLM_BIG"),
		LlmSmall:   os.Getenv("LLM_SMALL"),
		DBConnLink: os.Getenv("DB_CONN_LINK"),
		Secret:     os.Getenv("JWT_SECRET"),
	}
}
