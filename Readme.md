# Distributed File Storage System with Erasure Coding

## 🧩 Overview

This is a distributed file storage system that uses **erasure coding** to store data with high availability and fault tolerance. It is capable of surviving up to `M` storage node failures using parity encoding and decoding.

---

## 🏗️ Architecture

### Components

- **Naming Node** – Tracks file metadata.
- **Data Node** – Stores data chunks.
- **Parity Node** – Stores parity chunks.
- **Encoder Node (Client)** – Uploads and encodes files.
- **Decoder Node (Client)** – Downloads and decodes files.

---

## 🚀 Getting Started

### Step 1: Build and Start the Handler (Encoder/Decoder Node)

```bash
docker build -t ubuntu-c .
docker run -d --name handler -p 50051:50051 -p 50052:50052 -v C:\Users\<username>\Downloads\:/app/decoder-rpc/server/decoded ubuntu-c
```

> Replace `<username>` with your local username.

---

### Step 2: Start gRPC Servers Inside Container

```bash
docker exec -it handler bash
cd decoder-rpc/server && go run main.go server.go
cd test-rpc/server && go run main.go server.go
```

---

### Step 3: Setup Naming Server Database

```bash
docker cp schema.sql mysql-container:/
```

Create the `test` database in MySQL using the copied schema.

---

### Step 4: Start Storage Nodes

```bash
docker build -t server-image .
docker run --name <server-node-1> -p 9443:9443 server-image
```

> Ensure all `K + M` servers are running and note their IPs.

---

### Step 5: Update Server Configs

Update IPs in:

- `handler/app/test-rpc/server/server_config.json`
- `handler/app/decoder-rpc/server/server_config.json`

> Docker Compose support coming soon.

---

## 📽️ Let’s See It in Action

Default: `K=3`, `M=2` (3 data nodes, 2 parity nodes).

### View Running Containers

```bash
docker ps
```

![Running servers](./images/all_servers.PNG)

---

### Uploading a File

```bash
cd test-rpc/client
go run main.go <file_name>
```

![Upload file](./images/before_decode.PNG)
![Upload action](./images/uploading_file.PNG)

---

### 🔄 Chunked File Transfer

#### Why Chunks?

- gRPC 4 MB limit → we use 3 MB chunks.

#### Transfer Flow:

- Split file → Send chunks → Encode → Distribute.

---

## 🛡️ High Availability Test

Shutdown:

- `data0`
- `parity0`

![Two nodes stopped](./images/stopped_storage_servers.PNG)

System can still decode the file.

---

### Downloading a File

```bash
cd decoder-rpc/client
go run main.go <file_name>
```

![Decode Output](./images/download_complete.PNG)
![Downloaded File](./images/after_decode.PNG)

---

## ✅ Implemented Features

- ✅ Chunked gRPC upload
- ✅ Naming server for metadata
- ✅ Redundant storage (K+M)
- ✅ Fault-tolerant downloads

---

## 🔧 In Development

- [ ] Docker Compose support
- [ ] Concurrent uploads/downloads
- [ ] Retry logic
- [ ] Builder Node

---

## 🤝 Contributing

Feel free to fork and submit PRs!
