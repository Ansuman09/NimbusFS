package main

import (
	"fmt"
	"log"
	"net"

	pb "test-rpc/proto"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterEncoderServer(s, &encoderServer{})
	fmt.Println("gRPC server started on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
