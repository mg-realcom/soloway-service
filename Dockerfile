FROM golang:alpine AS builder
RUN apk add --no-cache git
RUN --mount=type=secret,id=github_token,required \
  git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/".insteadOf "https://github.com/"
WORKDIR /build
COPY . .
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o soloway-service /build/cmd/server/

FROM alpine
WORKDIR /app
COPY --from=builder /build/soloway-service ./
RUN chmod +x soloway-service
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
ENTRYPOINT ["./soloway-service"]