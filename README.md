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

   > The variables in `config.go` are used when running the app locally. The environnement variables set in the `docker-compose.yml` will override the values set in `backend/config.go` when running the project with Docker Compose.

## Running the project

### Run locally (without Docker)

From the `backend` directory:

```bash
go run .
```

This will start the API server on http://127.0.0.1:8080.

- Health endpoint: http://127.0.0.1:8080/health (GET)
- Login endpoint: http://127.0.0.1:8080/login (POST)
- Home page endpoint: http://127.0.0.1:8080/ (GET)
- Orders page endpoint: http://127.0.0.1:8080/order (GET)

> You must have a MySQL instance running on your machine for all uses cases to work

### Run with Docker Compose

From the project root:

```bash
docker compose down -v # To remove existing containers and their volumes
docker compose up --build -d
```

This starts:

- The Go backend (`brokerx_app`) on [http://localhost:8080](http://localhost:8080)
- A MySQL database (`brokerx_db`) on port `3306`

## Running tests

**Important** : Before running all tests, you must have the test database up, because some of the tests are integration tests needing a real database.

Run the following command to start the test database:

```bash
#From the root of the project
docker compose -f docker-compose.test.yml up -d
```

### Running all tests (with coverage report)

Inside the `backend` folder:

```bash
go test ./... -coverprofile=coverage
```

### Generate HTML coverage report

```bash
go tool cover -html=coverage
```

## Deployment

At this stage, the application is deployed locally or remotely using Docker Compose.
A production-ready deployment would likely use Kubernetes or cloud-based services, but that is outside the current scope.

### Deploying locally

To deploy locally, you just have to run the following commands

```bash
docker compose down -v # Ensure Docker is clean with no previous deployment
docker compose up --build -d # Build the Docker image and run docker compose in detached mode
```

### Deploying remotely

The GitHub Actions Workflow should take care of deploying the application to the ETS Virtual Machine self hosted runner automatically on every push. See `.github/workflows/ci_cd.yml`

_The current deployment pipeline is broken because of issues with the storage on the ETS VM._

> [!NOTE]
> To access the remote deployment, you must be connected to the ETS Cisco Secure Client via accesvpn.etsmtl.ca

## Using BrokerX

See `docs/Runbook.md` for information on how to use BrokerX when it is deployed.

## Documentation

Refer to the architectural documentation [here](https://github.com/cjayneb/log430-project/blob/main/docs/arc42.md).

### Refs

- https://threedots.tech/post/database-transactions-in-go/

## Authors

- [Jean-Christophe Benoît](https://github.com/cjayneb)
