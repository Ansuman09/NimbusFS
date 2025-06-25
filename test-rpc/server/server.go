package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	pb "test-rpc/proto"

	_ "github.com/go-sql-driver/mysql"
)

type encoderServer struct {
	pb.UnimplementedEncoderServer
}

type fileinfo struct {
	filename string
}

type chunk_metadata struct {
	filename   string
	chunk_name string
	iteration  int32
}

func createFileinfoEntry(db *sql.DB, filedata fileinfo) (int64, error) {
	res, err := db.Exec("INSERT INTO fileinfo (filename) VALUES (?)", filedata.filename)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func getfilePresentByName(db *sql.DB, filename string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM fileinfo WHERE filename = ?)`
	err := db.QueryRow(query, filename).Scan(&exists)
	return exists, err
}

func createChunkFileinfoEntry(db *sql.DB, chunkdata chunk_metadata) (int64, error) {
	res, err := db.Exec("INSERT INTO chunk_metadata (file_chunk_name,iteration,file_name) VALUES (?,?,?)", chunkdata.chunk_name, chunkdata.iteration, chunkdata.filename)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *encoderServer) Encode(ctx context.Context, req *pb.FileRequest) (*pb.EncodeResponse, error) {
	inputDir := "uploads"
	outputDir := "encoded_output"

	//get full filename and iteration
	// fileNameWithFormat := req.Filefullname
	var iteration int32 = req.Iteration
	// creates directory with permissions 777
	os.MkdirAll(inputDir, os.ModePerm)
	os.MkdirAll(outputDir, os.ModePerm)

	// filepath:=
	fp := filepath.Join(inputDir, req.Filename)

	//file path , data in bytes and permission
	err := os.WriteFile(fp, req.FileData, 0644)

	if err != nil {
		return &pb.EncodeResponse{
			Success: false,
			Message: "Failed to save file: " + err.Error(),
		}, nil
	}

	fmt.Printf("cmd params %s and %s %s \n", fp, outputDir, req.Filename)
	cmd := exec.Command("./encoder", fp, outputDir, req.Filename)

	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return &pb.EncodeResponse{
	// 		Success: false,
	// 		Message: fmt.Sprintf("Encoding failed: %s\n%s", err, output),
	// 	}, nil
	// }

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &pb.EncodeResponse{
			Success: false,
			Message: fmt.Sprintf("Encoding failed: metadata creation failed: %s\n", err),
		}, nil
	}

	if err := cmd.Start(); err != nil {
		return &pb.EncodeResponse{
			Success: false,
			Message: fmt.Sprintf("Encoding failed: metadata creation failed: %s\n", err),
		}, nil
	}

	//connect to the sql database on the nameserver to send the chunkfile update

	db, err := sql.Open("mysql", "root:qwerty11@tcp(172.17.0.8:3306)/test")
	if err != nil {
		return &pb.EncodeResponse{
			Message: "Unable to connect to the sql database",
			Success: false,
		}, nil
	}

	//check if filename exists if not create it ------------------------- update this according ly
	filedetails := fileinfo{filename: req.Filefullname}
	exists, err := getfilePresentByName(db, filedetails.filename)
	if err != nil {
		log.Fatal("Error checking for file:", err)
	}

	if !exists {
		id, err := createFileinfoEntry(db, filedetails)
		if err != nil {
			log.Fatal("Error inserting fileinfo:", err)
		}
		fmt.Println("Inserted ID:", id)
	} else {
		fmt.Printf("\nFile is present in db, updating %s\n", filedetails.filename)
	}

	//mark if all files were send successfully
	var successFulFiletransfer int = 1
	scanner := bufio.NewScanner(stdout)

	chunkdata := chunk_metadata{chunk_name: req.Filename, iteration: iteration, filename: req.Filefullname}

	metadataFilePath := filepath.Join(outputDir, chunkdata.chunk_name+"_metadata")

	//rename later
	data_parity_servers := make(map[string]string)
	data_parity_servers["data0"] = "172.17.0.2:9443"
	data_parity_servers["data1"] = "172.17.0.3:9443"
	data_parity_servers["data2"] = "172.17.0.4:9443"
	data_parity_servers["parity0"] = "172.17.0.5:9443"
	data_parity_servers["parity1"] = "172.17.0.6:9443"

	server_connections := make(map[string]net.Conn)

	for server, ip_port := range data_parity_servers {
		// var ip_port string = data_parity_servers[server]
		conn, err := net.Dial("tcp", ip_port)
		if err != nil {
			return &pb.EncodeResponse{
				Message: fmt.Sprintf("Unable to connect to server %s\n %s ", server, err),
				Success: false,
			}, nil
		}

		// add log

		server_connections[server] = conn
	}

	chunkMetadata_file_name_for_nameserver := chunkdata.chunk_name + "_metadata"

	tcpconn, err := net.Dial("tcp", "172.17.0.8:9443")
	if err != nil {
		fmt.Printf("Unable to connect to nameserver to send metadata %s\n", err)
		return &pb.EncodeResponse{
			Message: "Unable to connect to metadata server to send data",
			Success: false,
		}, nil
	}

	defer tcpconn.Close()
	_, err = fmt.Fprintf(tcpconn, "UPLOAD %s\n", chunkMetadata_file_name_for_nameserver)
	if err != nil {
		fmt.Println("Error sending filename:", err)
		return &pb.EncodeResponse{
			Message: "Unable to send data to metadata server to send data",
			Success: false,
		}, nil
	}

	metadataFile, err := os.OpenFile(metadataFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	for scanner.Scan() {
		// line := scanner.Text()
		name_of_file_to_send := scanner.Text()
		fmt.Printf("%s\n", name_of_file_to_send)
		parts_of_out := strings.Split(name_of_file_to_send, "_")

		//does the file belong to which of the data or parity nodes
		desgnated_server := data_parity_servers[parts_of_out[len(parts_of_out)-2]]

		//work to do:
		// create the 5 servers note ip and update
		// send metadata file name and metadata content over this tcp.
		// send files over tcp
		server_tcp_conn, err := net.Dial("tcp", desgnated_server)
		if err != nil {
			fmt.Printf("Unable to send data for file %s %s", name_of_file_to_send, err)
			break
		}
		_, err = fmt.Fprintf(server_tcp_conn, "UPLOAD %s \n", name_of_file_to_send)
		if err != nil {
			return &pb.EncodeResponse{
				Message: "unable to create the file for 64 byte chunnk\n",
				Success: false,
			}, nil
		}

		path_of_file_to_send := filepath.Join(outputDir, name_of_file_to_send)
		byte64_chunk_file, err := os.Open(path_of_file_to_send)

		if err != nil {
			fmt.Println(path_of_file_to_send)
			return &pb.EncodeResponse{
				Message: "Unable to open the encoded 64 byte file\n",
				Success: false,
			}, nil
		}
		bytesSentOfChunk, err := io.Copy(server_tcp_conn, byte64_chunk_file)
		if err != nil {
			return &pb.EncodeResponse{
				Message: fmt.Sprintf("Error sending the encoded 64 byte chunk: %s", err),
				Success: false,
			}, nil
		}

		fmt.Printf("Successfully sent %d bytes of data for the file %s", bytesSentOfChunk, name_of_file_to_send)

		// chunkMetadata_file_name_for_nameserver
		// fmt.Printf("Closed server %s", desgnated_server)
		byte64_chunk_file.Close()
		successFulFiletransfer = successFulFiletransfer * 1
		_, err = metadataFile.WriteString(name_of_file_to_send + "\n")
		if err != nil {
			log.Fatal(err)
		}
		///if error while sending successFulFiletransfer * 0
	}
	metadataFile.Close()

	metadataFileR, err := os.Open(metadataFilePath)

	if err != nil {
		return &pb.EncodeResponse{
			Message: "Unable to open file after writing data",
			Success: false,
		}, nil
	}

	bytesSent, err := io.Copy(tcpconn, metadataFileR)
	if err != nil {
		fmt.Println("Error sending file:", err)
		return &pb.EncodeResponse{
			Message: "Error while sending metadata file to nameserver",
			Success: false,
		}, nil
	}

	fmt.Printf("File '%s' sent successfully (%d bytes)\n", chunkMetadata_file_name_for_nameserver, bytesSent)

	server_connections["data0"].Close()
	server_connections["data1"].Close()
	server_connections["data2"].Close()
	server_connections["parity0"].Close()
	server_connections["parity1"].Close()
	defer metadataFile.Close()

	fmt.Printf("chunk name now %s\n", chunkdata.chunk_name)
	_, err = createChunkFileinfoEntry(db, chunkdata)
	if err != nil {
		log.Fatal("Error inserting chunk:", err)
	}

	cmd.Wait()

	return &pb.EncodeResponse{
		Success: true,
		Message: "Encoding completed successfully",
	}, nil

}
