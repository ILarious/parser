FROM golang:1.26-alpine AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/parser ./cmd/app

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /bin/parser /app/parser

CMD ["/app/parser"]
