#!/bin/bash
echo "Container is running..."
tail -f /dev/null

/usr/local/go/bin/go /app/decoder-rpc/server/main.go /app/decoder-rpc/server/main.go 