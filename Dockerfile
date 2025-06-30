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

# Copy application files (update this as needed)
COPY test-rpc .
COPY decoder-rpc .
COPY encode_decode/decoder.c decoder-rpc/server/
COPY encode_decode/encoder.c test-rpc/server/
COPY run.sh .

# Build step (optional: can do at runtime too)
# RUN go build -o server ./server
RUN chmod +x run.sh
# Default CMD (can be changed if needed)
CMD ["./run.sh"]

