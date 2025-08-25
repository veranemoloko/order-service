## Order Service

### Description
#### This project is a microservice for processing orders. Key features include:
- Receiving and storing orders via Kafka.
- Caching orders for fast access. The cache size can be configured in the .env file.
- REST API for retrieving orders by UID. (GET /orders/:uid)
- Generating test orders (both valid and invalid) for testing purposes.
- Graceful shutdown of the HTTP server and Kafka consumer upon receiving termination signals.

### Prerequisites
- Docker & Docker Compose
- Go (1.25.0)

### Installation and Run
#### 1. Clone the repository and configure .env
#### 2. Build and start all services using docker-compose:
```bash
make up
```
#### 3. Run the service and create the Kafka topic "orders":
```bash
make run
```
#### 4. In another terminal, you can send and retrieve orders using the scripts and generate test orders:
```bash
make post-get-order
```

#### The service will be accessible at: 
- http://localhost:8081/orders/:uid

#### Kafka UI is available at:
- http://localhost:8080 for monitoring topics and messages.


