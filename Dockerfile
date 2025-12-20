FROM golang:1.25-alpine AS builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /go/bin/ffmpeg-api \
        cmd/main.go

FROM jrottenberg/ffmpeg:8.0-scratch AS ffmpeg

FROM scratch

COPY --from=ffmpeg /bin /bin
COPY --from=ffmpeg /lib /lib
COPY --from=ffmpeg /share /share
COPY --from=ffmpeg /usr/share/fonts /usr/share/fonts
COPY --from=ffmpeg /usr/share/fontconfig /usr/share/fontconfig

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/ffmpeg-api /usr/local/bin/ffmpeg-api

EXPOSE 8080

ENTRYPOINT ["ffmpeg-api"]
