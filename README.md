# Project Name

This is a Go-based web application that provides  marketplace API services.

## Main Features

- Marketplace Functions
- Configuration Management
- Tracing System

## Project Structure
```
├── api
│   ├── marketplace
│   └── admin
├── cmd
├── config
│   ├── config.go
│   └── config.toml.example
├── develop
├── internal
│   ├── db      Database instances
│   ├── entity  Request and response entities
│   ├── errno   Error codes
│   └── services
│       ├── marketplace
├── pkg
├── main.go
└── README.md
```

## Installation and Running

1. Clone repository:
   ```
   git clone [repository URL]
   cd [project directory]
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Configure application:
   Copy `config/config.yaml.example` to `config/config.yaml` and modify settings as needed.

4. Run application:
   ```
   go run main.go
   ```

## Configuration

Configuration file is located at `config/config.yaml`. Main configuration items include:

- Database connection
- Server settings
- Logging configuration
- Other application-specific settings

For detailed configuration information, please refer to `config/config.go`.

## Development and Debugging

For local development, you can use Docker Compose to start MySQL and Redis:

```
cd develop

./develop.sh
```

## API Routes

API routes are defined in the `api` directory.

## Tracing System

The project uses a custom tracing system implemented in `pkg/trace/trace.go`, using Jaeger for tracing.
