# Start from the official Golang base image
FROM golang:1.22-alpine as builder


WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scalingo_assesment .


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/


COPY --from=builder /app/scalingo_assesment .


COPY --from=builder /app/config.yaml .

EXPOSE 8095


CMD ["./scalingo_assesment"]