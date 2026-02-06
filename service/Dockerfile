FROM golang:1.24-bullseye AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY service/go.mod service/go.sum ./service/

WORKDIR /src/service
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY service ./service

WORKDIR /src/service
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out/server /server
EXPOSE 9090
ENTRYPOINT ["/server"]