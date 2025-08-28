# 📦 Order Service

## ✨ Description
This project is a service for processing orders.  

### 🔑 Key features:
-  Receiving and storing orders via **Kafka**.  
-  **Dead Letter Queue (DLQ)**: invalid orders are redirected to a separate Kafka topic.  
-  **Caching** orders for fast access (cache size is configurable in `.env`).  
-  **REST API**: retrieve orders by UID → `GET /orders/:uid`.  
-  Generate test orders (valid ✅ and invalid ❌).  
- **Graceful shutdown** of the HTTP server and Kafka consumer.  
-  **Flexible logging**: configure **log level** and **output format (JSON or text)** via `.env`.  
- **Cache warm-up on startup**: preloads recent orders from the database into the cache for faster first access.


---

## ⚙️ Prerequisites
- 🐳 Docker & Docker Compose  
- 🐹 Go (1.25.0)  

📍 **Service endpoint:**  
- http://localhost:8081/orders/:uid  

📊 **Kafka UI:**  
- http://localhost:8080  

---


```mermaid
flowchart TD
    %% Producers
    P[Test Orders] --> K[Kafka Topic]

    %% Kafka consumer
    K -->|valid| C[Service Consumer]
    K -->|invalid| DLQ[Kafka DLQ]

    %% Repository + DB + Cache
    C --> R[Repository]
    R -->|save| DB[(Postgres)]
    R -->|save| Cache[In-Memory Cache]

    %% Cache warm-up at startup
    DB -.warmup.-> Cache
    %% Cache miss fallback inside Repository
    Cache -.miss.-> DB

    %% REST API GET
    API[API GET /orders/:uid] -->|request| R
    R -->|read| Cache
    Cache -->|response| API

    %% Colors + black text
    style P fill:#f9f,stroke:#333,stroke-width:2px,color:#000
    style K fill:#bbf,stroke:#333,stroke-width:2px,color:#000
    style DLQ fill:#fbb,stroke:#333,stroke-width:2px,color:#000
    style C fill:#bfb,stroke:#333,stroke-width:2px,color:#000
    style R fill:#ffb,stroke:#333,stroke-width:2px,color:#000
    style DB fill:#bbb,stroke:#333,stroke-width:2px,color:#000
    style Cache fill:#fbf,stroke:#333,stroke-width:2px,color:#000
    style API fill:#bff,stroke:#333,stroke-width:2px,color:#000
```

## 🚀 Installation and Run

### 1️⃣ Clone the repository and configure `.env`

### 2️⃣ Build and start all services
```bash
make up
```
#### 3️⃣ Run the service and create the Kafka topic "orders":
```bash
make run
```
#### 4️⃣ In another terminal, you can send and retrieve orders using the scripts and generated test orders:
```bash
make post-get-order
```
![Demo](order_show.gif)
