FROM golang:1.21-alpine AS build

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o migrate ./cmd/migrate

FROM scratch

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /work/migrate /migrate

ENTRYPOINT ["/migrate"]
