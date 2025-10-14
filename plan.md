# Implementation Plan: Bulk Product Deletion API

This document outlines the plan to implement a bulk product deletion API endpoint.

## 1. Task Breakdown

### Task 1: Extend the `ProductService` Interface

- **File:** `internal/services/product_service.go`
- **Change:** Add a new method to the `ProductService` interface:
  ```go
  BulkDeleteProducts(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID) (*models.BulkOperationResult, error)
  ```

### Task 2: Implement the `BulkDeleteProducts` Method

- **File:** `internal/services/product_service.go`
- **Change:** Implement the `BulkDeleteProducts` method in the `productService` struct. This method will:
    - Iterate through the provided product IDs.
    - Call the `productRepo.Delete` method for each product ID.
    - Track successful and failed deletions.
    - Return a `BulkOperationResult` summarizing the outcome.

### Task 3: Extend the `ProductRepository` Interface

- **File:** `internal/repositories/product_repo.go`
- **Change:** Add a new method to the `ProductRepository` interface:
  ```go
  BulkDelete(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID) (int64, error)
  ```

### Task 4: Implement the `BulkDelete` Method

- **File:** `internal/repositories/product_repo.go`
- **Change:** Implement the `BulkDelete` method in the `productRepo` struct. This method will use a single `DELETE` statement with a `WHERE id = ANY($2)` clause to efficiently delete multiple products.

### Task 5: Create a New Handler for Bulk Deletion

- **File:** `internal/handlers/product_handlers.go`
- **Change:** Create a new handler function `BulkDeleteProducts` that:
    - Binds the request body to a struct containing the product IDs.
    - Validates the product IDs.
    - Calls the `productService.BulkDeleteProducts` method.
    - Returns an appropriate JSON response.

### Task 6: Register the New Route

- **File:** `cmd/main.go`
- **Change:** Register the new `DELETE /products/bulk/delete` route and associate it with the `BulkDeleteProducts` handler.

## 2. Mermaid Diagram

```mermaid
sequenceDiagram
    participant client as Client
    participant router as Echo Router
    participant handler as ProductHandlers
    participant service as ProductService
    participant repo as ProductRepository
    participant db as Database

    client->>router: DELETE /v1/products/bulk/delete
    router->>handler: BulkDeleteProducts(c)
    handler->>service: BulkDeleteProducts(ctx, tenantID, productIDs)
    service->>repo: BulkDelete(ctx, tenantID, productIDs)
    repo->>db: DELETE FROM products WHERE tenant_id = $1 AND id = ANY($2)
    db-->>repo: (int64, error)
    repo-->>service: (int64, error)
    service-->>handler: (*BulkOperationResult, error)
    handler-->>client: JSON Response
end
```

## 3. Approval

Please review this plan. Once you approve, I will switch to `code` mode and begin implementation.