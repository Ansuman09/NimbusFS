## This is a file storage application that uses erasure coding

### Architecture

#### Nodes
- Naming node
- Data node
- Parity node
- Builder node
- Encoder node(Client)
- Decoder node(Client)



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