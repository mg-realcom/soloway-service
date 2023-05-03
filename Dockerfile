FROM golang:alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o soloway-service /build/cmd/server/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /build/soloway-service ./
RUN chmod +x soloway-service
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
ENTRYPOINT ["./soloway-service", "--env"]
