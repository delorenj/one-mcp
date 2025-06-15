#!/bin/bash
set -e
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
fi
echo "Building frontend..."
cd frontend
npm run build
cd ..

echo "Building backend..."
go build -o one-mcp main.go

echo "Build complete"
./one-mcp