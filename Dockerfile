
FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Change this line in Stage 1 of your Dockerfile:
RUN CGO_ENABLED=0 GOOS=linux go build -o httpserver ./cmd/httpserver

#stage 2


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/httpserver .

COPY assets ./assets

# Run the application
CMD ["./httpserver"]