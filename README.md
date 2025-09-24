# log430-project

Repository for **BrokerX** project from LOG430 (Software Architecture).

## Prerequisites

To run this project locally, you need the following tools installed:

- [Go 1.21+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)

## Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/cjayneb/log430-project.git
   cd log430-project
   ```

2. **Configure environment variables**

   Change the values in the `backend/config.go` file if you want to override defaults.
   Example:

   ```env
    Port  string `env:"APP_PORT" envDefault:"8181"`
    DBUrl string `env:"DATABASE_URL" envDefault:"user:pass@tcp(127.0.0.1:3306)/brokerx"`
   ```

   > The environnement variables set in the `docker-compose.yml` will override the values set in `backend/config.go` when running the project with Docker Compose.

## Running the project

### Run locally (without Docker)

From the `backend` directory:

```bash
go run .
```

This will start the API server on http://127.0.0.1:8080.

- Health endpoint: http://127.0.0.1:8080/health (GET)
- Login endpoint: http://127.0.0.1:8080/login (POST)

> You must have a MySQL instance running on your machine for this to work

### Run with Docker Compose

From the project root:

```bash
docker compose down -v # To remove existing containers and their volumes
docker compose up --build -d
```

This starts:

- The Go backend (`brokerx_app`) on [http://localhost:8080](http://localhost:8080)
- A MySQL database (`brokerx_db`) on port `3306`

### Run tests

Inside the `backend` folder:

```bash
go test ./...
```

### Generate coverage report

```bash
go test ./... -cover
```

## Deployment

At this stage, the application is deployed manually using Docker Compose.  
A production-ready deployment would likely use Kubernetes or cloud-based services, but that is outside the current scope.

## Documentation

Refer to the architectural documentation [here](https://github.com/cjayneb/log430-project/blob/main/docs/arc42.md).

## Authors

- [Jean-Christophe Benoît](https://github.com/cjayneb)
