package main

import (
	"context"
	"log"
	"net"
	"net/http"
	SYS "syscall"
	pb "vreco/chat/gen/chat/v1"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
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

	//twirp server
	go func() {
		twirpServer := &ChatServer{}
		twirpHandler := pb.NewChatServiceServer(twirpServer)

		http.ListenAndServe(":8080", twirpHandler)
	}()
	//grpc server
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

func (s *ChatServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (msg *pb.SendMessageResponse, err error) {
	msg = &pb.SendMessageResponse{}
	success := true
	msg.Success = &success
	return msg, err
}

func (s *ChatServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (msg *pb.GetMessagesResponse, err error) {
	msg = &pb.GetMessagesResponse{}
	return msg, err
}

func NewChatServer() *ChatServer {
	return &ChatServer{}
}
