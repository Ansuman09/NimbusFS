# Use official Ubuntu base image
FROM ubuntu:22.04

# Install OS dependencies
FROM ubuntu:22.04

# Install OS dependencies and build tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    autoconf \
    automake \
    libtool \
    pkg-config \
    git \
    curl \
    ca-certificates \
    wget \
    unzip \
    nasm \
    && git clone https://github.com/intel/isa-l.git /tmp/isa-l \
    && cd /tmp/isa-l \
    && ./autogen.sh \
    && ./configure \
    && make \
    && make install \
    && ldconfig \
    && rm -rf /tmp/isa-l \
    && rm -rf /var/lib/apt/lists/*


# --- Install Go ---
ENV GO_VERSION=1.21.10
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Set Go env
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

# Create app directory
WORKDIR /app

# Install gRPC Go plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Set PATH for protoc-gen plugins
ENV PATH=$PATH:/go/bin

# Copy application files 
COPY test-rpc/server/* /app/test-rpc/server/
COPY test-rpc/proto/* /app/test-rpc/proto/
COPY test-rpc/go* /app/decoder-rpc/

COPY decoder-rpc/server/* /app/decoder-rpc/server/
COPY decoder-rpc/proto/* /app/decoder-rpc/proto/
COPY decoder-rpc/go* /app/decoder-rpc/

#Copy over the coder and decoder file to the server locations
COPY encode_decode/decoder.c /app/decoder-rpc/server/
COPY encode_decode/encoder.c /app/test-rpc/server/
COPY encode_decode/config.txt /app/decoder-rpc/server/ 
COPY encode_decode/config.txt /app/test-rpc/server/
COPY run.sh .


#BUILD the encoder and decoder
RUN gcc /app/test-rpc/server/encoder.c -o /app/test-rpc/server/encoder -L/usr/lib -lisal
RUN gcc /app/decoder-rpc/server/decoder.c -o  /app/decoder-rpc/server/decoder -lisal

# RUN go build -o server ./server
RUN chmod +x run.sh
# Default CMD 
CMD ["./run.sh"]

# CMD ["/bin/sh", "-c", "go run /app/decoder-rpc/server/main.go & go run /app/test-rpc/server/main.go && wait"]

#start the service


