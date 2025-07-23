#!/bin/bash
set -e
echo "🐳 Building multi-platform release …"
docker build -t wuziqi-builder .
rm -rf release && mkdir release
docker run --rm -v "$PWD/release:/out" wuziqi-builder sh -c "cp -r /release/* /out"
echo "Done! Check ./release/"