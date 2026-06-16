package mmm

import "github.com/spf13/viper"

func databasePath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if configured := viper.GetString("database.path"); configured != "" {
		return configured
	}
	return "mmm.db"
}

func serverAddress(host, port string) string {
	if host == "" {
		host = viper.GetString("server.host")
	}
	if host == "" {
		host = "0.0.0.0"
	}

	if port == "" {
		port = viper.GetString("server.port")
	}
	if port == "" {
		port = "8080"
	}

	return host + ":" + port
}
