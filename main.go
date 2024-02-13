package main

import (
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/SaiNageswarS/go-api-boot/server"
	"github.com/rs/cors"
)

var grpcPort = ":50051"
var webPort = ":8081"

func main() {
	// go-api-boot picks up keyvault name from environment variable.
	inject := NewInject()
	inject.cloudFns.LoadSecretsIntoEnv()

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
