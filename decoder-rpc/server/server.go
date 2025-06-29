package main

import (
	"bufio"
	"context"
	"database/sql"
	pb "decoder-rpc/proto"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func ReadConfigFileAndInitalizeServers(nameserver *map[string]string, data_parity_servers *map[string]string, config string) error {

	file, err := os.Open(config)
	if err != nil {
		fmt.Println("File not found", err)
		return err
	}

	var currentMap map[string]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "#nameserver") {
				currentMap = *nameserver
			} else if strings.Contains(line, "#data-parity") {
				currentMap = *data_parity_servers
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && currentMap != nil {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			currentMap[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}
	return nil
}

func getFileinfoFromName(db *sql.DB, filename string) ([]fileinfo, error) {
	resRows, err := db.Query("SELECT file_name,file_chunk_name,iteration FROM chunk_metadata WHERE file_name=?", filename)

	if err != nil {
		return nil, err
	}
	defer resRows.Close()

	var infos []fileinfo

	for resRows.Next() {
		var info fileinfo
		err := resRows.Scan(&info.filename, &info.chunk_name, &info.iteration)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)

	}

	if err = resRows.Err(); err != nil {
		return nil, err
	}

	return infos, nil
}

func InitializeServersToAllReachable(data_parity_servers *map[string]string) map[string]bool {

	serverReachability := make(map[string]bool)

	for name, _ := range *data_parity_servers {
		serverReachability[name] = true
	}

	return serverReachability
}

type fileinfo struct {
	filename   string
	chunk_name string
	iteration  int32
}

type decoderServer struct {
	pb.UnimplementedDecoderServer
}

func (s *decoderServer) Decode(ctx context.Context, req *pb.FileRequest) (*pb.DecodeResponse, error) {
	// fileDir := ("encoded_output")

	//declare servers
	data_parity_servers := make(map[string]string)
	nameserver := make(map[string]string)

	ReadConfigFileAndInitalizeServers(&nameserver, &data_parity_servers, "server_config.txt")

	fileName := req.Filename

	//get chunk names
	db, err := sql.Open("mysql", "root:qwerty11@tcp("+nameserver["mysql"]+")/test")
	if err != nil {
		return &pb.DecodeResponse{
			Message: "Unable to connect to the sql database",
			Success: false,
		}, nil
	}

	fileChunkInfos, err := getFileinfoFromName(db, fileName)

	var reachable map[string]bool = InitializeServersToAllReachable(&data_parity_servers)

	for _, info := range fileChunkInfos {
		tcpconn_for_metadata, err := net.Dial("tcp", nameserver["mysql-tcp"])
		if err != nil {
			fmt.Printf("Unable to connect to nameserver to send metadata %s\n", err)
			return &pb.DecodeResponse{
				Message: "Unable to connect to metadata server to send data",
				Success: false,
			}, nil
		}

		_, err = fmt.Fprintf(tcpconn_for_metadata, "DOWNLOAD %s_metadata\n", info.chunk_name)
		if err != nil {
			fmt.Println("Unable to download complete metadata", err)
			return &pb.DecodeResponse{
				Message: "Unable to fetch metadata from nameserver",
				Success: false,
			}, nil
		}

		outFile, err := os.Create(info.chunk_name + "_metadata")
		if err != nil {
			return &pb.DecodeResponse{
				Message: "Unable to creat the file chunk in download contianer",
				Success: false,
			}, nil
		}

		defer outFile.Close()

		writer := bufio.NewWriter(outFile)
		bytesWritten, err := io.Copy(writer, tcpconn_for_metadata)

		writer.Flush()
		fmt.Printf("Wrote %d bytes to the file %s", bytesWritten, info.chunk_name)
	}

	var byte64filesList []string
	//read the metadata file and lines to download the files

	for _, info := range fileChunkInfos {
		chunkFile, err := os.Open(info.chunk_name + "_metadata")

		if err != nil {
			return &pb.DecodeResponse{
				Message: fmt.Sprintln("Unable to read chunk file ", err),
				Success: false,
			}, nil
		}

		chunkFileContent := bufio.NewScanner(chunkFile)

		//this dials every time to get data from all servers change this later to
		//single dial

		for chunkFileContent.Scan() {
			byte64fileName := chunkFileContent.Text()
			byte64filesList = append(byte64filesList, byte64fileName)
			fileSplit := strings.Split(byte64fileName, "_")
			designatedServer := data_parity_servers[fileSplit[len(fileSplit)-2]]

			if !reachable[fileSplit[len(fileSplit)-2]] {
				fmt.Printf("%s server is not reachable\n", fileSplit[len(fileSplit)-2])
				continue
			}
			conn, err := net.Dial("tcp", designatedServer)
			if err != nil {
				reachable[fileSplit[len(fileSplit)-2]] = false
				fmt.Println("Unable to contact server", err)
				continue
			}
			_, err = fmt.Fprintf(conn, "DOWNLOAD %s\n", byte64fileName)
			if err != nil {
				fmt.Println("Download requeste failed to  server "+designatedServer, err)
				continue
			}
			outFile, err := os.Create(byte64fileName)
			if err != nil {
				fmt.Println("Unable to create download file ", err)
				continue
			}

			writer := bufio.NewWriter(outFile)
			bytesWritten, err := io.Copy(writer, conn)
			if err != nil {
				fmt.Println("Unable to write downloaded data into file ", err)
				fmt.Printf("missed %d bytes of data \n", bytesWritten)
				continue
			}

			writer.Flush()
			outFile.Close()

		}

		cmd := exec.Command("./decoder", info.filename, info.chunk_name, "3145728", strconv.Itoa(int(info.iteration)))
		err = cmd.Run()
		if err != nil {

			return &pb.DecodeResponse{
				Message: fmt.Sprintf("Unable to decode chunk %s err: %s", info.chunk_name, err),
				Success: false,
			}, nil
		}

		// fmt.Println(outPut)
		chunkFile.Close()

		//delete files after conversion
		for _, byte64file := range byte64filesList {
			// fmt.Println("%d files removed",i+1)
			os.Remove(byte64file)
		}

	}

	//for each chunk get metadata
	// need to  create another recover.go that handles download and upload
	//
	//maintain server info
	//download metadata exactly 3 using goroutine and WaitKey

	return &pb.DecodeResponse{
		Message: "Successfully decoded data",
		Success: true,
	}, nil
}
