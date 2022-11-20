FROM golang:1.19.3 AS builder
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app/
RUN go build -o exporter main.go

FROM gcr.io/distroless/static-debian11:nonroot AS runner

COPY --chown=nonroot:nonroot --from=builder /app/exporter /exporter
ENTRYPOINT [ "/exporter" ]

