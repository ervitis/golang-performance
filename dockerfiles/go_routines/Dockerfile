FROM golang:1.19-buster as builder

WORKDIR /application

COPY . .

RUN go mod download && \
    cd cmd/go_routines && \
    CGO_ENABLED=0 GOOS=linux go build -o app .

FROM debian:11.5-slim

WORKDIR /application
USER 1001

COPY --from=builder /application/cmd/go_routines/app /application
COPY --from=builder /application/data /application/data/
CMD ["/application/app"]