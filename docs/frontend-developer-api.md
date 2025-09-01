# Frontend Developer Guide - Agromart2 Backend API

## Overview

Agromart2 is a multi-tenant e-commerce platform designed for agricultural and rural businesses. The backend provides comprehensive APIs for product management, order processing, user management, and file handling with full multi-tenant support and role-based access control.

### Key Features
- **Authentication & User Management**: JWT-based authentication with multi-tenant support
- **Product Catalog Management**: Complete CRUD operations for products with category support
- **Order Processing**: Full e-commerce workflow from order creation to invoice generation
- **File Management**: Product image upload/download with MinIO integration
- **Multi-Tenant Architecture**: Isolated data per tenant with shared platform resources
- **Role-Based Access Control**: Granular permissions for different user types
- **Business Operations**: Inventory, supplier, distributor, and warehouse management

## Getting Started

### Base URL
```
http://localhost:8080
```

### Content-Type
All requests should use JSON content type:
```
Content-Type: application/json
```

### Authentication
Most endpoints require JWT authentication. Include the Bearer token in headers:
```
Authorization: Bearer <your-jwt-token>
```

---

## Authentication APIs

### User Registration
Register a new user account.

**Endpoint**: `POST /v1/auth/signup`
**Authentication**: None required

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response** (201):
```json
{
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "tenant_id": "tenant-uuid",
    "role": "user",
    "created_at": "2025-01-01T10:00:00Z"
  },
  "access_token": "jwt-access-token",
  "refresh_token": "jwt-refresh-token",
  "expires_in": 3600
}
```

### User Login
Authenticate and get access tokens.

**Endpoint**: `POST /v1/auth/login`
**Authentication**: None required

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response** (200):
```json
{
  "access_token": "jwt-access-token",
  "refresh_token": "jwt-refresh-token",
  "expires_in": 3600,
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "tenant_id": "tenant-uuid"
  }
}
```

### User Logout
Invalidate the current access token.

**Endpoint**: `POST /v1/auth/logout`
**Authentication**: Required

**Response** (200):
```json
{
  "message": "Logged out successfully"
}
```

---

## User Management APIs

### Get User Profile
Retrieve current user information.

**Endpoint**: `GET /v1/me`
**Authentication**: Required

**Response** (200):
```json
{
  "id": "uuid-string",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "tenant_id": "tenant-uuid",
  "role": "user",
  "created_at": "2025-01-01T10:00:00Z",
  "updated_at": "2025-01-01T10:00:00Z"
}
```

### List Users
Get paginated list of users.

**Endpoint**: `GET /v1/users`
**Authentication**: Required

**Query Parameters**:
- `limit` (optional): Records per page (default: 10)
- `offset` (optional): Records to skip (default: 0)

**Response** (200):
```json
{
  "users": [
    {
      "id": "uuid-string",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "user",
      "created_at": "2025-01-01T10:00:00Z"
    }
  ],
  "total": 25,
  "limit": 10,
  "offset": 0
}
```

---

## Product Management APIs

### List Products
Get paginated list of products.

**Endpoint**: `GET /v1/products`
**Authentication**: Required

**Query Parameters**:
- `category_id` (optional): Filter by category
- `search` (optional): Search term
- `limit` (optional): Records per page (default: 10)
- `offset` (optional): Records to skip (default: 0)

**Response** (200):
```json
{
  "products": [
    {
      "id": "uuid-string",
      "name": "Rice Seeds",
      "description": "High-quality rice seeds",
      "price": 25.50,
      "category": {
        "id": "uuid-string",
        "name": "Seeds"
      },
      "tenant_id": "tenant-uuid",
      "created_at": "2025-01-01T10:00:00Z"
    }
  ],
  "total": 15,
  "limit": 10,
  "offset": 0
}
```

### Create Product
Create a new product.

**Endpoint**: `POST /v1/products`
**Authentication**: Required

**Request Body**:
```json
{
  "name": "Wheat Seeds",
  "description": "Premium wheat seeds",
  "price": 35.50,
  "category_id": "category-uuid"
}
```

**Response** (201):
```json
{
  "id": "uuid-string",
  "name": "Wheat Seeds",
  "description": "Premium wheat seeds",
  "price": 35.50,
  "category": {
    "id": "category-uuid",
    "name": "Seeds"
  },
  "tenant_id": "tenant-uuid",
  "created_at": "2025-01-01T10:00:00Z"
}
```

### Get Product
Retrieve a specific product by ID.

**Endpoint**: `GET /v1/products/{id}`
**Authentication**: Required

**Response** (200): Same as single product from list

### Update Product
Update product information.

**Endpoint**: `PUT /v1/products/{id}`
**Authentication**: Required

**Request Body**:
```json
{
  "name": "Premium Wheat Seeds",
  "description": "Premium quality wheat seeds",
  "price": 40.00
}
```

**Response** (200): Updated product object

### Delete Product
Remove a product.

**Endpoint**: `DELETE /v1/products/{id}`
**Authentication**: Required

**Response** (200):
```json
{
  "message": "Product deleted successfully"
}
```

---

## Product Images APIs

### Upload Product Image
Upload an image for a product.

**Endpoint**: `POST /v1/products/{id}/images`
**Authentication**: Required

**Content-Type**: `multipart/form-data`

**Form Fields**:
- `image`: Image file (PNG, JPG, JPEG)

**Response** (201):
```json
{
  "id": "image-uuid",
  "product_id": "product-uuid",
  "filename": "seed-image.jpg",
  "content_type": "image/jpeg",
  "size": 2048000,
  "uploaded_at": "2025-01-01T10:00:00Z"
}
```

### List Product Images
Get all images for a product.

**Endpoint**: `GET /v1/products/{id}/images`
**Authentication**: Required

**Response** (200):
```json
{
  "images": [
    {
      "id": "image-uuid",
      "filename": "seed-image.jpg",
      "content_type": "image/jpeg",
      "size": 2048000,
      "uploaded_at": "2025-01-01T10:00:00Z"
    }
  ]
}
```

### Get Image Download URL
Get presigned URL for image download.

**Endpoint**: `GET /v1/products/{id}/images/{imageId}/url`
**Authentication**: Required

**Response** (200):
```json
{
  "url": "https://minio.example.com/product-images/product-uuid/image-uuid.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256...",
  "expires_in": 86400
}
```

---

## Order Processing APIs

### Create Order
Create a new order.

**Endpoint**: `POST /v1/orders`
**Authentication**: Required

**Request Body** (single item):
```json
{
  "product_id": "product-uuid",
  "quantity": 5,
  "delivery_address": "123 Farm Road, Rural District"
}
```

**Request Body** (bulk items):
```json
{
  "orders": [
    {
      "product_id": "product-uuid-1",
      "quantity": 5,
      "delivery_address": "123 Farm Road"
    },
    {
      "product_id": "product-uuid-2",
      "quantity": 10,
      "delivery_address": "123 Farm Road"
    }
  ]
}
```

**Response** (201):
```json
{
  "id": "order-uuid",
  "product_id": "product-uuid",
  "quantity": 5,
  "total_amount": 67.50,
  "status": "pending",
  "delivery_address": "123 Farm Road",
  "created_at": "2025-01-01T10:00:00Z"
}
```

### List Orders
Get paginated orders (purchase orders).

**Endpoint**: `GET /v1/orders`
**Authentication**: Required

**Query Parameters**:
- `limit`, `offset`: Pagination parameters

**Response** (200):
```json
{
  "orders": [
    {
      "id": "order-uuid",
      "product": {
        "id": "product-uuid",
        "name": "Rice Seeds"
      },
      "quantity": 5,
      "total_amount": 67.50,
      "status": "pending",
      "created_at": "2025-01-01T10:00:00Z"
    }
  ],
  "total": 10,
  "limit": 10,
  "offset": 0
}
```

---

## Invoice Management APIs

### List Invoices
Get paginated invoices.

**Endpoint**: `GET /v1/invoices`
**Authentication**: Required

**Response** (200):
```json
{
  "invoices": [
    {
      "id": "invoice-uuid",
      "order_id": "order-uuid",
      "total_amount": 67.50,
      "status": "unpaid",
      "issued_date": "2025-01-01T10:00:00Z",
      "created_at": "2025-01-01T10:00:00Z"
    }
  ],
  "total": 5,
  "limit": 10,
  "offset": 0
}
```

### Get Invoice
Retrieve specific invoice.

**Endpoint**: `GET /v1/invoices/{id}`
**Authentication**: Required

### Update Invoice Status
Update invoice payment status.

**Endpoint**: `PUT /v1/invoices/{id}`
**Authentication**: Required

**Request Body**:
```json
{
  "status": "paid"
}
```

### Generate Invoice PDF
Generate invoice PDF and get download URL.

**Endpoint**: `POST /v1/invoices/{id}/generate-pdf`
**Authentication**: Required

**Response** (200):
```json
{
  "message": "PDF generated and uploaded successfully",
  "pdf_url": "https://minio.example.com/invoices/download-url",
  "expires_in": 86400
}
```

---

## Business Management APIs

### Categories
- `GET /v1/categories` - List all categories
- `POST /v1/categories` - Create new category
- `GET /v1/categories/{id}` - Get specific category
- `PUT /v1/categories/{id}` - Update category
- `DELETE /v1/categories/{id}` - Delete category

### Warehouses
- `GET /v1/warehouses` - List warehouses
- `POST /v1/warehouses` - Create warehouse
- `GET /v1/warehouses/{id}` - Get warehouse details

### Suppliers and Distributors
- `GET /v1/suppliers` - List suppliers ( RBAC permissions may be required)
- `GET /v1/distributors` - List distributors ( RBAC permissions may be required)

---

## System Health APIs

### Health Check
Check system health (no authentication).

**Endpoint**: `GET /health`

**Response**:
```json
{
  "service": "agromart2",
  "status": "ok",
  "timestamp": "2025-01-01T10:00:00Z"
}
```

### Readiness Check
Check service readiness.

**Endpoint**: `GET /health/ready`

### Detailed Health
Get detailed health with dependencies.

**Endpoint**: `GET /health/detailed`

### Metrics
Get application metrics.

**Endpoint**: `GET /metrics`

---

## Documentation APIs

### API Guide
Get general API usage guide.

**Endpoint**: `GET /v1/docs/guide`

### API Specification
Get API endpoint specifications.

**Endpoint**: `GET /v1/docs/spec`

---

## Error Handling

### Common HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

### Error Response Format
```json
{
  "message": "Error description",
  "error": "detailed_error_message"
}
```

### Common Validation Errors
```json
{
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "Email is required"
    }
  ]
}
```

---

## Integration Examples

### 1. User Registration Flow
```javascript
// Register new user
const registerUser = async (userData) => {
  const response = await fetch('/v1/auth/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(userData)
  });

  const data = await response.json();
  // Store tokens for future requests
  if (data.access_token) {
    localStorage.setItem('accessToken', data.access_token);
    localStorage.setItem('refreshToken', data.refresh_token);
  }

  return data;
};
```

### 2. Authenticated API Call
```javascript
// Make authenticated request
const makeAuthRequest = async (url, options = {}) => {
  const token = localStorage.getItem('accessToken');

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options.headers
    }
  });

  return response.json();
};

// Usage
const products = await makeAuthRequest('/v1/products');
```

### 3. File Upload Integration
```javascript
// Upload product image
const uploadProductImage = async (productId, imageFile) => {
  const token = localStorage.getItem('accessToken');
  const formData = new FormData();
  formData.append('image', imageFile);

  const response = await fetch(`/v1/products/${productId}/images`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  return response.json();
};
```

### 4. Order Creation
```javascript
const createOrder = async (orderData) => {
  return await makeAuthRequest('/v1/orders', {
    method: 'POST',
    body: JSON.stringify(orderData)
  });
};

// Usage
const order = await createOrder({
  product_id: 'product-uuid',
  quantity: 5,
  delivery_address: '123 Farm Road'
});
```

---

## Frontend Integration Best Practices

### 1. Authentication Management
- Store JWT tokens securely (localStorage/secure storage)
- Implement token refresh logic
- Handle token expiration gracefully
- Redirect to login when tokens are invalid

### 2. Error Handling
- Create centralized error handling
- Handle network errors and timeouts
- Parse and display API errors appropriately
- Implement retry logic for failed requests

### 3. Loading States
- Show loading indicators during API calls
- Handle pagination loading states
- Implement skeleton screens for better UX

### 4. Data Caching
- Cache user profile data
- Implement smart cache invalidation
- Store offline data where appropriate

### 5. File Upload Handling
- Validate file types and sizes
- Show upload progress
- Handle failed uploads gracefully
- Display uploaded images properly

### 6. Multi-Tenant Awareness
- Include tenant context in UI components
- Handle tenant-specific data isolation
- Implement tenant switching (if applicable)

### 7. API State Management
- Use a global state manager (Redux, Zustand, etc.)
- Centralize API calls in services
- Handle loading, error, and success states

---

## Common Data Models

### User
```typescript
interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  tenant_id: string;
  role: 'admin' | 'user' | 'manager';
  created_at: string;
  updated_at: string;
}
```

### Product
```typescript
interface Product {
  id: string;
  name: string;
  description?: string;
  price: number;
  category: {
    id: string;
    name: string;
  };
  tenant_id: string;
  created_at: string;
}
```

### Order
```typescript
interface Order {
  id: string;
  product: Product;
  quantity: number;
  total_amount: number;
  status: 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled';
  delivery_address: string;
  created_at: string;
}
```

### Invoice
```typescript
interface Invoice {
  id: string;
  order_id: string;
  total_amount: number;
  status: 'unpaid' | 'paid' | 'overdue';
  gstin?: string;
  issued_date: string;
}