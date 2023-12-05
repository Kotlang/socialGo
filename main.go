package main

import (
	"os"

	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/server"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

var grpcPort = ":50051"
var webPort = ":8081"

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", zap.Error(err))
	}
}

func main() {
	// go-api-boot picks up keyvault name from environment variable.
	os.Setenv("AZURE-KEYVAULT-NAME", "kotlang-secrets")
	server.LoadSecretsIntoEnv(true)
	inject := NewInject()

	corsConfig := cors.New(
		cors.Options{
			AllowedHeaders: []string{"*"},
		})
	bootServer := server.NewGoApiBoot(corsConfig)
	socialPb.RegisterUserPostServer(bootServer.GrpcServer, inject.FeedPostService)
	socialPb.RegisterActionsServer(bootServer.GrpcServer, inject.ActionsService)
	socialPb.RegisterFollowGraphServer(bootServer.GrpcServer, inject.FollowGraphService)
	socialPb.RegisterSocialStatsServer(bootServer.GrpcServer, inject.SocialStatsService)
	socialPb.RegisterEventsServer(bootServer.GrpcServer, inject.EventService)

	bootServer.Start(grpcPort, webPort)
}
