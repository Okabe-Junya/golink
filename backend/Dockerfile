FROM golang:1.24.3-bookworm as builder

WORKDIR /app
COPY go.mod go.sum ./
ENV CGO_ENABLED=0
RUN go mod download
COPY . .

RUN go build -o bin/server cmd/server/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/bin/server /app/server

EXPOSE 8080
CMD ["./server"]
