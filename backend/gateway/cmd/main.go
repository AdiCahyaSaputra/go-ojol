package main

import (
	"log"
	"os"

	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/providers"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/script"
	"github.com/joho/godotenv"
	"github.com/samber/do"

	"github.com/gin-gonic/gin"
)

func args() bool {
	if len(os.Args) > 1 {
		return script.Commands()
	}

	return true
}

func run(server *gin.Engine) {
	server.Static("/assets", "./assets")

	port := os.Getenv("GOLANG_PORT")
	if port == "" {
		port = "8888"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "0.0.0.0:" + port
	} else {
		serve = ":" + port
	}

	if err := server.Run(serve); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}

func main() {
	_ = godotenv.Load(".env")

	injector := do.New()
	providers.RegisterDependencies(injector)

	if !args() {
		return
	}

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

	// Register module routes here

	run(server)
}
