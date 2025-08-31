# Frontend Developer API Documentation

## Overview

This document provides comprehensive API documentation for the Agromart2 Multi-Tenant SaaS Inventory Management Platform. It covers all REST endpoints, authentication mechanisms, data models, and integration patterns required for frontend application development.

## Table of Contents

- [Authentication](#authentication)
- [Users](#users)
- [Products](#products)
- [Inventory](#inventory)
- [Orders](#orders)
- [Invoices](#invoices)
- [Categories](#categories)
- [Warehouses](#warehouses)
- [Suppliers & Distributors](#suppliers--distributors)
- [Analytics](#analytics)
- [Notifications](#notifications)
- [System](#system)
- [Error Handling](#error-handling)
- [Multi-Tenancy](#multi-tenancy)
- [Environment Setup](#environment-setup)

## Base URLs

- **Production**: `https://api.agromart.com/v1`
- **Staging**: `https://staging-api.agromart.com/v1`
- **Development**: `http://localhost:8080/v1`

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. All authenticated endpoints require a valid Bearer token in the Authorization header.

### Login

Logs in a user and returns JWT tokens.

**Endpoint:** `POST /auth/login`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Signup/Register

Creates a new user account.

**Endpoint:** `POST /auth/signup`

**Request:**
```json
{
  "email": "newuser@example.com",
  "password": "secure_password_123",
  "first_name": "Jane",
  "last_name": "Smith",
  "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"  // Optional, defaults to development tenant
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "456e7890-e89b-12d3-a456-426614174000",
  "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "user": {
    "id": "456e7890-e89b-12d3-a456-426614174000",
    "email": "newuser@example.com",
    "first_name": "Jane",
    "last_name": "Smith",
    "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "status": "active",
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

### Refresh Token

Refreshes access token using refresh token.

**Endpoint:** `POST /auth/refresh`

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "grant_type": "refresh_token"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Logout

Revokes JWT tokens.

**Endpoint:** `POST /auth/logout`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "token_type_hint": "access_token"  // Optional: "access_token" or "refresh_token"
}
```

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

### Get Current User Profile

Gets the current authenticated user's profile.

**Endpoint:** `GET /auth/me`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Users

### User Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "status": "string",  // "active", "inactive"
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

## Products

### Product Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "category_id": "uuid",  // Optional
  "name": "string",
  "batch_number": "string",  // Optional
  "expiry_date": "date",  // Optional
  "quantity": "integer",
  "unit_price": "number",
  "barcode": "string",  // Optional
  "unit_of_measure": "string",  // Optional
  "description": "string",  // Optional
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Create Product

**Endpoint:** `POST /products`
**Permission:** `products:create`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "name": "Premium Wheat Seeds",
  "category_id": "123e4567-e89b-12d3-a456-426614174001",  // Optional
  "batch_number": "BATCH001",
  "expiry_date": "2025-12-31",  // Optional
  "quantity": 100,
  "unit_price": 25.99,
  "barcode": "123456789012",  // Optional
  "unit_of_measure": "kg",  // Optional
  "description": "High-quality wheat seeds for agriculture"  // Optional
}
```

**Response:**
```json
{
  "message": "Product created successfully",
  "product": {
    "id": "456e7890-e89b-12d3-a456-426614174000",
    "tenant_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "name": "Premium Wheat Seeds",
    "category_id": "123e4567-e89b-12d3-a456-426614174001",
    "batch_number": "BATCH001",
    "expiry_date": "2025-12-31T00:00:00Z",
    "quantity": 100,
    "unit_price": 25.99,
    "barcode": "123456789012",
    "unit_of_measure": "kg",
    "description": "High-quality wheat seeds for agriculture",
    "created_at": "2024-01-15T12:00:00Z",
    "updated_at": "2024-01-15T12:00:00Z"
  }
}
```

### List Products

**Endpoint:** `GET /products`
**Permission:** `products:list`

**Query Parameters:**
- `limit`: Maximum number of products to return (default: 10, max: 100)
- `offset`: Number of products to skip (default: 0)

### Get Product by ID

**Endpoint:** `GET /products/{id}`
**Permission:** `products:read`

### Update Product

**Endpoint:** `PUT /products/{id}`
**Permission:** `products:update`

**Request:**
```json
{
  "name": "Premium Wheat Seeds - Updated",
  "quantity": 120,
  "unit_price": 26.99,
  "description": "High-quality wheat seeds with improved germination"
}
```

### Delete Product

**Endpoint:** `DELETE /products/{id}`
## Categories

### Category Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "string",
  "description": "string",  // Optional
  "parent_id": "uuid",  // Optional - for hierarchical categories
  "level": "integer",  // Hierarchy level
  "status": "string",  // "active", "inactive"
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### List Categories

**Endpoint:** `GET /categories`
**Permission:** `categories:list`

**Query Parameters:**
- `parent_id`: Filter by parent category (for hierarchical listing)
- `level`: Filter by hierarchy level
- `limit`: Results limit
- `offset`: Results offset

### Create Category

**Endpoint:** `POST /categories`
**Permission:** `categories:create`

**Request:**
```json
{
  "name": "Fertilizers",
  "description": "Agricultural fertilizers and nutrients",
  "parent_id": "123e4567-e89b-12d3-a456-426614174000"  // Optional
}
```

### Get Category by ID

**Endpoint:** `GET /categories/{id}`
**Permission:** `categories:read`

### Update Category

**Endpoint:** `PUT /categories/{id}`
**Permission:** `categories:update`

### Delete Category

**Endpoint:** `DELETE /categories/{id}`
**Permission:** `categories:delete`

### Get Category Hierarchy

**Endpoint:** `GET /categories/hierarchy`
**Permission:** `categories:list`

Returns complete hierarchical tree structure of categories.
**Permission:** `products:delete`

### Search Products

**Endpoint:** `GET /products/search`
**Permission:** `products:list`

**Query Parameters:**
- `q`: Search query (searches name, description, barcode)
- `category_id`: Filter by category UUID
- `limit`: Results limit
- `offset`: Results offset

### Product Analytics

**Endpoint:** `GET /products/analytics`
**Permission:** `products:read`

**Response:**
```json
{
  "analytics": {
    "Seeds": 15,
    "Fertilizers": 8,
    "Pesticides": 5,
    "Uncategorized": 3
  },
  "description": "Category distribution of products"
}
```

### Upload Product Image

**Endpoint:** `POST /products/{id}/images`
**Permission:** `products:update`

**Content-Type:** `multipart/form-data`

**Form Data:**
- `image`: Image file (JPEG, PNG, GIF, WebP, max 5MB)
- `alt_text`: Alternative text for the image (optional)

### Get Product Images

**Endpoint:** `GET /products/{id}/images`
**Permission:** `products:read`

### Get Product Image URL

**Endpoint:** `GET /products/{id}/images/{imageId}/url`
**Permission:** `products:read`

**Query Parameters:**
- `expiry_minutes`: URL expiry time (default: 1440 minutes = 24 hours)

### Delete Product Image

**Endpoint:** `DELETE /products/{id}/images/{imageId}`
**Permission:** `products:update`

## Inventory

### Inventory Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "warehouse_id": "uuid",
  "product_id": "uuid",
  "quantity": "integer",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### List Inventory

**Endpoint:** `GET /inventory`
**Permission:** `inventories:list`

**Query Parameters:**
- `limit`: Results limit (default: 10, max: 100)
- `offset`: Results offset (default: 0)

### Create/Update Inventory

**Endpoint:** `POST /inventory`
**Permission:** `inventories:create`

**Request:**
```json
{
  "warehouse_id": "123e4567-e89b-12d3-a456-426614174002",
  "product_id": "456e7890-e89b-12d3-a456-426614174001",
  "quantity": 100
}
```

### Get Inventory by ID

**Endpoint:** `GET /inventory/{id}`
**Permission:** `inventories:read`

### Update Inventory

**Endpoint:** `PUT /inventory/{id}`
**Permission:** `inventories:update`

### Delete Inventory

**Endpoint:** `DELETE /inventory/{id}`
**Permission:** `inventories:delete`

### Adjust Stock

**Endpoint:** `POST /inventory/adjust`
**Permission:** `inventories:update`

**Request:**
```json
{
  "warehouse_id": "123e4567-e89b-12d3-a456-426614174002",
  "product_id": "456e7890-e89b-12d3-a456-426614174001",
  "quantity_change": 50  // Positive for increase, negative for decrease
}
```

### Check Stock Availability

**Endpoint:** `POST /inventory/availability`
**Permission:** `inventories:read`

**Request:**
```json
{
  "warehouse_id": "123e4567-e89b-12d3-a456-426614174002",
  "product_id": "456e7890-e89b-12d3-a456-426614174001",
  "quantity": 20
}
```

**Response:**
```json
{
  "available": true,
  "requested": 20,
  "available_quantity": 150
}
```

### Transfer Stock

**Endpoint:** `POST /inventory/transfer`
**Permission:** `inventories:update`

**Request:**
```json
{
  "product_id": "456e7890-e89b-12d3-a456-426614174001",
  "from_warehouse_id": "123e4567-e89b-12d3-a456-426614174002",
  "to_warehouse_id": "789e0123-e89b-12d3-a456-426614174003",
  "quantity": 25
}
```

### Search Inventory

**Endpoint:** `GET /inventory/search`
**Permission:** `inventories:list`

Supports advanced filters for warehouse, product, quantity ranges, etc.

## Orders

### Order Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "order_type": "string",  // "purchase" or "sales"
  "supplier_id": "uuid",  // For purchase orders
  "distributor_id": "uuid",  // For sales orders
  "product_id": "uuid",
  "warehouse_id": "uuid",
  "quantity": "integer",
  "unit_price": "number",
  "total_amount": "number",
  "status": "string",  // "pending", "approved", "processing", "shipped", "delivered", "cancelled"
  "order_date": "datetime",
  "expected_delivery": "date",  // Optional
  "actual_delivery": "datetime",  // Optional
  "notes": "string",  // Optional
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Create Order

**Endpoint:** `POST /orders`

**Request:**
```json
{
  "order_type": "purchase",
  "supplier_id": "123e4567-e89b-12d3-a456-426614174004",
  "product_id": "456e7890-e89b-12d3-a456-426614174001",
  "warehouse_id": "789e0123-e89b-12d3-a456-426614174002",
  "quantity": 50,
  "unit_price": 25.99,
  "expected_delivery": "2024-02-01",
  "notes": "Urgent delivery required"
}
```

### List Orders

**Endpoint:** `GET /orders`

**Query Parameters:**
- `limit`: Results limit
- `offset`: Results offset

### Get Order by ID

**Endpoint:** `GET /orders/{id}`

### Order Analytics

**Endpoint:** `GET /orders/analytics`

**Query Parameters:**
- `start_date`: Analytics start date (default: 30 days ago)
- `end_date`: Analytics end date (default: today)

### Search Orders

**Endpoint:** `GET /orders/search`

**Query Parameters:**
- `status`: Filter by order status
- `limit`: Results limit
- `offset`: Results offset

### Approve Order

**Endpoint:** `POST /orders/{id}/approve`

### Process Order

**Endpoint:** `POST /orders/{id}/process`

### Ship Order

**Endpoint:** `POST /orders/{id}/ship`

**Request:**
```json
{
  "expected_delivery": "2024-02-05"  // Optional
}
```

### Deliver Order

**Endpoint:** `POST /orders/{id}/deliver`

### Cancel Order

**Endpoint:** `POST /orders/{id}/cancel`

### Get Order History

**Endpoint:** `GET /orders/{id}/history`

## Invoices

### Invoice Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "order_id": "uuid",
  "gstin": "string",  // GSTIN number
  "gst_rate": "number",  // GST rate (e.g., 18.0)
  "taxable_amount": "number",  // Amount before tax
  "cgst": "number",  // Central GST
  "sgst": "number",  // State GST
  "igst": "number",  // Inter-state GST (optional)
  "total_amount": "number",  // Total amount including tax
  "status": "string",  // "unpaid", "paid", "overdue"
  "issued_date": "datetime",
  "paid_date": "datetime",  // Optional
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Create Invoice

**Endpoint:** `POST /invoices`

**Request:**
```json
{
  "order_id": "123e4567-e89b-12d3-a456-426614174005",
  "gstin": "22AAAAA0000A1Z5"  // Optional
}
```

### List Invoices

**Endpoint:** `GET /invoices`

### Get Invoice by ID

**Endpoint:** `GET /invoices/{id}`

### Update Invoice Status

**Endpoint:** `PUT /invoices/{id}/status`

**Request:**
```json
{
  "status": "paid"  // "unpaid", "paid", "overdue"
}
```

### Get Unpaid Invoices

**Endpoint:** `GET /invoices/unpaid`

### Generate Invoice PDF

Generates a professional PDF invoice for the specified invoice ID and stores it in cloud storage for download. The PDF includes complete invoice details with GST calculations, itemized billing, and company information.

**Endpoint:** `POST /v1/invoices/{id}/generate-pdf`

**Authentication:** Required
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (string, required): UUID of the invoice to generate PDF for

**Request Body:** None required

**Response:**
```json
{
  "message": "PDF generated and uploaded successfully",
  "pdf_url": "https://storage.agromart.com/invoices/tenant123-invoice456.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&...",
  "expires_in": "24 hours"
}
```

**Success Response (200):**
- `message`: Confirmation message
- `pdf_url`: Presigned URL for PDF download (valid for 24 hours)
- `expires_in`: URL expiration time

**Error Responses:**

**400 Bad Request:**
```json
{
  "message": "Invalid invoice ID"
}
```
*Cause:* Invalid UUID format in request path

**401 Unauthorized:**
```json
{
  "message": "Tenant not found"
}
```
*Cause:* Missing or invalid JWT token, unable to extract tenant context

**404 Not Found:**
```json
{
  "message": "Invoice not found"
}
```
*Cause:* Invoice with specified ID doesn't exist or belongs to different tenant

**404 Not Found:**
```json
{
  "message": "Order not found for this invoice"
}
```
*Cause:* Associated order record not found

**500 Internal Server Error:**
```json
{
  "message": "Failed to generate PDF: <specific_error>"
}
```
*Cause:* PDF generation failed due to template or data issues

**500 Internal Server Error:**
```json
{
  "message": "Failed to upload PDF to storage"
}
```
*Cause:* MinIO storage upload failure

**500 Internal Server Error:**
```json
{
  "message": "Failed to generate download URL"
}
```
*Cause:* Presigned URL generation failure

**PDF Content Structure:**

The generated PDF includes the following sections:

1. **Company Header**
   - Company name: "AGROMART INVOICE"
   - Professional layout with company branding

2. **Invoice Details**
   - Invoice number (UUID)
   - Invoice date
   - Order ID reference
   - GSTIN (if provided)

3. **Billing Information**
   - Bill-to: Customer details (configurable)
   - Company contact information

4. **Item Table**
   - Description (product name + optional additional details)
   - Quantity ordered
   - Unit price
   - Line total amount

5. **GST Calculations**
   - Subtotal (pre-tax amount)
   - CGST (9% of taxable amount)
   - SGST (9% of taxable amount)
   - IGST (if applicable for inter-state transactions)
   - Total amount including all taxes

6. **Terms & Conditions**
   - Payment terms (30 days)
   - Late payment charges
   - Goods return policy
   - Computer generated document notice

7. **Footer**
   - Thank you message
   - Support contact information

**Integration Examples:**

**JavaScript/TypeScript:**
```javascript
// Generate PDF for invoice
async function generateInvoicePDF(invoiceId) {
  const response = await fetch(`/v1/invoices/${invoiceId}/generate-pdf`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${getAuthToken()}`,
      'Content-Type': 'application/json'
    }
  });

  if (response.ok) {
    const result = await response.json();
    console.log('PDF generated:', result.pdf_url);

    // Automatically download the PDF
    window.open(result.pdf_url, '_blank');

    return result;
  } else {
    const error = await response.json();
    throw new Error(error.message);
  }
}

// Usage
generateInvoicePDF('123e4567-e89b-12d3-a456-426614174000')
  .then(result => {
    // Handle success - PDF available at result.pdf_url
  })
  .catch(error => {
    // Handle error
    console.error('PDF generation failed:', error.message);
  });
```

**React Integration:**
```javascript
import React, { useState } from 'react';

function InvoicePDFButton({ invoiceId }) {
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState(null);

  const handleGeneratePDF = async () => {
    setIsGenerating(true);
    setError(null);

    try {
      const response = await fetch(`/v1/invoices/${invoiceId}/generate-pdf`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message);
      }

      const data = await response.json();

      // Open PDF in new tab
      window.open(data.pdf_url, '_blank');

      // Show success message
      alert('Invoice PDF generated successfully!');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div>
      <button
        onClick={handleGeneratePDF}
        disabled={isGenerating}
        className="pdf-generate-btn"
      >
        {isGenerating ? 'Generating PDF...' : 'Generate PDF'}
      </button>
      {error && <p className="error-message">Error: {error}</p>}
    </div>
  );
}

export default InvoicePDFButton;
```

**Python Integration:**
```python
import requests

def generate_invoice_pdf(invoice_id, auth_token):
    """
    Generate PDF for invoice and return download URL

    Args:
        invoice_id (str): UUID of the invoice
        auth_token (str): JWT authentication token

    Returns:
        dict: Response containing pdf_url and metadata
    """
    url = f"https://api.agromart.com/v1/invoices/{invoice_id}/generate-pdf"
    headers = {
        'Authorization': f'Bearer {auth_token}',
        'Content-Type': 'application/json'
    }

    response = requests.post(url, headers=headers)

    if response.status_code == 200:
        return response.json()
    else:
        # Parse error response
        error_data = response.json()
        raise Exception(f"PDF generation failed: {error_data.get('message')}")

# Usage example
try:
    result = generate_invoice_pdf(
        invoice_id="123e4567-e89b-12d3-a456-426614174000",
        auth_token="your_jwt_token_here"
    )
    print(f"PDF generated successfully: {result['pdf_url']}")
    print(f"Expires in: {result['expires_in']}")
except Exception as e:
    print(f"Error: {e}")
```

**Testing the Endpoint:**
```javascript
// Test endpoint functionality
async function testPDFGeneration() {
  // Assume you have valid token and invoice ID
  const invoiceId = '123e4567-e89b-12d3-a456-426614174000';
  const authToken = 'your_jwt_token_here';

  try {
    const response = await fetch(`/v1/invoices/${invoiceId}/generate-pdf`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      }
    });

    if (response.ok) {
      const data = await response.json();
      console.log('✅ PDF Generation Successful');
      console.log('PDF URL:', data.pdf_url);
      console.log('Expires In:', data.expires_in);

      // Test if URL is accessible
      const downloadResponse = await fetch(data.pdf_url);
      if (downloadResponse.ok) {
        console.log('✅ PDF Download URL is valid');
      } else {
        console.log('❌ PDF Download URL failed');
      }
    } else {
      console.log('❌ PDF Generation Failed');
      const errorData = await response.json();
      console.log('Error:', errorData.message);
    }
  } catch (error) {
    console.log('❌ Network or other error:', error.message);
  }
}

// Test with invalid invoice ID
async function testInvalidInvoice() {
  const invalidId = 'invalid-uuid';
  // ... similar to above
}

// Test with unauthorized access
async function testUnauthorized() {
  const invoiceId = '123e4567-e89b-12d3-a456-426614174000';
  const invalidToken = 'invalid_token';

  // ... similar structure
}
```

**Security Considerations:**

1. **Authentication Required**: All PDF generation requests require valid JWT authentication
2. **Tenant Isolation**: Users can only access PDFs for invoices within their own tenant
3. **URL Expiration**: Generated download URLs expire after 24 hours for security
4. **Access Control**: Endpoint respects role-based permissions for invoice access
5. **Audit Logging**: All PDF generation requests are logged for compliance

**Rate Limiting:**

The PDF generation endpoint is subject to standard API rate limiting:
- 1000 requests per hour per user
- PDF generation uses additional computational resources
- Consider implementing client-side caching of generated PDFs

**Best Practices:**

1. **Caching**: Cache generated PDFs on the client side to avoid repeated generations
2. **Error Handling**: Implement robust error handling for all failure scenarios
3. **User Feedback**: Provide clear loading states and success/error messages
4. **URL Management**: Track and refresh expired download URLs appropriately
5. **Performance**: Generate PDFs asynchronously for large invoice sets if needed

## Error Handling

### Common Error Response Format

```json
{
  "message": "Error description"
}
```

### HTTP Status Codes

- **200**: Success
- **201**: Created
- **400**: Bad Request (validation errors)
- **401**: Unauthorized (missing/invalid token)
- **403**: Forbidden (insufficient permissions)
- **404**: Not Found
- **409**: Conflict (duplicate data)
- **422**: Unprocessable Entity (validation failed)
- **500**: Internal Server Error
- **503**: Service Unavailable (dependencies down)
### Search Categories

**Endpoint:** `GET /categories/search`
**Permission:** `categories:list`

**Query Parameters:**
- `name`: Search by category name
- `limit`: Results limit
- `offset`: Results offset

### Update Category

**Endpoint:** `PUT /categories/{id}`
**Permission:** `categories:update`

### Delete Category

**Endpoint:** `DELETE /categories/{id}`
**Permission:** `categories:delete`

## Warehouses

### Warehouse Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "string",
  "location": "string",
  "capacity": "integer",  // Optional
  "status": "string",  // "active", "inactive"
  "manager_id": "uuid",  // Optional
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### List Warehouses

**Endpoint:** `GET /warehouses`
**Permission:** `warehouses:list`

### Create Warehouse

**Endpoint:** `POST /warehouses`
**Permission:** `warehouses:create`

### Get Warehouse by ID

**Endpoint:** `GET /warehouses/{id}`
**Permission:** `warehouses:read`

### Update Warehouse

**Endpoint:** `PUT /warehouses/{id}`
**Permission:** `warehouses:update`

### Delete Warehouse

**Endpoint:** `DELETE /warehouses/{id}`
**Permission:** `warehouses:delete`

## Suppliers & Distributors

### Supplier Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "string",
  "contact_person": "string",
  "email": "string",
  "phone": "string",
  "address": "string",
  "gstin": "string",  // Optional
  "payment_terms": "string",  // Optional
  "status": "string",  // "active", "inactive"
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Distributor Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "string",
  "contact_person": "string",
  "email": "string",
  "phone": "string",
  "region": "string",  // Covered region
  "discount_rate": "number",  // Optional
  "status": "string",  // "active", "inactive"
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Supplier Endpoints

- `GET /suppliers` - List suppliers
- `POST /suppliers` - Create supplier
- `GET /suppliers/{id}` - Get supplier by ID
- `PUT /suppliers/{id}` - Update supplier
- `DELETE /suppliers/{id}` - Delete supplier

### Distributor Endpoints

- `GET /distributors` - List distributors
- `POST /distributors` - Create distributor
- `GET /distributors/{id}` - Get distributor by ID
- `PUT /distributors/{id}` - Update distributor
- `DELETE /distributors/{id}` - Delete distributor

## Analytics

### System Analytics

**Endpoint:** `GET /analytics/system`

Provides overview statistics:
```json
{
  "total_products": 154,
  "total_orders": 89,
  "total_revenue": 25890.50,
  "active_inventory_value": 18752.30,
  "low_stock_products": 7,
  "pending_orders": 12
}
```

### Dashboard Analytics

**Endpoint:** `GET /analytics/dashboard`

Provides dashboard metrics for frontend display.

### Revenue Reports

**Endpoint:** `GET /analytics/revenue`

**Query Parameters:**
- `start_date`: Report start date
- `end_date`: Report end date
- `group_by`: "daily", "weekly", "monthly"

### Inventory Reports

**Endpoint:** `GET /analytics/inventory`

Provides detailed inventory analytics including low stock alerts and turnover ratios.

## Notifications

### Notification Data Model

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "user_id": "uuid",
  "type": "string",  // "info", "warning", "error"
  "title": "string",
  "message": "string",
  "read": "boolean",
  "created_at": "datetime"
}
```

### List Notifications

**Endpoint:** `GET /notifications`

**Query Parameters:**
- `limit`: Results limit
- `offset`: Results offset
- `unread_only`: Filter for unread notifications only

### Mark as Read

**Endpoint:** `PUT /notifications/{id}/read`

### Mark All as Read

**Endpoint:** `PUT /notifications/read-all`

### Create Notification

**Endpoint:** `POST /notifications`

Used for sending notifications to users.

## System

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "storage": "healthy"
  }
}
```

### Version Information

**Endpoint:** `GET /version`

Provides API version and build information.

## Multi-Tenancy

### Tenant Context

All API endpoints automatically filter data by the authenticated user's tenant ID. The tenant ID is extracted from the JWT token and used for:

- **Data Isolation**: Each tenant sees only their own data
- **Resource Limits**: API responses respect tenant-specific limits
- **Audit Logging**: All operations are logged per tenant

### Example Tenant Context Flow

1. User authenticates with `/auth/login`
2. JWT token includes both user ID and tenant ID
3. All subsequent API calls include the token
4. System automatically filters responses by tenant ID
5. Audit logs capture tenant-specific operations

### Multi-Tenant Data Models

All entity models include a `tenant_id` field for proper data isolation:

```json
{
  "id": "uuid",
  "tenant_id": "uuid",  // Ensures data belongs to specific tenant
  "name": "string",
  // ... other fields
}
```

## Environment Setup

### Development Environment

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd agromart2
   ```

2. **Copy environment configuration:**
   ```bash
   cp .env.example .env
   ```

3. **Update environment variables:**
   ```bash
   # Database Configuration
   DATABASE_URL=postgresql://user:password@localhost:5432/agromart_dev

   # Redis Configuration
   REDIS_URL=redis://localhost:6379

   # JWT Configuration
   JWT_SECRET=your-development-jwt-secret

   # Server Configuration
   PORT=8080
   ```

4. **Start dependencies:**
   ```bash
   # Using Docker Compose
   docker-compose up -d postgres redis minio

   # Or start locally if preferred
   ```

5. **Run the application:**
   ```bash
   go run cmd/main.go
   ```

6. **Access Swagger documentation:**
   - OpenAPI spec: `http://localhost:8080/swagger/index.html`
   - Frontend documentation: `docs/frontend-developer-api.md`

### Integration Patterns

#### Authentication Flow Pattern

```javascript
// Frontend integration example
async function login(email, password) {
  const response = await fetch('/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ email, password })
  });

  if (response.ok) {
    const data = await response.json();
    // Store tokens securely (localStorage, secure cookie, etc.)
    localStorage.setItem('accessToken', data.access_token);
    localStorage.setItem('refreshToken', data.refresh_token);
    // Set user context
    setUser(data.user);
    return data;
  } else {
    throw new Error('Login failed');
  }
}
```

#### Token Refresh Pattern

```javascript
async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) {
    throw new Error('No refresh token');
  }

  const response = await fetch('/v1/auth/refresh', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      refresh_token: refreshToken,
      grant_type: 'refresh_token'
    })
  });

  if (response.ok) {
    const data = await response.json();
    localStorage.setItem('accessToken', data.access_token);
    localStorage.setItem('refreshToken', data.refresh_token);
    return data;
  } else {
    // Handle refresh failure - redirect to login
    logout();
    throw new Error('Token refresh failed');
  }
}
```

#### API Request Wrapper Pattern

```javascript
async function apiRequest(url, options = {}) {
  const accessToken = localStorage.getItem('accessToken');

  const defaultOptions = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`
    }
  };

  const mergedOptions = { ...defaultOptions, ...options };

  // Merge headers properly
  if (options.headers) {
    mergedOptions.headers = { ...defaultOptions.headers, ...options.headers };
  }

  let response = await fetch(url, mergedOptions);

  if (response.status === 401) {
    // Try to refresh token automatically
    try {
      await refreshAccessToken();
      const newToken = localStorage.getItem('accessToken');
      mergedOptions.headers.Authorization = `Bearer ${newToken}`;
      response = await fetch(url, mergedOptions);
    } catch (refreshError) {
      // Refresh failed, redirect to login
      logout();
      throw new Error('Authentication required');
    }
  }

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.message || 'API request failed');
  }

  return response.json();
}
```

#### Error Handling Pattern

```javascript
async function handleApiError(error) {
  // Handle specific error types
  if (error.message.includes('401') || error.message.includes('Authentication')) {
    // Redirect to login
    redirectToLogin();
  } else if (error.message.includes('403')) {
    // Show permission error
    showError('You do not have permission to perform this action');
  } else if (error.message.includes('422') || error.message.includes('400')) {
    // Show validation errors
    showValidationErrors(error.details || []);
  } else if (error.message.includes('500')) {
    // Show generic error
    showError('A server error occurred. Please try again later.');
  } else {
    // Show generic error
    showError('An unexpected error occurred');
  }
}
```

### SDK Pattern Example

Create a client SDK for easier integration:

```javascript
// agromart-sdk.js
class AgromartAPI {
  constructor(baseURL = '/v1') {
    this.baseURL = baseURL;
    this.refreshPromise = null;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const response = await apiRequest(url, options);
    return response;
  }

  // Authentication methods
  async login(email, password) {
    return this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
  }

  async logout() {
    return this.request('/auth/logout', {
      method: 'POST'
    });
  }

  // Product methods
  async getProducts(limit = 10, offset = 0) {
    return this.request(`/products?limit=${limit}&offset=${offset}`);
  }

  async createProduct(productData) {
    return this.request('/products', {
      method: 'POST',
      body: JSON.stringify(productData)
    });
  }

  // Order methods
  async getOrders(limit = 10, offset = 0) {
    return this.request(`/orders?limit=${limit}&offset=${offset}`);
  }

  async createOrder(orderData) {
    return this.request('/orders', {
      method: 'POST',
      body: JSON.stringify(orderData)
    });
  }

// Invoice methods
  async getInvoices(limit = 10, offset = 0) {
    return this.request(`/invoices?limit=${limit}&offset=${offset}`);
  }

  async getUnpaidInvoices() {
    return this.request('/invoices/unpaid');
  }
}

// Export for use
export default AgromartAPI;
```

### React Integration Example

```javascript
// App.js
import React from 'react';
import { AgromartProvider, useAgromart } from './hooks/useAgromart';

function App() {
  return (
    <AgromartProvider apiBaseUrl="http://localhost:8080/v1">
      <Dashboard />
    </AgromartProvider>
  );
}

// Dashboard.js
import React, { useEffect, useState } from 'react';
import { useAgromart } from './hooks/useAgromart';

function Dashboard() {
  const { api, user, logout } = useAgromart();
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const data = await api.getProducts(20);
        setProducts(data.products);
      } catch (error) {
        console.error('Failed to fetch products:', error);
      } finally {
        setLoading(false);
      }
    };

    if (user) {
      fetchProducts();
    }
  }, [api, user]);

  if (!user) return <Login />;

  return (
    <div>
      <nav>
        <h1>Agromart Dashboard</h1>
        <button onClick={logout}>Logout</button>
      </nav>

      <div className="dashboard-content">
        <h2>Products ({products.length})</h2>
        {loading ? (
          <p>Loading...</p>
        ) : (
          <div className="product-list">
            {products.map(product => (
              <div key={product.id} className="product-card">
                <h3>{product.name}</h3>
                <p>Quantity: {product.quantity}</p>
                <p>Price: ${product.unit_price}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default Dashboard;
```

### Testing Integration

#### Authentication Testing
```javascript
// Login test
const testLogin = async () => {
  try {
    const response = await fetch('/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: 'test@example.com',
        password: 'test_password'
      })
    });
    const data = await response.json();
    console.log('Login successful:', data);
    return data.access_token;
  } catch (error) {
    console.error('Login failed:', error);
  }
};
```

#### API Testing Script
```javascript
// test-api-integration.js
async function testAPIEndpoints() {
  const token = await testLogin();

  if (!token) {
    console.error('Failed to authenticate');
    return;
  }

  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };

  // Test product endpoints
  console.log('Testing Product Endpoints:');

  // Get products
  try {
    const productsResponse = await fetch('/v1/products?limit=5', { headers });
    const productsData = await productsResponse.json();
    console.log('Products listing:', productsData);
  } catch (error) {
    console.error('Failed to fetch products:', error);
  }

  // Test order endpoints
  console.log('Testing Order Endpoints:');

  // Get orders
  try {
    const ordersResponse = await fetch('/v1/orders?limit=5', { headers });
    const ordersData = await ordersResponse.json();
    console.log('Orders listing:', ordersData);
  } catch (error) {
    console.error('Failed to fetch orders:', error);
  }

  // Test invoice endpoints
  console.log('Testing Invoice Endpoints:');

  // Get unpaid invoices
  try {
    const invoicesResponse = await fetch('/v1/invoices/unpaid', { headers });
    const invoicesData = await invoicesResponse.json();
    console.log('Unpaid invoices:', invoicesData);
  } catch (error) {
    console.error('Failed to fetch invoices:', error);
  }
}

// Run the tests
testAPIEndpoints();
```

## Common Issues and Solutions

### Authentication Issues

1. **Expired Token**
   - **Symptom**: 401 responses on valid requests
   - **Solution**: Implement automatic token refresh

2. **Invalid Tenant Context**
   - **Symptom**: 404 responses or empty data sets
   - **Solution**: Verify JWT token contains correct tenant_id

### Rate Limiting

The API implements rate limiting to prevent abuse:
- Standard requests: 1000 per hour per user
- Authentication endpoints: 100 per hour per IP

### Performance Optimization

1. **Pagination**: Always use limit and offset for large datasets
2. **Filtering**: Use search endpoints to reduce data transfer
3. **Caching**: Implement client-side caching for static data
4. **Image Optimization**: Use presigned URLs for images instead of direct data

### CORS Configuration

For frontend development, ensure CORS is configured:

```javascript
// Example CORS configuration for development
const corsOptions = {
  origin: [
    'http://localhost:3000',    // React dev server
    'http://localhost:4000',    // Vue dev server
    'http://localhost:8080'     // Angular dev server
  ],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization']
};
```

## Change Log

### Version 1.0.0
- Initial release with core inventory management features
- Multi-tenant architecture
- JWT authentication
- Complete CRUD operations for all entities
- Analytics and reporting endpoints
- Image upload and management
- PDF generation for invoices