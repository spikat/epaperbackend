package config

import (
	"os"
	"strconv"
	"strings"
)

func serviceEnvKey(serviceName, configName string) string {
	return "SERVICE_" + strings.ToUpper(serviceName) + "_" + strings.ToUpper(configName)
}

func GetServiceString(serviceName, configName, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(serviceEnvKey(serviceName, configName))); v != "" {
		return v
	}
	return fallback
}

func GetServiceInt(serviceName, configName string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(serviceEnvKey(serviceName, configName)))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
