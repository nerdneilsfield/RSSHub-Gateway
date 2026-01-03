FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/rsshub-gateway ./

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

WORKDIR /app
COPY --from=build /out/rsshub-gateway /app/rsshub-gateway
COPY config.example.yaml /app/config.example.yaml

EXPOSE 8080
ENTRYPOINT ["/app/rsshub-gateway", "serve", "-c", "/app/config.example.yaml"]
