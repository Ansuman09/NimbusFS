#Docker file for receivers all will use  the same image
FROM golang:1.24.3
WORKDIR /app

COPY go.mod ./

#create the storange-instance dir
RUN mkdir storage-instance

COPY . .

EXPOSE 9443

# Run the compiled binary
CMD ["go","run","receiver.go"]
