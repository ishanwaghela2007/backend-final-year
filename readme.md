# Final Year Project - Factory Management System

## Overview

This is the backend repository for the Final Year Project, designed as a distributed microservices system to manage a factory environment. It includes services for authentication, camera feed streaming, and machine learning-based defect detection.

## Architecture

The system follows a microservices architecture where each service is independent and communicates via HTTP/gRPC.

### Services

- **Auth Service (`/Auth`)**: Manages user authentication, authorization, and administration.
- **Camera Service (`/camera`)**: Handles CCTV camera stream access and channel management.
- **ML Model Service (`/MLmodel`)**: Python/FastAPI service for defect detection and batch inspection reporting.
- **Feedback Service (`/Feedback`)**: Handles user feedback (structure TBD).

## Tech Stack

- **Languages**: Go (Golang), Python
- **Frameworks**: Gin (Go), FastAPI (Python)
- **Databases**: 
    - **Apache Cassandra**: Primary NoSQL database for user data and inspection results.
    - **Redis**: In-memory store for rate limiting and session management.
- **Messaging**: Apache Kafka (for async tasks like email processing).
- **Containerization**: Docker & Docker Compose.

## Prerequisites

- [Go](https://go.dev/) (v1.20+)
- [Python](https://www.python.org/) (v3.9+)
- [Docker & Docker Compose](https://www.docker.com/)
- [Apache Kafka](https://kafka.apache.org/) (if running locally without Docker)
- [Cassandra](https://cassandra.apache.org/) (if running locally without Docker)
- [Redis](https://redis.io/) (if running locally without Docker)

## Setup & Running

### 1. Environment Setup

Ensure you have the required databases running. You can use the provided Docker Compose files in the `Auth` directory or run them individually.

### 2. Running Auth Service

```bash
cd Auth
# Create .env file based on .env.example
cp .env.example .env
go mod tidy
go run main.go
```
*Runs on Port: `8080` (default)*

### 3. Running Camera Service

```bash
cd camera
# Ensure .env is set up
go mod tidy
go run main.go
```
*Runs on Port: `3000`*

### 4. Running ML Model Service

```bash
cd MLmodel
# Install dependencies
pip install -r requirment.txt
# Run FastAPI server
uvicorn fastapi_server:app --reload
```
*Default FastAPI Port: `8000`*

## API Endpoints Overview

### Auth Service (`/api/v0`)
- `POST /login`: User login.
- `GET /oauth`: OAuth login.
- `POST /logout`: User logout.
- `GET /admin/users`: (Admin) Manage users.

### Camera Service (`/api/v0/cctv`)
- `GET /stream/channel[1-4]`: Stream video feeds from different camera channels.

### ML Service
- `POST /add_inspection`: Submit inspection results (defects like scratch, crack, bend, hole).
- `GET /batch_status/{batch_id}`: Get production status for a specific batch.

<!-- ## Database Schema -->

<!-- ### Cassandra Keyspaces
- `user`: Managed by Auth service.
- `tube_factory`: Managed by ML service (tables: `inspection_results`, `batch_status`). -->

## License

See `licence.md` for details.
