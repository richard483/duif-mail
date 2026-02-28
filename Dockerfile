FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gomail ./cmd/api

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/gomail /app/gomail

USER app
EXPOSE 50051

ENTRYPOINT ["./gomail"]
