package utils

import (
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func GenerateUlid() string {
	return ulid.Make().String()
}

func GenerateUUID() string {
	return uuid.NewString()
}
