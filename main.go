package main

import (
	"os"

	pb "github.com/Kotlang/socialGo/generated"
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
	pb.RegisterUserPostServer(bootServer.GrpcServer, inject.FeedPostService)
	pb.RegisterPostActionsServer(bootServer.GrpcServer, inject.PostActionsService)
	pb.RegisterFollowGraphServer(bootServer.GrpcServer, inject.FollowGraphService)
	pb.RegisterSocialStatsServer(bootServer.GrpcServer, inject.SocialStatsService)

	bootServer.Start(grpcPort, webPort)
}
