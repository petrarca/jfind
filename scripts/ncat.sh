#!/bin/sh
PORT=${1:-8000}
if ! [ -x "$(command -v nc)" ]; then
	echo "netcat is missing. Please install it" >&2
	exit 1
fi
trap 'exit 0' INT

echo "Netcating calls on port $PORT"

while true; do nc -l -p $PORT <<EOF
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 44
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: POST, OPTIONS
Access-Control-Allow-Headers: Content-Type

{"result": "ok", "scan_id": "mock-scan-123"}
EOF
done
