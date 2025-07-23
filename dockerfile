
FROM python:3.11-slim AS py-stage
WORKDIR /pybuild

RUN apt-get update && \
    apt-get install -y --no-install-recommends binutils && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt pyinstaller

COPY src/alphazero/ ./alphazero/
RUN pyinstaller --onefile --add-data "alphazero/best_policy_8_8_5.model:." \
                alphazero/game.py && \
    mv dist/game dist/wuziqi-ai

FROM golang:1.24-alpine AS go-stage
WORKDIR /build

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache \
        gcc musl-dev \
        mingw-w64-gcc \
        libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev glfw-dev\
        alsa-lib-dev pkgconfig

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 CC=gcc \
    GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o wuziqi-linux main.go

RUN CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
    GOOS=windows GOARCH=amd64 \
    go build -ldflags="-s -w" -o wuziqi-windows.exe main.go

FROM alpine:latest
WORKDIR /release

COPY --from=go-stage   /build/wuziqi-linux /build/wuziqi-windows.exe ./
COPY --from=py-stage   /pybuild/dist/wuziqi-ai ./
COPY src/assets        ./assets/
RUN ls -la
CMD ["/bin/sh"]