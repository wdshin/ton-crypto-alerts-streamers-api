package utils

import "os"

func ValidateEnvVariables() {
	envVarNames := []string{
		"DB_NAME",
		"DB_CONNECTION",
		"DB_STREAMERS_COLLECTION_NAME",
		"DB_DONATIONS_COLLECTION_NAME",
		"DB_WIDGETS_COLLECTION_NAME",
		"PORT",
		"CONTRACT_ADDRESS",
		"TON_NET",
		"NOTIFICATION_URL",
		"TON_CONFIG_URL",
	}

	for _, envVarName := range envVarNames {
		envVarValue := os.Getenv(envVarName)
		if envVarValue == "" {
			panic("Failed to load environmant variable: " + envVarName)
		}
	}
}
