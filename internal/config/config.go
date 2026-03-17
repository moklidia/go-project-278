package config

import "os"

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = fallback
	}

	return value
}

