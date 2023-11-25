FROM golang:1.21.4

WORKDIR /app

COPY go-motive-webhook/ ./

# Download and install any required dependencies
RUN go mod download

EXPOSE 8080
EXPOSE 80

# Run
CMD ["go", "run", "."]