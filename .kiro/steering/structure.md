# Project Structure

## Root Directory Layout

```
agromart2/
├── cmd/                    # Application entry points
├── internal/               # Private application code
├── pkg/                    # Public library code
├── frontend/               # Next.js React application
├── migrations/             # Database schema migrations
├── docs/                   # API documentation (Swagger/OpenAPI)
├── config/                 # Configuration files
├── k8s/                    # Kubernetes deployment manifests
├── tests/                  # Integration and performance tests
├── testhelpers/            # Test utilities and helpers
└── infrastructure/         # Infrastructure as code
```

## Backend Structure (internal/)

### Core Layers
- **handlers/**: HTTP request handlers (controllers)
- **services/**: Business logic layer
- **repositories/**: Data access layer
- **models/**: Domain models and entities
- **middleware/**: HTTP middleware (auth, RBAC, security)

### Supporting Modules
- **jobs/**: Background job processors
- **validation/**: Input validation and sanitization
- **security/**: Security utilities (CSRF, etc.)
- **analytics/**: Analytics and reporting services
- **caching/**: Cache service implementations
- **common/**: Shared utilities and helpers
- **config/**: Configuration management

## Frontend Structure

```
frontend/
├── app/                    # Next.js App Router
│   ├── dashboard/          # Dashboard pages
│   ├── login/              # Authentication pages
│   ├── components/         # Page-specific components
│   └── lib/                # Client-side utilities
├── components/             # Reusable UI components
│   ├── ui/                 # Base UI components
│   ├── auth/               # Authentication components
│   ├── dashboard/          # Dashboard-specific components
│   └── [feature]/          # Feature-specific components
├── hooks/                  # Custom React hooks
├── lib/                    # Utility functions and services
├── types/                  # TypeScript type definitions
└── public/                 # Static assets
```

## Naming Conventions

### Go Code
- **Files**: snake_case (e.g., `user_handlers.go`)
- **Packages**: lowercase, single word when possible
- **Types**: PascalCase (e.g., `UserRepository`)
- **Functions**: PascalCase for exported, camelCase for private
- **Constants**: UPPER_SNAKE_CASE

### Frontend Code
- **Files**: kebab-case for pages, PascalCase for components
- **Components**: PascalCase (e.g., `UserProfile.tsx`)
- **Hooks**: camelCase starting with "use" (e.g., `useAuth.ts`)
- **Types**: PascalCase (e.g., `UserData`)

## Key Patterns

### Backend Architecture
- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic encapsulation
- **Dependency Injection**: Constructor-based DI
- **Clean Architecture**: Separation of concerns
- **RBAC**: Role-based access control throughout

### Frontend Architecture
- **Component Composition**: Reusable UI building blocks
- **Custom Hooks**: Shared stateful logic
- **API Client**: Centralized HTTP communication
- **Type Safety**: Full TypeScript coverage

## File Organization Rules

1. **Group by Feature**: Related functionality stays together
2. **Separate Concerns**: Clear boundaries between layers
3. **Test Proximity**: Tests near the code they test
4. **Configuration Centralization**: All config in dedicated locations
5. **Documentation Co-location**: Docs near relevant code

## Import Conventions

### Go Imports
```go
// Standard library
import (
    "context"
    "fmt"
)

// Third-party packages
import (
    "github.com/labstack/echo/v4"
    "github.com/google/uuid"
)

// Internal packages
import (
    "agromart2/internal/models"
    "agromart2/internal/services"
)
```

### TypeScript Imports
```typescript
// React and Next.js
import React from 'react'
import { NextPage } from 'next'

// Third-party libraries
import axios from 'axios'
import { useQuery } from '@tanstack/react-query'

// Internal modules
import { Button } from '@/components/ui/button'
import { useAuth } from '@/hooks/useAuth'
```