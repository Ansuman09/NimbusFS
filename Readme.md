## This is a file storage application that uses erasure coding

### Architecture

#### Nodes
- Naming node
- Data node
- Parity node
- Encoder node(Client)
- Decoder node(Client)


### Steps to run the application
#### Start the Encoder Decoder Node and Mount the decoded folder to your file path
docker build -t ubuntu-c .
docker run -d --name handler -p 50051:50051 -p 50052:50052 -v C:\Users\<username>\Downloads\:/app/decoder-rpc/server/decoded ubuntu-c 


- Execute the below commands to build the encoder, prepare the protofiles and starting the server.
- go inside the helper container and run "go run main.go server.go" inside decoder-rpc/server and test-rpc/server
- This is to avoid creating another container. 
- But ideally you can create one for decoder and encoder from the same image just provide specific port-mappings for each one.
- Move the schema.sql to the nameserver and build the database "test"
- docker cp schema.sql mysql-container:/

- Now Head into the storage-instance directory and start the containers. Does not matter what name you give to them make sure to note the IPs.
- build image: docker build -t server-image .
- docker run --name <server-node-1> -p 9443:9443 server-image (make sure to start K+M servers)

### Lets See it in action. 😉
We are going to run the servers in containers. But it can be extended to VMs or standalone Servers.

The default setup has M=2 and K=3 i.e 3 data storage nodes and 2 parity storage nodes. Which means we can afford upto M = 2 failures and still generate the data.

This is where we are at after starting all the severs. Lets do a "docker ps"
![Install Step 1](./images/all_servser.PNG)

### Implemented
- Handle uploads via chunking to microservices for encoding
- Naming server that stores metadata
- Download client that pulls data and decodes

### Features in development
- High Availability during Upload
- Conccurent uploads
- Conccurent Downloads
- Upload/Download retries
- Builder Node