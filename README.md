# Agromart Inventory Management System

Agromart is a comprehensive inventory management system designed for AgroTech companies. It features a robust backend written in Go and a modern, responsive frontend built with Next.js and Tailwind CSS.

## Prerequisites

Before you begin, ensure you have the following installed:

-   **Go**: Version 1.21 or higher
-   **Node.js**: Version 18 or higher (or Bun)
-   **Docker**: For running the database and other dependent services
-   **Make**: (Optional) For running convenience commands

## Getting Started

### 1. Infrastructure Setup

The easiest way to set up the required infrastructure (PostgreSQL, Redis, MinIO, Mailpit) is using Docker Compose.

```bash
# Start all services
docker-compose up -d
```

This will start:
-   **PostgreSQL**: Port 5440 (Database: `testdb`, User: `testuser`, Password: `testpass`)
-   **Redis**: Port 6379
-   **MinIO**: Port 9000 (Console: 9001)
-   **Mailpit**: Port 1025 (Web UI: 8025)

### 2. Backend Setup

1.  **Navigate to the root directory.**

2.  **Environment Variables**:
    Ensure you have a `.env` file in the root directory. You can copy the example provided:
    ```bash
    cp .env.example .env
    ```
    *Note: The project comes with a pre-configured `.env` file for local development.*

3.  **Run Migrations**:
    Initialize the database schema.
    ```bash
    ./run_migrations.sh
    ```

4.  **Start the Server**:
    ```bash
    go run cmd/main.go
    ```
    The backend server will start on `http://localhost:8080`.

### 3. Frontend Setup

1.  **Navigate to the frontend directory**:
    ```bash
    cd frontend
    ```

2.  **Install Dependencies**:
    ```bash
    npm install
    # or
    bun install
    ```

3.  **Start the Development Server**:
    ```bash
    npm run dev
    # or
    bun dev
    ```
    The frontend application will be available at `http://localhost:3000`.

## Accessing the Application

-   **Frontend URL**: [http://localhost:3000](http://localhost:3000)
-   **Backend API**: [http://localhost:8080](http://localhost:8080)
-   **Mailpit (Email Testing)**: [http://localhost:8025](http://localhost:8025)

## Payments & Storage Configuration

- **Razorpay**: set `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, and `RAZORPAY_WEBHOOK_SECRET` in `.env`. The webhook endpoint is `POST /v1/webhooks/razorpay`; point the Razorpay dashboard to `https://<backend>/v1/webhooks/razorpay`.
- **Checkout config**: the frontend reads the public key from `GET /payments/config` (no secret values are exposed).
- **One-time payments**: create orders via `POST /payments/orders`, verify with `POST /payments/verify`, and webhooks handle `payment.authorized/captured/failed` with idempotency.
- **Subscriptions**: `POST /subscriptions` creates Razorpay subscriptions; webhooks update lifecycle events.
- **MinIO**: set `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_USE_SSL`. Buckets `product-images` and `invoices` are created automatically at startup.
- **Image uploads**: request a presigned URL via `POST /products/:id/images/presign`, upload directly to MinIO, then call `POST /products/:id/images/finalize` to generate thumbnails and persist metadata.

## Features

-   **Dashboard**: Real-time overview of sales, stock, and orders.
-   **Inventory Management**: Track products, batches, and stock levels across warehouses.
-   **Order Management**: Process orders and generate invoices.
-   **User Management**: Role-based access control (RBAC).
-   **Dark Mode**: Fully supported dark theme.

## Troubleshooting

-   **Database Connection Issues**: Ensure Docker is running and port 5440 is not occupied. Check the `DATABASE_URL` in your `.env` file.
-   **Migration Errors**: If migrations fail, try resetting the database volume or checking the logs.
