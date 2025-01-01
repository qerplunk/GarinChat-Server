package envconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Environment variables that are required
type envConfig struct {
	AllowedOrigins []string
	JwtSecret      string
	Port           string
}

var EnvConfig *envConfig

// Keeps track if ALL environment variables were loaded correctly
var loadedAllEnvs bool

func InitEnvConfig() bool {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file")
	}

	loadedAllEnvs = true

	EnvConfig = &envConfig{
		AllowedOrigins: getEnvArray("ALLOWED_ORIGINS"),
		JwtSecret:      getEnv("JWT_DECODE_SECRET"),
		Port:           getEnv("PORT"),
	}

	return loadedAllEnvs
}

// Returns single environment variable
func getEnv(envName string) string {
	env := os.Getenv(envName)

	if len(env) == 0 {
		fmt.Println("No env for:", envName)
		loadedAllEnvs = false
	}

	return env
}

// Returns an environment variable array
func getEnvArray(envName string) []string {
	envStr := os.Getenv(envName)

	// Check the env string first before splitting into an array
	if len(envStr) == 0 {
		fmt.Println("No env for:", envName)
		loadedAllEnvs = false
	}

	envArray := strings.Split(envStr, ",")
	return envArray
}
