package main

import (
	"log"
	"os"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/providers"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/script"
	"github.com/samber/do"

	"github.com/gin-gonic/gin"
)

func args(injector *do.Injector) bool {
	if len(os.Args) > 1 {
		flag := script.Commands(injector)
		return flag
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
	var (
		injector = do.New()
	)

	providers.RegisterDependencies(injector)

	if !args(injector) {
		return
	}

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

	trip.RegisterRoutes(server, injector)
	dispatch.RegisterRoutes(server, injector)
	dispatchws.RegisterRoutes(server, injector)

	run(server)
}
