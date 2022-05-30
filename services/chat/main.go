package main

import (
	"context"
	"fmt"
	"log"
	"net"
	SYS "syscall"
	pb "vreco/chat/gen/chat/v1"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	DEATH "github.com/vrecan/death/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)

	s := grpc.NewServer(
		grpc.StreamInterceptor(middleware.ChainStreamServer(
			recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(
			recovery.UnaryServerInterceptor(),
		)),
	)
	pb.RegisterChatServiceServer(s, NewChatServer())
	reflection.Register(s)

	go func() {
		lis, err := net.Listen("tcp", ":2020")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	death.WaitForDeathWithFunc(func() {
		s.GracefulStop()
	})
}

type ChatServer struct {
}

func (s *ChatServer) SayHello(ctx context.Context, req *pb.SayHelloRequest) (msg *pb.SayHelloResponse, err error) {
	msg = &pb.SayHelloResponse{}
	response := fmt.Sprint("Hello there ", *req.Name)
	msg.Message = &response
	return msg, err
}

func NewChatServer() *ChatServer {
	return &ChatServer{}
}
