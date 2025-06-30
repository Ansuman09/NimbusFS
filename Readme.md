## This is a file storage application that uses erasure coding

### Architecture

#### Nodes
- Naming node
- Data node
- Parity node
- Encoder node(Client)
- Decoder node(Client)


### Steps to run the application
#### Start the Encoder Decoder Node
docker build -t ubuntu-c .
docker run -p 50051:50051 -p 50052:50052 --name test-c ubuntu-c

- Execute the below commands
docker exec -it test-c sh -c "gcc /app/test-rpc/encoder.c -o /app/test-rpc/encoder -L/usr/lib -lisal"
install protobuf compiler
sudo apt update
sudo apt install -y protobuf-compiler
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative encoder.proto

### Implemented
- Handle uploads via chunking to microservices for encoding
- Naming server that stores metadata
- Download client that pulls data and decodes

The application still in building phase will be updaing the docs to get started

### Features in development
- High Availability during Upload
- Conccurent uploads
- Conccurent Downloads
- Upload/Download retries
- Builder Node