package main

import (
	"fmt"
	"log"
	"net"

	pb "decoder-rpc/proto"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDecoderServer(s, &decoderServer{})
	fmt.Println("gRPC server started on port 50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
