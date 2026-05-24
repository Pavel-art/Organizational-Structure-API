FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd

FROM alpine:3.21

WORKDIR /app
COPY --from=builder /bin/app /app/app
COPY migrations /app/migrations

EXPOSE 8081
CMD ["/app/app"]
