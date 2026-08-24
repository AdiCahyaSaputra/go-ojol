package main

import (
	"log"
	"os"

	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/providers"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/proxy"
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

	if os.Getenv("APP_ENV") != "localhost" {
		gin.SetMode(gin.ReleaseMode)
	}

	injector := do.New()
	providers.RegisterDependencies(injector)

	if !args() {
		return
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	tripURL := os.Getenv("TRIP_SERVICE_URL")
	if authURL == "" || tripURL == "" {
		log.Fatal("AUTH_SERVICE_URL and TRIP_SERVICE_URL are required")
	}

	server := gin.New()
	server.Use(gin.Logger())
	server.Use(middlewares.Recovery())
	server.Use(middlewares.CORSMiddleware())

	proxyServers := []proxy.ProxyCfg{
		{
			Name: "Auth Proxy",
			Url:  authURL,
			UrlPaths: []string{
				"/.well-known/jwks.json",
				"/api/auth",
				"/api/auth/*path",
			},
		},
		{
			Name: "Trip Proxy",
			Url:  tripURL,
			UrlPaths: []string{
				"/api/trip",
				"/api/trip/*path",
			},
		},
	}

	if err := proxy.Register(server, proxyServers); err != nil {
		log.Fatalf("register proxy: %v", err)
	}

	run(server)
}
