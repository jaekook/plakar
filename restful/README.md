# Plakar RESTful API Server

A complete RESTful API server implementation for Plakar backup solution using Go Echo framework and AWS S3 storage.

## Features

- **Complete API Coverage**: Implements all endpoints from the OpenAPI specification
- **AWS S3 Integration**: Native support for AWS S3 as storage backend
- **Authentication**: Bearer token authentication with configurable tokens
- **Validation**: Request validation using go-playground/validator
- **Error Handling**: Structured error responses with proper HTTP status codes
- **CORS Support**: Cross-origin resource sharing for web applications
- **Health Checks**: Built-in health check endpoint
- **Graceful Shutdown**: Proper server shutdown handling

## Quick Start

### Prerequisites

- Go 1.23.4 or higher
- AWS account with S3 bucket
- AWS credentials configured

### Installation

1. Clone the repository:
```bash
git clone https://github.com/PlakarKorp/plakar.git
cd plakar/restful
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set environment variables:
```bash
export AWS_REGION=us-east-1
export AWS_S3_BUCKET=your-backup-bucket
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AUTH_TOKEN=your-secure-token
export PORT=8080
```

4. Run the server:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Configuration

The server is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `HOST` | Server host | `0.0.0.0` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `AWS_S3_BUCKET` | S3 bucket name | **Required** |
| `AWS_ACCESS_KEY_ID` | AWS access key | From AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | From AWS credentials |
| `AUTH_TOKEN` | Bearer token for authentication | Empty (no auth) |

## API Endpoints

### System
- `GET /health` - Health check
- `GET /api/info` - API information

### Repository Management
- `POST /api/repository/create` - Create repository
- `GET /api/repository/info` - Get repository info
- `GET /api/repository/snapshots` - List snapshots
- `POST /api/repository/maintenance` - Run maintenance
- `POST /api/repository/prune` - Prune repository
- `POST /api/repository/sync` - Sync repositories

### Snapshot Operations
- `POST /api/snapshots/create` - Create backup
- `GET /api/snapshot/{id}` - Get snapshot details
- `POST /api/snapshot/{id}/restore` - Restore snapshot
- `POST /api/snapshot/{id}/check` - Check integrity
- `GET /api/snapshot/{id}/diff/{target}` - Compare snapshots
- `POST /api/snapshot/{id}/mount` - Mount snapshot
- `DELETE /api/snapshots/remove` - Remove snapshots

### File Operations
- `GET /api/snapshot/reader/*` - Read file content
- `GET /api/files/cat/*` - Get file with processing
- `GET /api/files/digest/*` - Get file digest
- `POST /api/snapshot/reader-sign-url/*` - Create signed URL

### VFS Operations
- `GET /api/snapshot/{id}/vfs/*` - Browse VFS
- `GET /api/snapshot/{id}/vfs/children/*` - List directory
- `GET /api/snapshot/{id}/vfs/search/*` - Search files
- `POST /api/snapshot/{id}/vfs/downloader/*` - Create download package

### Search
- `GET /api/search/locate` - Locate files across snapshots
- `GET /api/repository/locate-pathname` - Find file timeline

### Authentication
- `POST /api/authentication/login/github` - GitHub OAuth
- `POST /api/authentication/login/email` - Email login
- `POST /api/authentication/logout` - Logout

### Scheduler
- `POST /api/scheduler/start` - Start scheduler
- `POST /api/scheduler/stop` - Stop scheduler
- `GET /api/scheduler/status` - Get scheduler status

## Usage Examples

### Create Repository
```bash
curl -X POST http://localhost:8080/api/repository/create \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "location": "s3://your-bucket/repo",
    "passphrase": "secure-passphrase",
    "hashing": "SHA256"
  }'
```

### Create Backup
```bash
curl -X POST http://localhost:8080/api/snapshots/create \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "/path/to/backup",
    "tags": ["important", "daily"],
    "excludes": ["*.tmp", "*.log"],
    "check": true
  }'
```

### List Snapshots
```bash
curl -H "Authorization: Bearer your-token" \
  "http://localhost:8080/api/repository/snapshots?limit=10&sort=-Timestamp"
```

### Restore Files
```bash
curl -X POST http://localhost:8080/api/snapshot/abc123.../restore \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "destination": "/tmp/restore",
    "paths": ["/important/file.txt"]
  }'
```

## Architecture

### Directory Structure
```
restful/
├── main.go              # Server entry point
├── config/              # Configuration management
├── handlers/            # HTTP request handlers
├── middleware/          # Custom middleware
├── models/              # Data models and structs
├── storage/             # Storage interface and S3 implementation
└── utils/               # Utilities and helpers
```

### Key Components

1. **Echo Framework**: High-performance HTTP router and middleware
2. **AWS S3 Storage**: Cloud storage backend with native AWS SDK
3. **Plakar Integration**: Direct integration with Plakar core libraries
4. **Validation**: Request validation using struct tags
5. **Authentication**: Bearer token middleware with constant-time comparison
6. **Error Handling**: Structured error responses with proper HTTP codes

## Development

### Adding New Endpoints

1. Define the model in `models/models.go`
2. Add the storage interface method in `storage/storage.go`
3. Implement the method in `storage/s3.go`
4. Create the handler in appropriate `handlers/*.go` file
5. Register the route in `main.go`

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test ./handlers -v
```

### Building

```bash
# Build binary
go build -o plakar-api main.go

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o plakar-api-linux main.go
GOOS=windows GOARCH=amd64 go build -o plakar-api-windows.exe main.go
```

## Docker Support

### Dockerfile
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o plakar-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/plakar-api .
EXPOSE 8080
CMD ["./plakar-api"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  plakar-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - AWS_REGION=us-east-1
      - AWS_S3_BUCKET=your-bucket
      - AUTH_TOKEN=your-token
    volumes:
      - ~/.aws:/root/.aws:ro
```

## Security Considerations

1. **Authentication**: Always use strong bearer tokens in production
2. **HTTPS**: Use HTTPS in production environments
3. **CORS**: Configure CORS appropriately for your use case
4. **AWS Credentials**: Use IAM roles instead of access keys when possible
5. **Input Validation**: All inputs are validated using struct tags
6. **Error Handling**: Errors don't leak sensitive information

## Performance

- **Concurrent Operations**: Configurable concurrency for backup/restore operations
- **Streaming**: Large file operations use streaming to minimize memory usage
- **Connection Pooling**: AWS SDK handles connection pooling automatically
- **Graceful Shutdown**: Proper cleanup of resources on shutdown

## Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics
The server provides basic operational metrics through the health endpoint and can be extended with Prometheus metrics.

## Troubleshooting

### Common Issues

1. **AWS Credentials**: Ensure AWS credentials are properly configured
2. **S3 Permissions**: Verify S3 bucket permissions for read/write operations
3. **Network**: Check firewall and security group settings
4. **Memory**: Monitor memory usage for large backup operations

### Logs
The server uses structured logging with different levels:
- `INFO`: General operational information
- `WARN`: Warning conditions
- `ERROR`: Error conditions
- `DEBUG`: Detailed debugging information

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the same license as the main Plakar project.

## Support

For support and questions:
- GitHub Issues: [https://github.com/PlakarKorp/plakar/issues](https://github.com/PlakarKorp/plakar/issues)
- Discord: [https://discord.gg/uqdP9Wfzx3](https://discord.gg/uqdP9Wfzx3)
- Documentation: [https://docs.plakar.io](https://docs.plakar.io)