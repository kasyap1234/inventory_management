# Product Overview

Agromart2 is a comprehensive SaaS inventory management platform designed for agricultural businesses. The system provides multi-tenant capabilities with role-based access control (RBAC) for managing products, orders, suppliers, distributors, warehouses, and analytics.

## Key Features

- **Multi-tenant Architecture**: Supports multiple businesses with isolated data
- **Inventory Management**: Product catalog, stock tracking, warehouse management
- **Order Processing**: Order creation, tracking, and fulfillment
- **Supplier/Distributor Management**: Vendor relationship management
- **Analytics & Reporting**: Business intelligence and performance metrics
- **Authentication & Authorization**: JWT-based auth with RBAC
- **File Management**: Image uploads via MinIO object storage
- **Background Jobs**: Async processing with Redis/Asynq
- **Tally Integration**: ERP system integration for accounting

## Architecture

The application follows a clean architecture pattern with:
- RESTful API backend in Go
- React/Next.js frontend with TypeScript
- PostgreSQL database with Redis caching
- Microservices-ready design with Docker containerization