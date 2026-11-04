FROM golang:1.27-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server && \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /console ./cmd/console

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /server  .
COPY --from=builder /console .

EXPOSE 8000

CMD ["/app/server"]
