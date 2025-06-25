package main

// import (
// 	"context"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"os/exec"
// 	"path/filepath"

// 	pb "test-rpc/proto"
// )

// type encoderServer struct {
// 	pb.UnimplementedEncoderServer
// }

// func main() {
// 	const chunkSize = 3 * 1024 * 1024

// 	file, err := os.Open(os.Args[1])

// 	if err!=nil{
// 		fmt.Println("Error opening file:", err)
// 		return
// 	}
// 	defer file.Close();

// 	buffer := make([]byte, chunkSize)

// }
