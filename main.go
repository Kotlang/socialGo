package main

import (
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/server"
)

var grpcPort = ":50051"
var webPort = ":8081"

func main() {
	inject := NewInject()

	bootServer := server.NewGoApiBoot()
	pb.RegisterUserPostServer(bootServer.GrpcServer, inject.FeedPostService)
	pb.RegisterPostActionsServer(bootServer.GrpcServer, inject.PostActionsService)

	bootServer.Start(grpcPort, webPort)
}
