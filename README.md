# PON Watcher

PON Watcher is an intelligent optical network monitoring application built in Go that provides real-time monitoring of Optical Network Units (ONUs) in PON (Passive Optical Network) infrastructures.

## Features

- **Real-time ONU Monitoring**: Monitor optical network units across PON networks
- **Intelligent Categorization**: Automatically categorizes ONUs as online/offline based on power levels and voltage status
- **TL1 Protocol Support**: Communicates with UNM (Universal Network Manager) servers via TL1 protocol
- **RESTful API**: HTTP API for querying ONU information and statistics
- **Concurrent Processing**: High-performance concurrent data fetching with configurable worker pools
- **Smart Logging**: Advanced logging with colored output, benchmarking, and structured logging
- **Configuration Management**: Flexible configuration with YAML files and environment variable support
- **Graceful Shutdown**: Proper cleanup and graceful shutdown handling
- **Dependency Injection**: Clean architecture using Uber FX framework

## Architecture

The application follows a clean architecture pattern with the following layers:

- **HTTP Layer**: Gin-based REST API handlers
- **Use Case Layer**: Business logic for fetching and processing ONU data
- **Repository Layer**: Data access abstraction (in-memory OLT repository)
- **Adapter Layer**: External service integrations (UNM server communication)
- **Infrastructure**: TL1 transport, logging, configuration management

## Prerequisites

- Go 1.21 or higher
- Access to a UNM server supporting TL1 protocol
- Network connectivity to target OLT devices

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/raykavin/ponwatcher-go
cd pon_watcher
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o pon_watcher cmd/main.go
```

## Configuration

The application supports configuration via YAML files and environment variables.

### Configuration File

Create a `config.yaml` file in the application directory:

```yaml
application:
  name: PONWatcher
  description: Intelligent Optical Monitoring
  version: 0.0.1
  logger_level: debug
  
  web:
    listen: 3000
    ssl_enabled: false
    ssl_cert: 
    ssl_key:
    no_route_to: https://example.com
    allow_origins:
      - "*"
    allow_methods:
      - GET
    allow_headers:
      - Content-Type 
      - Content-Length

services:
  unm:
    host: your-unm-server.com
    port: 3337
    username: your-username
    password: your-password
```

### Environment Variables

You can override configuration values using environment variables:

```bash
export APP_SERVICES_UNM_HOST=your-unm-server.com
export APP_SERVICES_UNM_PORT=3337
export APP_SERVICES_UNM_USERNAME=your-username
export APP_SERVICES_UNM_PASSWORD=your-password
export APP_APPLICATION_WEB_LISTEN=3000
export APP_APPLICATION_LOGGER_LEVEL=info
```

## Usage

### Starting the Application

```bash
# Using default config file (config.yaml)
./pon_watcher

# Using custom config file
./pon_watcher -config /path/to/config.yaml

# Enable configuration file watching for live reloads
./pon_watcher -watch-config

# Enable FX dependency injection debug logs
./pon_watcher -fx-debug
```

### Command Line Options

- `-config`: Path to configuration file (default: "config.yaml")
- `-watch-config`: Watch configuration file for changes
- `-fx-debug`: Enable/disable FX dependency injector logs

## API Endpoints

### Get PON Information

```http
GET /api/v1/pon-watcher?slot=<slot>&card=<card>&olt=<olt_id>&s=<filter>
```

**Parameters:**
- `slot` (required): PON slot number (positive integer)
- `card` (required): PON card number (positive integer)  
- `olt` (required): OLT ID (integer)
- `s` (optional): Filter string for ONU names/descriptions

**Response:**
```json
{
  "status": "OK",
  "data": {
    "online": {
      "onus": [
        {
          "splitter_name": "SPLITTER-01",
          "splitter_port": "1",
          "client_name": "CLIENT-NAME",
          "device_name": "SPLITTER-01|1-CLIENT-NAME",
          "onu_id": "12345",
          "rx_power": "-25.50",
          "rx_power_status": "Normal",
          "tx_power": "2.50",
          "tx_power_status": "Normal",
          "curr_tx_bias": "25.5",
          "curr_tx_bias_status": "Normal",
          "temperature": "45.2",
          "temperature_status": "Normal",
          "voltage": "3.3",
          "voltage_status": "Normal",
          "p_tx_power": "2.50",
          "p_rx_power": "-25.50"
        }
      ],
      "total": 1
    },
    "offline": {
      "onus": [],
      "total": 0
    },
    "total": 1
  }
}
```

## ONU Status Classification

ONUs are automatically categorized as:

- **Online**: ONUs with normal voltage status and non-zero RX power
- **Offline**: ONUs with "Low" voltage status and "0,00" RX power

## Performance Features

- **Concurrent Processing**: Configurable worker pools (max 20 workers) for parallel ONU data fetching
- **Connection Pooling**: Persistent TL1 connections with automatic reconnection
- **Request Debouncing**: Intelligent handling of configuration file changes
- **Benchmarking**: Built-in performance monitoring and timing

## Logging

The application features advanced logging capabilities:

- **Structured Logging**: JSON and console output formats
- **Colored Output**: Beautiful colored console logs with symbols
- **Log Levels**: trace, debug, info, warn, error, fatal, panic
- **Benchmarking**: Performance timing for operations
- **API Logging**: HTTP request/response logging with duration
- **Smart Context**: Contextual logging with fields and errors

## Development

### Project Structure

```
pon_watcher/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── adapter/                # External service adapters
│   ├── config/                 # Configuration management
│   ├── dto/                    # Data transfer objects
│   ├── fx/module/              # Dependency injection modules
│   ├── http/                   # HTTP handlers and responses
│   ├── repository/             # Data repositories
│   └── usecase/                # Business logic
├── pkg/                        # Shared packages
│   ├── banner/                 # Application banner
│   ├── http/                   # HTTP utilities
│   ├── log/                    # Logging framework
│   ├── tl1/                    # TL1 protocol transport
│   ├── unm/                    # UNM server client
│   └── viper/                  # Configuration loader
└── config.yaml                 # Configuration file
```

### Running in Development

```bash
# Run with debug logging
go run cmd/main.go -config config.debug.yml

# Run with live config reload
go run cmd/main.go -watch-config

# Run with FX debug logs
go run cmd/main.go -fx-debug
```

### Building

```bash
# Build for current platform
go build -o pon_watcher cmd/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o pon_watcher-linux cmd/main.go

# Build with version info
go build -ldflags "-X main.version=v1.0.0" -o pon_watcher cmd/main.go
```

## Dependencies

Key dependencies include:

- **Gin**: HTTP web framework
- **Uber FX**: Dependency injection framework
- **Zerolog**: High-performance logging
- **Viper**: Configuration management
- **CORS**: Cross-Origin Resource Sharing middleware

## Error Handling

The application includes comprehensive error handling:

- **Graceful Degradation**: Continues operation when some ONUs fail to respond
- **Connection Recovery**: Automatic reconnection to UNM servers
- **Validation**: Input parameter validation with detailed error messages
- **Timeout Handling**: Configurable timeouts for operations
- **Context Cancellation**: Proper handling of cancelled operations

## Monitoring

Built-in monitoring features:

- **Health Checks**: Connection status monitoring
- **Performance Metrics**: Operation timing and benchmarks  
- **Request Logging**: Detailed HTTP request/response logging
- **Error Tracking**: Comprehensive error logging and reporting

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Support

For support and questions:

- Create an issue in the repository
- Check the logs for detailed error information
- Ensure network connectivity to UNM servers
- Verify configuration parameters

## Changelog

### v0.0.1
- Initial release
- Basic ONU monitoring functionality
- TL1 protocol support
- RESTful API
- Concurrent processing
- Smart logging system