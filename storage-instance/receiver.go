package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":9443")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on port 9443...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	passedArgs, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading file name : ", err)
		return

	}
	allArgs := strings.Fields(strings.TrimSpace(passedArgs))

	if len(allArgs) < 2 {
		fmt.Println("Invalid input: expected 'JOB filename', got:", allArgs)
		conn.Close() // explicitly close it here too
		return
	}

	job := strings.ToUpper(allArgs[0])
	filename := allArgs[1]

	if err != nil {
		fmt.Println("Unable to read the job :", err)
	}

	switch job {
	case "UPLOAD":
		outFile, err := os.Create(filename)
		if err != nil {
			fmt.Println("Unable to create new file:", err)
			return
		}

		defer outFile.Close()

		bytesWritten, err := io.Copy(outFile, reader)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}

		fmt.Printf("File %s received successfuly size %d bytes\n", filename, bytesWritten)

	case "DOWNLOAD":
		inFile, err := os.Open(filename)

		if err != nil {
			fmt.Printf("unable to open downloadable file\n", err)
			return
		}

		defer inFile.Close()

		_, err = io.Copy(conn, inFile)
		if err != nil {
			fmt.Println("Error sending file to client", err)
		} else {
			fmt.Printf("Sent file %s", filename)
		}

	}

}
