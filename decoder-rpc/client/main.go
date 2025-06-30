package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "decoder-rpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// const chunkSize = 3 * 1024 * 1024 // 3 MB

	conn, err := grpc.Dial(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure mode
		grpc.WithBlock(), // optional: block until connected
	)
	defer conn.Close()

	client := pb.NewDecoderClient(conn)

	filePath := os.Args[1]

	fmt.Println("Fetching metadata for file ...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	resp, err := client.Decode(ctx, &pb.FileRequest{
		Filename: filePath,
	})

	cancel()

	fmt.Printf("Success: %v\nMessage: %s\n ", resp.Success, resp.Message)

	if err != nil {
		log.Fatalf("Error during decoding: %v", err)
	}
}
