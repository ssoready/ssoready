FROM golang:1.21-alpine AS build

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-w -s" -o with-secrets ./cmd/with-secrets
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-w -s" -o api ./cmd/api

FROM scratch

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /work/with-secrets /with-secrets
COPY --from=build /work/api /api

EXPOSE 8081
ENTRYPOINT ["/with-secrets", "api", "--", "/api"]
