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
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure mode
		grpc.WithBlock(), // optional: block until connected
	)
	defer conn.Close()

	client := pb.NewEncoderClient(conn)

	filePath := os.Args[1]

	// fileData, err := os.ReadFile(filePath)
	// if err != nil {
	// 	log.Fatalf("Failed to read file: %v", err)
	// }

	// open the file and send it in chunks
	// returns *os.File pointer
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	buffer := make([]byte, chunkSize)

	defer file.Close()

	var chunkCounter int32 = 0
	for {

		//change pointer by buffer
		n, fileerr := io.ReadFull(file, buffer)
		print(n)

		// Process in chunk
		var fileData []byte = buffer[:n]
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		formattedName := fmt.Sprintf("%s_%d", os.Args[1], chunkCounter)
		resp, err := client.Encode(ctx, &pb.FileRequest{
			Filename:     formattedName,
			FileData:     fileData,
			Iteration:    chunkCounter,
			Filefullname: filePath,
		})

		cancel()

		fmt.Printf("Success: %v\nMessage: %s\n ", resp.Success, resp.Message)

		if err != nil {
			log.Fatalf("Error during encoding: %v", err)
		}

		// send the last chunks and break off
		if fileerr == io.EOF {
			break // End of file reached
		} else if fileerr == io.ErrUnexpectedEOF {
			break // Less than chunkSize bytes left; we skip them
		} else if fileerr != nil {
			fmt.Println("Read error:", err)
			break
		}

		chunkCounter++

	}

}
