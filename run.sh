#!/bin/bash

# Set error handling
set -e
echo "Building frontend..."
cd frontend && npm install && npm run build && cd ..

# Load .env environment variables
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
fi

# Ensure PATH includes /usr/local/bin
export PATH=$PATH:/usr/local/bin

# Set port number from environment variable, default to 3000
PORT=${PORT:-3000}

# Clean up existing processes
echo "Cleaning up existing processes..."
# Kill processes using backend port
lsof -ti:$PORT | xargs kill -9 2>/dev/null || echo "No existing backend processes found on port $PORT"

# Store process IDs
BACKEND_PID=""
# Cleanup function
cleanup() {
    echo -e "\nShutting down servers..."
    
    # Clean up backend process
    if [ ! -z "$BACKEND_PID" ] && ps -p $BACKEND_PID > /dev/null; then
        echo "Killing backend process $BACKEND_PID"
        kill -TERM $BACKEND_PID 2>/dev/null || kill -9 $BACKEND_PID 2>/dev/null
    fi

    # Backend port
    pid=$(lsof -ti :$PORT 2>/dev/null)
    if [ ! -z "$pid" ]; then
        echo "Killing lingering backend process on port $PORT (PID: $pid)"
        kill -9 $pid 2>/dev/null || true
    fi
    exit 0
}

# Set signal handling
trap cleanup INT TERM

# Build backend service
echo "Building backend service..."
go build -o one-mcp .

# Start backend service
echo "Starting backend service..."
nohup ./one-mcp > backend.log 2>&1 &
BACKEND_PID=$!
echo -e "\nServer started:"
echo "- Port: $PORT (PID: $BACKEND_PID)"
echo "Press Ctrl+C to stop all servers."

# Wait for all processes
wait 