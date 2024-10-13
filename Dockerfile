FROM golang:1.23.1-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/compiler-wrapper/

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/.env ./
COPY --from=builder main /bin/main
COPY --from=builder /build/config ./config

EXPOSE 8082

ENTRYPOINT ["/bin/main"]