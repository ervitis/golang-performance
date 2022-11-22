FROM golang:1.19-buster as builder

WORKDIR /application

COPY . .

RUN go mod download && \
    cd cmd/channels && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM debian:11.5-slim

WORKDIR /application
USER 1001

COPY --from=builder /application/cmd/channels/app /application
COPY --from=builder /application/data /application/data/
CMD ["/application/app"]