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

	fileName := req.Filename

	//get chunk names
	db, err := sql.Open("mysql", "root:qwerty11@tcp(172.17.0.8:3306)/test")
	if err != nil {
		return &pb.DecodeResponse{
			Message: "Unable to connect to the sql database",
			Success: false,
		}, nil
	}

	fileChunkInfos, err := getFileinfoFromName(db, fileName)

	//declare servers
	data_parity_servers := make(map[string]string)

	data_parity_servers["data0"] = "172.17.0.2:9443"
	data_parity_servers["data1"] = "172.17.0.3:9443"
	data_parity_servers["data2"] = "172.17.0.4:9443"
	data_parity_servers["parity0"] = "172.17.0.5:9443"
	data_parity_servers["parity1"] = "172.17.0.6:9443"

	// server_connections := make(map[string]net.Conn)

	// for server, ip_port := range data_parity_servers {
	// 	// var ip_port string = data_parity_servers[server]
	// 	conn, err := net.Dial("tcp", ip_port)
	// 	if err != nil {
	// 		return &pb.DecodeResponse{
	// 			Message: fmt.Sprintf("Unable to connect to server %s\n %s ", server, err),
	// 			Success: false,
	// 		}, nil
	// 	}

	// 	server_connections[server] = conn
	// }

	//get chunk metadata

	//get metadata
	var reachable map[string]bool = make(map[string]bool)
	reachable["data0"] = true
	reachable["data1"] = true
	reachable["data2"] = true
	reachable["parity0"] = true
	reachable["parity1"] = true
	for _, info := range fileChunkInfos {
		tcpconn_for_metadata, err := net.Dial("tcp", "172.17.0.8:9443")
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
