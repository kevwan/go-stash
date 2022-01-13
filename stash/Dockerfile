FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build/zero

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
COPY stash/etc /app/etc
RUN GOARCH=amd64 go build -ldflags="-s -w" -o /app/stash stash/stash.go
RUN apk update --no-cache && apk add upx && upx /app/stash


FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/stash /app/stash
COPY --from=builder /app/etc /app/etc

CMD ["./stash", "-f", "etc/config.yaml"]
