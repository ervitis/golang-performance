FROM golang:1.19-buster as builder

WORKDIR /application

COPY . .

RUN go mod download && \
    cd cmd/normal_process && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM scratch

WORKDIR /application
USER 1001

COPY --from=builder /application/cmd/normal_process/app /application
CMD ["/application/app"]