package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "test-rpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	const chunkSize = 3 * 1024 * 1024 // 3 MB

	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewEncoderClient(conn)

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <file_path>", os.Args[0])
	}

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)
	var chunkCounter int32 = 0

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Fatalf("Error reading file: %v", err)
		}

		if n == 0 {
			break // Nothing more to send
		}

		fileData := buffer[:n]
		formattedName := fmt.Sprintf("%s_%d", filePath, chunkCounter)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

		resp, err := client.Encode(ctx, &pb.FileRequest{
			Filename:     formattedName,
			FileData:     fileData,
			Iteration:    chunkCounter,
			Filefullname: filePath,
		})
		cancel()

		if err != nil {
			log.Fatalf("Error during encoding: %v", err)
		}

		fmt.Printf("Chunk %d: Success: %v, Message: %s\n", chunkCounter, resp.Success, resp.Message)

		chunkCounter++

		if err == io.EOF {
			break
		}
	}
}
