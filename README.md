# Go Autoconfig

> A simple, flexible autoconfig service for email servers

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go version](https://img.shields.io/badge/go-1.26.3%2B-00ADD8.svg)](https://golang.org/dl/)

## Problem

Everyone hates email configuration. Server maintainers especially hate it because users need to run clients to connect, and those clients have all created their own standards. Every mail client seems to have a different standard, and if everyone has a standard, nothing is standard.

[![XKCD has a comic for everything](https://imgs.xkcd.com/comics/standards.png)](https://xkcd.com/927/)

## Solution

Go Autoconfig is a lightweight, non-opinionated autoconfig service that speaks the language of major email clients. It abstracts away the complexity of supporting multiple mail clients (Mozilla Thunderbird, Microsoft Outlook, Apple Mail, etc.) by providing standardized endpoints that clients know how to query.

### Key Features

- **Multi-vendor support** — Automatically detects and serves configuration for Mozilla, Microsoft, Apple, and custom clients
- **Zero opinion** — Bring your own configuration JSON files; the service just serves them
- **Flexible validation** — Validate domains against environment variables or a database (MySQL, PostgreSQL, MariaDB)
- **Multiple domains** — Support as many domains as you need from a single instance
- **Simple setup** — Minimal dependencies, straightforward configuration
- **Golang** — Fast, single binary deployment with no runtime dependencies
- **Postfix/Dovecot ready** — Purpose-built for common mail server setups (but works anywhere)

## Quick Start

### Prerequisites

- Go 1.26.3 or higher
- (Optional) A database for domain/user validation

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/go-autoconfig.git
   cd go-autoconfig
   ```

2. **Build the binary**
   ```bash
   go build -o autoconfig
   ```

3. **Set up configuration**
   ```bash
   cp env.example .env
   # Edit .env with your settings
   nano .env
   ```

4. **Configure client settings**
   - Copy example JSON files from `clientConfigs/` and customize them
   - Update email servers, ports, and authentication methods as needed
   - Supported vendors: `autoconfig`, `autodiscover`, `mobileconfig`

5. **Run the service**
   ```bash
   ./autoconfig
   ```

The service listens on `:8080` by default. Configure with `LISTEN_ADDR` in your `.env` file.

## Configuration

### Environment Variables

Create a `.env` file in the project root (see `env.example`):

```bash
# Domain validation via environment variable (simple mode)
SUPPORTED_DOMAINS=example.com,mail.example.com

# OR validate against a database
ISDBENABLED=true
DBDRIVER=postgres              # mysql, mariadb, or postgres
DBHOST=localhost
DBPORT=5432
DBNAME=mailserver
DBUSER=mail_user
DBPASSW=secure_password
QUERY_DOMAINS=SELECT domain FROM domain WHERE domain = $1

# HTTP server
LISTEN_ADDR=:8080

# Per-vendor template overrides (optional)
TMPL_MOZILLA_HOSTNAME=mail.example.com
TMPL_MICROSOFT_HOSTNAME=mail.example.com
```

### Client Configuration

Client configurations live in `clientConfigs/` as JSON files:

- **`autoconfig.json`** — Mozilla Thunderbird, Evolution
- **`autodiscover.json`** — Microsoft Outlook
- **`mobileconfig.json`** — Apple Mail, iOS, macOS

Each file defines:
- Supported endpoints
- Email server details (IMAP, SMTP)
- Port numbers and encryption types
- Template variables for dynamic substitution

See `clientConfigs/` for examples.

### Templates

Email client configuration templates live in `templates/`:

- `autoconfig.xml.tmpl` — Thunderbird format
- `autodiscover.xml.tmpl` — Outlook format
- `mobileconfig.xml.tmpl` — Apple format

Templates use Go's `text/template` syntax with variables from client configs.

## API Endpoints

The service registers endpoints based on your client configurations. Standard endpoints include:

```
GET /.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=user@example.com
POST /autodiscover/autodiscover.xml (body: <Autodiscover><Request>...</Request></Autodiscover>)
GET /.well-known/mobileconfig/?emailaddress=user@example.com
```

Clients query these endpoints to retrieve their email configuration.

## Deployment

### Container Deployment (Recommended: Podman)

This project includes a production-ready `Dockerfile` with security best practices:

- **Multi-stage build** for minimal image size
- **Non-root user** (uid 1000) for container security
- **Alpine Linux** for a tiny, secure base image
- **Static linking** (CGO_ENABLED=0) for maximum portability

#### Using Podman (Recommended)

[Podman](https://podman.io/) is a secure, daemonless alternative to Docker that runs containers as your user (rootless by default). We **strongly recommend** Podman for production deployments.

```bash
# Build the image
podman build -t autoconfig .

# Run the container
podman run -d \
  --name autoconfig \
  -p 8080:8080 \
  --env-file .env \
  autoconfig
```

No root privileges required!

#### Using Docker (Rootless Mode Recommended)

If you use Docker, we recommend running it in **rootless mode** for security:

```bash
# Enable rootless mode (one-time setup)
dockerd-rootless-setuptool.sh install

# Build and run
docker build -t autoconfig .
docker run -d \
  --name autoconfig \
  -p 8080:8080 \
  --env-file .env \
  autoconfig
```

If you must use traditional Docker:

```bash
docker build -t autoconfig .
docker run -d \
  --name autoconfig \
  -p 8080:8080 \
  --env-file .env \
  autoconfig
```

#### Verify the Container

```bash
# Check logs
podman logs autoconfig  # or: docker logs autoconfig

# Test the service by querying an autoconfig endpoint
podman exec autoconfig wget -q -O - http://localhost:8080/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=test@example.com
```

#### Container Orchestration

For Kubernetes, Compose, or Swarm deployments, see the `Dockerfile` as your base. Environment variables are passed via `--env-file` or orchestration config.

**Using Docker Compose (Quickest way to get started):**

This repository includes a `docker-compose.yml` file for one-command deployment:

```bash
# Start the service and all dependencies
docker-compose up -d

# View logs
docker-compose logs -f autoconfig

# Stop the service
docker-compose down
```

The included `docker-compose.yml` provides:
- Pre-configured autoconfig service
- Optional PostgreSQL database (commented out by default)
- Log rotation and volume mounts
- Read-only config volume mounts

**Custom Docker Compose Example:**

```yaml
version: '3.8'
services:
  autoconfig:
    build: .
    ports:
      - "8080:8080"
    env_file: .env
    restart: unless-stopped
```

### Systemd Service

Create `/etc/systemd/system/autoconfig.service`:

```ini
[Unit]
Description=Go Autoconfig Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/autoconfig
EnvironmentFile=/opt/autoconfig/.env
ExecStart=/opt/autoconfig/autoconfig
Restart=always
User=mail
Group=mail

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable autoconfig
sudo systemctl start autoconfig
```

### Nginx Reverse Proxy

```nginx
upstream autoconfig {
    server localhost:8080;
}

server {
    listen 80;
    server_name mail.example.com;

    location / {
        proxy_pass http://autoconfig;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Development

### Project Structure

```
.
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration loading and validation
│   ├── db/                # Database connectivity
│   ├── handler/           # HTTP request handlers
│   ├── loader/            # Client config and template loading
│   ├── render/            # Template rendering
│   └── validate/          # Domain and user validation
├── clientConfigs/         # Client configuration JSON files
├── templates/             # XML template files
└── go.mod                 # Module definition
```

### Running Tests

```bash
go test ./...
```

### Building a Release

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o autoconfig-linux-amd64
```

## Troubleshooting

### No client configs found

**Error**: `no clientConfigs found — add at least one JSON file to clientConfigs/`

**Solution**: Ensure JSON files exist in the `clientConfigs/` directory and are valid JSON.

### Domain lookup fails

**Env-var mode**: Verify `SUPPORTED_DOMAINS` contains the domain in lowercase, comma-separated.

**Database mode**: Confirm:
- `ISDBENABLED=true` (not `ISDENABLED`)
- Database credentials are correct
- SQL queries are properly formatted (use `?` for MySQL, `$1` for PostgreSQL)
- Database contains domain/user records

### Connection refused

**Error**: `listen tcp :8080: bind: address already in use`

**Solution**: Change `LISTEN_ADDR` to a different port, e.g., `LISTEN_ADDR=:8081`

## Contributing

Contributions are welcome! Here's how:

1. **Fork** the repository
2. **Create a feature branch**: `git checkout -b feature/my-feature`
3. **Make your changes** and test them
4. **Submit a pull request** with a clear description

For major changes, please open an issue first to discuss.

## License

This project is licensed under the **GNU Affero General Public License v3** ([AGPL-3.0](LICENSE)).

The AGPL requires that any modifications to this software, when deployed as a network service, must be made available to users. See the [LICENSE](LICENSE) file for full details.

## Support

- **Issues**: Have a question or found a bug? [Open an issue](https://github.com/yourusername/go-autoconfig/issues)
- **Discussions**: Want to discuss a feature? [Start a discussion](https://github.com/yourusername/go-autoconfig/discussions)
- **Pull requests**: Ready to contribute? We'd love to see your changes!

## Acknowledgments

- Built with [Gin Web Framework](https://github.com/gin-gonic/gin)
- Inspired by Mozilla's Autoconfig and Microsoft's Autodiscover protocols
- Thanks to the Postfix and Dovecot communities for their excellent mail server software
