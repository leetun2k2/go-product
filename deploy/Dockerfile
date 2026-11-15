# syntax=docker/dockerfile:1

FROM golang:1.27.1 AS dependencies

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM dependencies AS build

COPY cmd ./cmd
COPY config ./config
COPY internal ./internal

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/go-product \
    ./cmd/server

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/go-product /go-product

USER 65532:65532

EXPOSE 4040 5050

ENTRYPOINT ["/go-product"]
