FROM golang:1.25-bookworm AS builder

WORKDIR /server

ENV GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o server ./cmd/server
FROM scratch

WORKDIR /server

COPY --from=builder /server/server .

EXPOSE 8080

CMD ["./server"]
