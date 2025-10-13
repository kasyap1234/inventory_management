# Technical Debt Remediation Plan

## 1. Overview

This document outlines a strategic plan to address the technical debt identified in the "Feature Gap & Technical Debt Report." The goal is to systematically remediate incomplete features and architectural shortcomings to improve the platform's stability, scalability, security, and maintainability. This plan provides a detailed roadmap for developers to follow, ensuring that fixes are implemented consistently and effectively.

## 2. Guiding Principles

The remediation process will be guided by the following principles:

*   **Prioritize Security and Stability:** Address vulnerabilities and error handling first to protect user data and ensure system reliability.
*   **Enhance Observability:** Implement robust logging, monitoring, and error tracking to provide deep insights into the application's health and behavior.
*   **Improve Scalability:** Refactor components that are not designed for growth, ensuring the architecture can support an expanding user base and workload.
*   **Promote Maintainability:** Adopt consistent coding standards, improve documentation, and create clean, modular code that is easy to understand and modify.
*   **Configuration-Driven Functionality:** Avoid hardcoded values by using environment variables and tenant-specific settings to allow for flexible and dynamic configuration.

## 3. Remediation Roadmap

This section details the architectural changes required for each identified `TODO` item.

---

### **A. Centralized Error Tracking**

*   **`TODO`:** `// TODO: Log to error tracking service (Sentry, LogRocket, etc.)`
*   **Objective:** To implement a centralized error tracking system that captures, aggregates, and alerts on all application exceptions, providing a unified view of system health.
*   **Proposed Architecture:**
    *   Integrate a third-party error tracking service (e.g., Sentry) into both the Go backend and the Next.js frontend.
    *   **Backend:** Create a new error-handling middleware in Go that intercepts panics and critical errors. This middleware will capture the stack trace, request context, and authenticated user ID, then send this information to the error tracking service.
    *   **Frontend:** Initialize the error tracking client in a top-level component (e.g., `ErrorBoundary.tsx` or `app/layout.tsx`). The client will automatically capture unhandled exceptions and performance data.
*   **Key Components to be Modified:**
    *   `internal/handlers/error_handler.go`: Modify to include the error reporting logic.
    *   `cmd/main.go`: Add middleware initialization.
    *   `frontend/lib/error-tracking.ts`: Create a new file to encapsulate client-side error tracking initialization and configuration.
    *   `frontend/components/ErrorBoundary.tsx`: Wrap the application with this component to catch rendering errors.
*   **Configuration Changes:**
    *   `SENTRY_DSN`: The Data Source Name for the error tracking project.
    *   `SENTRY_ENVIRONMENT`: The deployment environment (e.g., `development`, `production`).
*   **Verification Strategy:**
    *   Create a test endpoint that deliberately triggers a panic and verify the error appears in the Sentry dashboard with the correct context.
    *   Trigger a client-side exception in the browser and confirm it is reported to Sentry.

---

### **B. Distributed Rate Limiting**

*   **`TODO`:** `// TODO: Use shared store (e.g., Redis) for cluster-wide enforcement`
*   **Objective:** To implement a robust, scalable rate-limiting mechanism that is enforced consistently across all instances of the application cluster.
*   **Proposed Architecture:**
    *   Introduce a Redis instance to the architecture, managed via `docker-compose.yml` for local development and provisioned separately in production.
    *   Modify the existing rate-limiting middleware (`internal/middleware/redis_rate_limiter.go`) to use Redis as a shared store for tracking request counts per IP address or user ID.
    *   Implement a fixed-window or sliding-window algorithm using Redis commands (`INCR`, `EXPIRE`) to manage rate-limiting counters.
*   **Key Components to be Modified:**
    *   `internal/middleware/redis_rate_limiter.go`: Replace the in-memory store with a Redis-backed implementation.
    *   `docker-compose.yml`: Add a Redis service.
    *   `config/config.go`: Add Redis connection settings.
*   **Configuration Changes:**
    *   `REDIS_URL`: The connection string for the Redis instance.
    *   `RATE_LIMIT_MAX_REQUESTS`: The number of allowed requests per window.
    *   `RATE_LIMIT_WINDOW_SECONDS`: The duration of the rate-limiting window in seconds.
*   **Verification Strategy:**
    *   Run a load test using a tool like `k6` or `siege` to send requests from a single IP address, confirming that the API returns `429 Too Many Requests` after the limit is exceeded.
    -   Deploy two instances of the application and verify that the rate limit is shared between them.

---

### **C. Dynamic Tenant-Specific Settings**

*   **`TODO`:** `// TODO: Fetch from tenant settings when available`
*   **Objective:** To centralize tenant-specific configurations in the database, allowing for dynamic customization without requiring code changes or deployments.
*   **Proposed Architecture:**
    *   Extend the `tenants` table in the database with a `settings` column of type `JSONB`.
    *   Create a caching layer (using Redis or an in-memory cache with TTL) to store tenant settings and avoid database lookups on every request.
    *   Modify the tenant service (`internal/services/tenant_service.go`) to include a function like `GetTenantSettings(tenantID)`. This function will first check the cache and fall back to the database if the settings are not found.
    *   Update relevant handlers to use these settings instead of hardcoded values.
*   **Key Components to be Modified:**
    *   `internal/models/tenant.go`: Add the `Settings` field to the `Tenant` struct.
    *   `migrations/`: Create a new migration to add the `settings` column to the `tenants` table.
    *   `internal/services/tenant_service.go`: Implement the logic to fetch and cache tenant settings.
    *   Various handlers (e.g., `internal/handlers/auth_handlers.go`) that may use tenant-specific logic.
*   **Configuration Changes:**
    *   None (settings will be managed via the application's UI or an admin panel).
*   **Verification Strategy:**
    *   Create a unit test for the `GetTenantSettings` function that verifies the caching logic.
    *   Manually update the `settings` for a test tenant in the database and confirm that the application's behavior changes accordingly without a restart.

---

### **D. API Response Field Filtering**

*   **`TODO`:** `// TODO: Implement reflection-based field filtering to exclude sensitive fields`
*   **Objective:** To prevent sensitive data leakage by implementing a fine-grained, dynamic field-filtering mechanism for API responses based on user permissions.
*   **Proposed Architecture:**
    *   Develop a utility function that uses Go's `reflect` package to dynamically inspect and modify structs before they are serialized to JSON.
    *   This function will accept a data struct and a list of allowed or disallowed fields based on the user's role and permissions.
    *   Field tags (e.g., `` `json:"password" security:"sensitive"` ``) will be used to mark fields that should be excluded by default unless explicitly permitted.
    *   Integrate this utility into a response-writing function used by API handlers.
*   **Key Components to be Modified:**
    *   `internal/security/field_filter.go`: Create a new file for the reflection-based filtering logic.
    *   `internal/handlers/`: Update handlers to use the new response-writing function.
    *   `internal/models/`: Add security-related field tags to structs (e.g., `User`, `Tenant`).
*   **Configuration Changes:**
    *   None.
*   **Verification Strategy:**
    *   Write a unit test for the field-filtering utility with various structs and permission levels.
    *   Create an integration test where two users with different roles request the same resource and verify that the sensitive fields are correctly filtered for the user with fewer privileges.

---

### **E. Multi-Channel Notification System**

*   **`TODO`:** `// TODO: Send notifications via email/SMS`
*   **Objective:** To implement a scalable, multi-channel notification system that can deliver alerts and messages via email, SMS, and potentially other channels (e.g., webhooks, push notifications).
*   **Proposed Architecture:**
    *   Design a generic `Notification` service interface that abstracts the details of the delivery channels.
    *   Create concrete implementations for each channel (e.g., `EmailSender`, `SmsSender`). Use third-party APIs like SendGrid for email and Twilio for SMS.
    *   Use a background job queue (e.g., Asynq) to process and send notifications asynchronously, improving API response times and handling retries.
    *   The `notification_service.go` will be responsible for dispatching a notification to the appropriate channel based on user preferences and notification type.
*   **Key Components to be Modified:**
    *   `internal/services/notification_service.go`: Refactor to support multiple channels and dispatch notifications to a background queue.
    *   `internal/jobs/notification_jobs.go`: Create a new file to define background jobs for sending emails and SMS.
    *   `internal/models/notification.go`: Update the model to include channel and status information.
*   **Configuration Changes:**
    *   `SENDGRID_API_KEY`: API key for the email service.
    *   `TWILIO_AUTH_TOKEN`: Auth token for the SMS service.
    *   `TWILIO_ACCOUNT_SID`: Account SID for the SMS service.
*   **Verification Strategy:**
    *   Trigger a notification and verify that a job is enqueued in the background queue.
    *   Confirm that the email/SMS is successfully delivered to a test recipient.
    -   Write a test to ensure that if a primary channel fails, the system retries or fails gracefully.

---

### **F. Filter Usage Tracking**

*   **`TODO`:** `// TODO: Implement proper filter usage tracking table`
*   **Objective:** To collect and analyze data on how users interact with filters, providing insights for feature development and performance optimization.
*   **Proposed Architecture:**
    *   Create a new database table named `filter_usage_logs` with columns such as `id`, `user_id`, `tenant_id`, `filter_name`, `filter_value`, `timestamp`.
    *   Create a new service, `AnalyticsService`, responsible for asynchronously writing filter usage data to this table. To minimize performance impact, this service should batch-write logs.
    *   Modify the API handlers that process filters to call the `AnalyticsService`.
*   **Key Components to be Modified:**
    *   `migrations/`: Add a new migration to create the `filter_usage_logs` table.
    *   `internal/services/analytics_service.go`: Create a new service to handle the logging.
    *   `internal/handlers/`: Update relevant handlers (e.g., `product_handlers.go`, `order_handlers.go`) to log filter usage.
*   **Configuration Changes:**
    *   None.
*   **Verification Strategy:**
    *   Make an API request with filter parameters and verify that a corresponding entry is created in the `filter_usage_logs` table.
    *   Perform a load test to ensure the asynchronous logging does not negatively impact API response times.

## 4. Prioritization

The remediation tasks are grouped into the following priorities:

### **Immediate Priority**
*   **Centralized Error Tracking:** Essential for system stability and proactive issue detection.
*   **Distributed Rate Limiting:** Critical for scalability and protecting the service from abuse.
*   **API Response Field Filtering:** A crucial security measure to prevent data leakage.
*   **Dynamic Tenant-Specific Settings:** Foundational for building a flexible, multi-tenant architecture.

### **Medium Priority**
*   **Multi-Channel Notification System:** Important for user engagement but can be deferred briefly. Asynchronous processing is the key focus.
*   **Filter Usage Tracking:** Provides valuable business intelligence but does not impact core functionality.

### **Low Priority**
*   **Authentication Features & Documentation Cleanup:** These items from the original report are important for the long term but can be addressed after the more critical architectural issues are resolved.

## 5. Summary

This plan provides a clear and actionable roadmap for addressing the most critical technical debt in the platform. By focusing on the prioritized areas, we can significantly improve the system's security, scalability, and observability. The proposed architectural changes are designed to be modular and maintainable, setting a strong foundation for future development. Execution of this plan will require a coordinated effort, and each remediation item should be thoroughly tested to ensure it meets the stated objectives.
