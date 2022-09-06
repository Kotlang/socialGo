package main

import (
	"os"

	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/server"
)

var grpcPort = ":50051"
var webPort = ":8081"

func main() {
	// go-api-boot picks up keyvault name from environment variable.
	os.Setenv("AZURE-KEYVAULT-NAME", "kotlang-secrets")
	server.LoadSecretsIntoEnv(true)
	inject := NewInject()

	bootServer := server.NewGoApiBoot()
	pb.RegisterUserPostServer(bootServer.GrpcServer, inject.FeedPostService)
	pb.RegisterPostActionsServer(bootServer.GrpcServer, inject.PostActionsService)
	pb.RegisterFollowGraphServer(bootServer.GrpcServer, inject.FollowGraphService)
	pb.RegisterSocialStatsServer(bootServer.GrpcServer, inject.SocialStatsService)

	bootServer.Start(grpcPort, webPort)
}
