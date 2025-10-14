# Permissions Definition

This document defines the permissions for all resources in the application.

## 1. Products

-   **`products:create`**: Allows creating a new product.
-   **`products:read`**: Allows viewing a list of products and individual product details.
-   **`products:update`**: Allows updating product information.
-   **`products:delete`**: Allows deleting a product.
-   **`products:publish`**: Allows changing a product's status to "published".
-   **`products:unpublish`**: Allows changing a product's status to "draft" or "archived".

## 2. Orders

-   **`orders:create`**: Allows creating a new order.
-   **`orders:read`**: Allows viewing a list of orders and individual order details.
-   **`orders:update`**: Allows updating order information (e.g., status, shipping address).
-   **`orders:delete`**: Allows canceling or deleting an order.
-   **`orders:process`**: Allows processing an order (e.g., fulfilling items).

## 3. Users

-   **`users:create`**: Allows inviting or creating a new user.
-   **`users:read`**: Allows viewing a list of users and their profiles.
-   **`users:update`**: Allows updating a user's profile information.
-   **`users:delete`**: Allows deactivating or deleting a user.
-   **`users:assign_roles`**: Allows assigning or unassigning roles to a user.

## 4. Roles

-   **`roles:create`**: Allows creating a new role.
-   **`roles:read`**: Allows viewing a list of roles and their permissions.
-   **`roles:update`**: Allows updating a role's name, description, and assigned permissions.
-   **`roles:delete`**: Allows deleting a role.

## 5. Permissions

-   **`permissions:read`**: Allows viewing the list of all available permissions in the system.

## 6. Categories

-   **`categories:create`**: Allows creating a new product category.
-   **`categories:read`**: Allows viewing the category hierarchy.
-   **`categories:update`**: Allows updating category details (e.g., name, parent category).
-   **`categories:delete`**: Allows deleting a category.

## 7. Warehouses

-   **`warehouses:create`**: Allows creating a new warehouse.
-   **`warehouses:read`**: Allows viewing a list of warehouses and their details.
-   **`warehouses:update`**: Allows updating warehouse information.
-   **`warehouses:delete`**: Allows deleting a warehouse.

## 8. Distributors

-   **`distributors:create`**: Allows creating a new distributor.
-   **`distributors:read`**: Allows viewing a list of distributors.
-   **`distributors:update`**: Allows updating distributor information.
-   **`distributors:delete`**: Allows deleting a distributor.

## 9. Suppliers

-   **`suppliers:create`**: Allows creating a new supplier.
-   **`suppliers:read`**: Allows viewing a list of suppliers.
-   **`suppliers:update`**: Allows updating supplier information.
-   **`suppliers:delete`**: Allows deleting a supplier.

## 10. Invoices

-   **`invoices:create`**: Allows creating a new invoice.
-   **`invoices:read`**: Allows viewing a list of invoices and individual invoice details.
-   **`invoices:update`**: Allows updating an invoice.
-   **`invoices:delete`**: Allows deleting an invoice.
-   **`invoices:send`**: Allows sending an invoice to a customer.

## 11. Subscriptions

-   **`subscriptions:create`**: Allows creating a new subscription for a user.
-   **`subscriptions:read`**: Allows viewing subscription details.
-   **`subscriptions:update`**: Allows updating a subscription (e.g., plan, status).
-   **`subscriptions:delete`**: Allows canceling a subscription.

## 12. Webhooks

-   **`webhooks:create`**: Allows creating a new webhook.
-   **`webhooks:read`**: Allows viewing a list of webhooks and their configurations.
-   **`webhooks:update`**: Allows updating a webhook's URL, events, and status.
-   **`webhooks:delete`**: Allows deleting a webhook.
-   **`webhooks:read_logs`**: Allows viewing the delivery logs for a webhook.

## 13. Analytics

-   **`analytics:read`**: Allows viewing the analytics dashboard.
-   **`analytics:export`**: Allows exporting analytics data.

## 14. Audit Logs

-   **`auditlogs:read`**: Allows viewing the audit trail of user actions.
-   **`auditlogs:export`**: Allows exporting audit logs.

## 15. Notifications

-   **`notifications:read`**: Allows viewing system notifications.
-   **`notifications:send`**: Allows sending system-wide or targeted notifications.

## 16. Jobs

-   **`jobs:read`**: Allows viewing the status of background jobs.
-   **`jobs:manage`**: Allows pausing, resuming, or re-running jobs.

## 17. Settings

-   **`settings:read`**: Allows viewing application settings.
-   **`settings:update`**: Allows updating application settings.

## 18. Tenants

-   **`tenants:create`**: Allows creating a new tenant.
-   **`tenants:read`**: Allows viewing tenant information.
-   **`tenants:update`**: Allows updating tenant settings.
-   **`tenants:delete`**: Allows deleting a tenant.