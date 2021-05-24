package util

import (
	"github.com/spf13/viper"
	"os"
)

func GetEnv(key, fallback string) string {
	value := viper.GetString(key)
	if value == "" {
		value = os.Getenv(key)
		if value == "" {
			value = fallback
		}
	}
	return value
}
