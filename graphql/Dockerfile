FROM golang:1.24-bullseye AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY graphql/go.mod graphql/go.sum ./graphql/

WORKDIR /src/graphql
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY graphql ./graphql

WORKDIR /src/graphql
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/graphql ./cmd/graphql

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out/graphql /graphql
EXPOSE 8080
ENTRYPOINT ["/graphql"]