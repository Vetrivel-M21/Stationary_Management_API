FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .backend/cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/.env .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/seed ./seed

EXPOSE 8080

CMD ["./server"]
