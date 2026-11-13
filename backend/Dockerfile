# 1. Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/seeder ./cmd/seeder

# 2. Final Stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/seeder /app/seeder
COPY --from=builder /app/migrations /app/migrations

ENV TZ=Asia/Yekaterinburg
EXPOSE 8080

CMD ["/app/server"]
