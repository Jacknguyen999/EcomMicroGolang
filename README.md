# ECOMMERCE MICROSERVICES

This repository hosts a sample **e-commerce platform** demonstrating a **microservices architecture** using multiple **Go services** alongside a **Python-based Recommender service**. The project showcases:

- gRPC communication between services
- Kafka-based event streaming pipeline
- A unified GraphQL API gateway for clients
- Elasticsearch integration for product search

---

## 📚 Table of Contents

- [Overview](#-overview)
- [Architecture Diagram](#-architecture-diagram)
- [Services](#-services)
- [Getting Started](#-getting-started)
- [Usage](#-usage-graphql)
- [Contributing](#-contributing)
- [Author](#-author)
- [License](#-license)

---

## 🧭 Overview

The system comprises several microservices:

- **Account** (Go): Manages user accounts, authentication, and authorization.
- **Product** (Go): CRUD for products; indexes product data in **Elasticsearch**.
- **Order** (Go): Handles order creation and persistence; publishes events to Kafka.
- **Recommender** (Python): Consumes Kafka events and builds product recommendations.
- **API Gateway** (Go): A GraphQL service exposing a unified API for front-end clients.

The entire ecosystem is containerized using **Docker Compose**. Datastores include **PostgreSQL**, **Elasticsearch**, and **Kafka**.

---

## 🏗 Architecture Diagram

Below is a high-level overview of the system architecture:

![API Design](./design.png)

### Communication Overview

- `API Gateway (GraphQL)` talks to:

  - `Account client` → `Account server` → `Postgres`
  - `Product client` → `Product server` → `ElasticSearch`
  - `Order client` → `Order server` → `Postgres` + Kafka
    (also communicates with Product service via gRPC)
  - `Recommender client` → `Recommender server` (Python) → `Postgres (Replica)`

- **Event Flow**:
  - `Order` and `Product` services act as **Kafka producers**.
  - `Recommender` service is a **Kafka consumer**, ingesting order/product events and updating internal state for recommendations.

---

## ⚙ Services

### 🧑‍💼 Account Service (Go)

- Responsibilities: Register, login, fetch account data, generate JWT tokens.
- Database: PostgreSQL

### 📦 Product Service (Go)

- Responsibilities: Product CRUD operations, indexing to Elasticsearch, event publishing to Kafka.
- Features: Advanced filtering by price range and category, sorting by price and other criteria.
- Database: Elasticsearch

### 🛒 Order Service (Go)

- Responsibilities: Order creation, price calculation, data persistence, Kafka event publishing.
- Dependencies: Calls product service to retrieve product info.

### 🧠 Recommender Service (Python)

- Responsibilities: Kafka consumer that builds recommendations based on product/order events.
- Tech Stack: Python + gRPC + PostgreSQL (replica of product DB)

### 🚪 API Gateway (Go)

- Responsibilities: Unified GraphQL endpoint at `/graphql`.
- Implementation: Uses gRPC clients for all microservices and schema stitching.

---

## 🚀 Getting Started

### ✅ Prerequisites

Before running the project, ensure you have the following installed:

- [Docker](https://www.docker.com/get-started) & [Docker Compose](https://docs.docker.com/compose/)
- [Git](https://git-scm.com/)

---

### 📥 Clone the Repository

```bash
  git clone https://github.com/Jacknguyen999/EcomMicroGolang.git
  cd EcomMicroGolang
```

---

### 🐳 Run the Stack

To build and start all services using Docker Compose, run:

```bash
  docker-compose up --build
```

This will start:

- Go microservices (`account`, `order`, `product`, `graphql`)
- Python-based `recommender` service
- Databases: PostgreSQL, Elasticsearch
- Kafka + Zookeeper
- GraphQL gateway

---

### 🌐 Access the API

Once everything is running, open your browser to:

- **GraphQL API endpoint**:
  [http://localhost:8080/graphql](http://localhost:8080/graphql)

- **GraphQL Playground (interactive testing)**:
  [http://localhost:8080/playground](http://localhost:8080/playground)

---

## 📬 Usage (GraphQL)

Below are example GraphQL queries and mutations you can test in the [GraphQL Playground](http://localhost:8080/playground).

---

### 📝 Register a New Account

```graphql
mutation {
  register(
    account: {
      name: "Alice"
      email: "alice@example.com"
      password: "secret123"
    }
  ) {
    token
  }
}
```

---

### 🔐 Login

```graphql
mutation {
  login(account: { email: "alice@example.com", password: "secret123" }) {
    token
  }
}
```

---

### ➕ Create a Product

```graphql
mutation {
  createProduct(
    product: { name: "Camera", description: "A digital camera", price: 99.99 }
  ) {
    id
    name
  }
}
```

---

### 🔍 Query Products

```graphql
query {
  product(pagination: { skip: 0, take: 10 }) {
    id
    name
    price
  }
}
```

### 🔍 Filter Products by Price Range

```graphql
query FilterProducts($priceRange: PriceRangeInput) {
  product(pagination: { skip: 0, take: 10 }, priceRange: $priceRange) {
    id
    name
    price
  }
}
```

Variables:

```json
{
  "priceRange": {
    "min": 50,
    "max": 100
  }
}
```

### 🔍 Filter Products by Category

```graphql
query FilterProducts($category: String) {
  product(pagination: { skip: 0, take: 10 }, category: $category) {
    id
    name
    price
  }
}
```

Variables:

```json
{
  "category": "electronics"
}
```

### 🔍 Sort Products

```graphql
query FilterProducts($sortBy: SortOrder) {
  product(pagination: { skip: 0, take: 10 }, sortBy: $sortBy) {
    id
    name
    price
  }
}
```

Variables:

```json
{
  "sortBy": "PRICE_ASC"
}
```

Available sort orders:

- `PRICE_ASC`: Sort by price, lowest to highest
- `PRICE_DESC`: Sort by price, highest to lowest
- `NEWEST`: Sort by newest first
- `POPULARITY`: Sort by popularity (not implemented yet)

### 🔍 Combine Filters and Sorting

```graphql
query FilterProducts(
  $category: String
  $priceRange: PriceRangeInput
  $sortBy: SortOrder
) {
  product(
    pagination: { skip: 0, take: 10 }
    category: $category
    priceRange: $priceRange
    sortBy: $sortBy
  ) {
    id
    name
    price
  }
}
```

Variables:

```json
{
  "category": "electronics",
  "priceRange": {
    "min": 50,
    "max": 100
  },
  "sortBy": "PRICE_DESC"
}
```

---

### 🛒 Create an Order

```graphql
mutation {
  createOrder(order: { products: [{ id: "PRODUCT_ID", quantity: 2 }] }) {
    id
    totalPrice
    products {
      name
      quantity
    }
  }
}
```

## 🤝 Contributing

We welcome contributions! To contribute:

Fork the repository

Create a new branch

Commit and push your changes

Open a Pull Request

## 👤 Author

## 🪪 License

This project is licensed under the Apache License 2.0.
