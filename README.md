# Jackpot Bet Admin Stats API

A high-performance Go-based REST API designed to aggregate casino transaction data from MongoDB. This project includes a data seeder and uses Gin for the API, Redis for caching, and Zap for structured logging.

## 🚀 Quick Start (Docker)

The easiest way to run the application is using Docker Compose.

### Prerequisites
- Docker (with BuildKit enabled)
- Docker Compose

### 1. Setup Environment
Cloning the project is the first step. Before running, you **must create a `.env` file** based on the provided example:

```bash
cp .env.example .env
```
Ensure the values in `.env` match your desired configuration.

### 2. Launch the Application
Start the infrastructure (MongoDB, Redis) and the API:

```bash
make docker-run
```
The API will be available at `http://localhost:8080`.

> [!TIP]
> If `make docker-run` fails because MongoDB is taking too long to start (healthcheck timeout), simply run the command again: `make docker-run`. Once the database is healthy, the API will start successfully.

### 3. Seed the Data
The application requires a large dataset to demonstrate aggregation performance. Run the seeder using the following command:

```bash
make docker-seed
```
*Note: The seeder inserts ~2,000,000 rounds (~4,000,000 transactions) by default. This is optimized to use the same Docker image as the API to avoid redundant builds.*

---

## 🛠 Manual Development

If you prefer to run the services locally without Docker:

```bash
# 1. Start DBs
docker-compose up -d mongodb redis

# 2. Run API
make run

# 3. Seed data
make seed
```

---

## 📡 API Endpoints

All endpoints require the `Authorization` header.

| Endpoint | Description |
| :--- | :--- |
| `GET /gross_gaming_rev` | Calculate GGR (Wagers - Payouts) per currency. |
| `GET /daily_wager_volume` | Get daily wager volume trends. |
| `GET /user/:id/wager_percentile` | Find a user's wager ranking percentile. |

### Example Queries

```bash
# Gross Gaming Revenue (GGR)
curl --location 'http://localhost:8080/gross_gaming_rev?from=2025-01-01&to=2025-06-30' \
--header 'Authorization: super-secret-admin-key'

# Daily Wager Volume
curl --location 'http://localhost:8080/daily_wager_volume?from=2025-01-01&to=2025-06-30' \
--header 'Authorization: super-secret-admin-key'

# User Wager Percentile
# Replace :userID with one of the sample User IDs printed during 'make docker-seed'
curl --location 'http://localhost:8080/user/:userID/wager_percentile?from=2025-01-01&to=2025-06-30' \
--header 'Authorization: super-secret-admin-key'
```

> **Tip**: When you run the seeder, it logs a set of sample User IDs to the console. Use those IDs to test the percentile endpoint!

---

## 🧪 Makefile Commands

```bash
make build       # Build binaries
make run         # Run API locally
make seed        # Run seeder locally
make docker-run   # Start docker containers (API + DB)
make docker-seed  # Run seeder in docker
make docker-down  # Stop docker containers
```
