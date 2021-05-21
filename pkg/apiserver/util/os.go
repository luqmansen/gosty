package util

import (
	"github.com/spf13/viper"
)

func GetEnv(key, fallback string) string {
	value := viper.GetString(key)
	if value == "" {
		value = fallback
	}
	return value
}
