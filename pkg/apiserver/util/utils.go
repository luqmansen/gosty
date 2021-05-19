package util

import (
	"github.com/google/uuid"
	"strings"
)

func GenerateID() string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)
}
