package main

//go:generate go run github.com/swaggo/swag/cmd/swag init
//go:generate go run github.com/google/wire/cmd/wire

import (
	"github.com/tarkiman/go/configs"
	"github.com/tarkiman/go/shared/logger"
)

var config *configs.Config

// @securityDefinitions.apikey OauthToken
// @in header
// @name Authorization
func main() {
	// Initialize logger
	logger.InitLogger()

	// Initialize config
	config = configs.Get()

	// Set desired log level
	logger.SetLogLevel(config)

	// Wire everything up
	http := InitializeService()

	// Run server
	http.SetupAndServe()
}
