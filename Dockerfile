# 在仓库根目录构建：docker build -t chuyaji-api .
FROM golang:1.22-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/chuyaji-api ./services/api/cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/chuyaji-api /usr/local/bin/chuyaji-api
ENV CHUYAJI_HTTP_ADDR=:8282
ENV CHUYAJI_DB_PATH=/data/chuyaji.db
ENV CHUYAJI_UPLOAD_DIR=/data/uploads
VOLUME ["/data"]
EXPOSE 8282
ENTRYPOINT ["/usr/local/bin/chuyaji-api"]
