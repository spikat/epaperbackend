package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               int
	Debug              bool
	DebugPort          int
	DataDir            string
	DebugPreviewWidth  int
	DebugPreviewHeight int
	MainAPIBaseURL     string
}

func Load() Config {
	port := getIntEnv("PORT", 5678)
	debugPort := getIntEnv("DEBUG_PORT", 4242)
	return Config{
		Port:               port,
		Debug:              getBoolEnv("DEBUG", false),
		DebugPort:          debugPort,
		DataDir:            getStringEnv("DATA_DIR", "./data"),
		DebugPreviewWidth:  getIntEnv("DEBUG_PREVIEW_WIDTH", 800),
		DebugPreviewHeight: getIntEnv("DEBUG_PREVIEW_HEIGHT", 480),
		MainAPIBaseURL:     getStringEnv("MAIN_API_BASE_URL", "http://127.0.0.1:"+strconv.Itoa(port)),
	}
}

func getStringEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getBoolEnv(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
