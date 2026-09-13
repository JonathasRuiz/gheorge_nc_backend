# syntax=docker/dockerfile:1

FROM docker.io/golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /bin/server /app/server
RUN mkdir -p /app/data && chown -R app:app /app
USER app
ENV DB_PATH=/app/data/data.db \
    PORT=80
EXPOSE 80
VOLUME ["/app/data"]
ENTRYPOINT ["/app/server"]
