# DDD Parser API

A simple REST API service that parses DDD (Digital Tachograph) files and returns the parsed data as JSON.

## Features

- **Auto-detect** file type (Card or VU)
- **RESTful API** with JSON responses
- **File upload** or raw binary data support
- **CORS enabled** for web client access
- **Health check** endpoint

## Usage

### Option 1: Run directly

```bash
go run cmd/dddapi/main.go
```

Or build and run:

```bash
go build -o dddapi cmd/dddapi/main.go
./dddapi
```

### Option 2: Docker

Build the Docker image:

```bash
# From the project root
docker build -f Dockerfile.dddapi -t dddapi:latest .
```

Run the container:

```bash
docker run -p 8080:8080 dddapi:latest
```

Or use docker-compose:

```bash
cd cmd/dddapi
docker-compose up
```

### Configuration

- `-port`: Port to listen on (default: 8080)
- `-addr`: Address to bind to (default: "" = all interfaces)

Example:
```bash
./dddapi -port 3000 -addr 127.0.0.1
```

## API Endpoints

### GET /

API information and available endpoints.

**Response:**
```json
{
  "service": "DDD Parser API",
  "version": "1.0.0",
  "endpoints": { ... }
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "service": "ddd-parser-api"
}
```

### POST /parse

Auto-detect file type (Card or VU) and parse.

**Request:**
- Option 1: Multipart form with key `file`
- Option 2: Raw binary data in request body

**Response:**
```json
{
  "type": "card" | "vu",
  "verified": true | false,
  "data": { ... }  // Card or Vu struct as JSON
}
```

### POST /parse/card

Parse data as Card (TLV format).

**Request:** Same as `/parse`

**Response:**
```json
{
  "type": "card",
  "verified": true | false,
  "data": { ... }  // Card struct as JSON
}
```

### POST /parse/vu

Parse data as VU (TV format).

**Request:** Same as `/parse`

**Response:**
```json
{
  "type": "vu",
  "verified": true | false,
  "data": { ... }  // Vu struct as JSON
}
```

## Examples

### Using curl with file upload

```bash
curl -X POST http://localhost:8080/parse \
  -F "file=@testfile.DDD" \
  -H "Content-Type: multipart/form-data"
```

### Using curl with raw binary data

```bash
curl -X POST http://localhost:8080/parse \
  --data-binary @testfile.DDD \
  -H "Content-Type: application/octet-stream"
```

### Parse as Card

```bash
curl -X POST http://localhost:8080/parse/card \
  -F "file=@drivercard.ddd"
```

### Parse as VU

```bash
curl -X POST http://localhost:8080/parse/vu \
  -F "file=@vu.ddd"
```

## Error Responses

All errors return JSON with an `error` field:

```json
{
  "error": "Failed to parse as Card",
  "details": "..."
}
```

## CORS

CORS is enabled by default, allowing requests from any origin. This is useful for web-based clients.

