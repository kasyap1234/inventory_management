package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"agromart2/internal/common"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type SearchRequest struct {
	Query   string                 `json:"query"`
	Filters map[string]interface{} `json:"filters"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
}

type SearchResult struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Relevance   float64                `json:"relevance"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Query   string         `json:"query"`
}

type SearchHandlers struct {
	db *pgxpool.Pool
}

func NewSearchHandlers(db *pgxpool.Pool) *SearchHandlers {
	return &SearchHandlers{db: db}
}

// sanitizeTsQuery escapes PostgreSQL full-text search special characters
// and converts input into a safe tsquery format
func sanitizeTsQuery(input string) string {
	// Characters that need escaping in PostgreSQL tsquery
	specialChars := []string{"&", "|", "!", "(", ")", ":", "'", "*", "<", ">"}

	sanitized := input
	for _, char := range specialChars {
		sanitized = strings.ReplaceAll(sanitized, char, " ")
	}

	// Split into words and join with & for AND logic
	words := strings.Fields(sanitized)
	if len(words) == 0 {
		return ""
	}

	// Filter out empty words and create tsquery terms
	var terms []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			terms = append(terms, word)
		}
	}

	return strings.Join(terms, " & ")
}

// UnifiedSearch performs full-text search across multiple entities
func (h *SearchHandlers) UnifiedSearch(c echo.Context) error {
	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant ID not found")
	}

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	results := []SearchResult{}

	// Search query - using PostgreSQL full-text search
	searchQuery := strings.TrimSpace(req.Query)

	if searchQuery == "" {
		return c.JSON(http.StatusOK, SearchResponse{
			Results: results,
			Total:   0,
			Query:   req.Query,
		})
	}

	// Prepare search term for PostgreSQL full-text search
	// Sanitize input to escape special PostgreSQL tsquery characters
	tsQuery := sanitizeTsQuery(searchQuery)

	// Search Products
	productResults, err := h.searchProducts(ctx, tenantID, tsQuery, req.Limit)
	if err == nil {
		results = append(results, productResults...)
	}

	// Search Orders
	orderResults, err := h.searchOrders(ctx, tenantID, tsQuery, req.Limit)
	if err == nil {
		results = append(results, orderResults...)
	}

	// Search Invoices
	invoiceResults, err := h.searchInvoices(ctx, tenantID, tsQuery, req.Limit)
	if err == nil {
		results = append(results, invoiceResults...)
	}

	// Search Categories
	categoryResults, err := h.searchCategories(ctx, tenantID, tsQuery, req.Limit)
	if err == nil {
		results = append(results, categoryResults...)
	}

	// Search Warehouses
	warehouseResults, err := h.searchWarehouses(ctx, tenantID, tsQuery, req.Limit)
	if err == nil {
		results = append(results, warehouseResults...)
	}

	// Limit total results
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}

	return c.JSON(http.StatusOK, SearchResponse{
		Results: results,
		Total:   len(results),
		Query:   req.Query,
	})
}

func (h *SearchHandlers) searchProducts(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]SearchResult, error) {
	sqlQuery := `
		SELECT 
			p.id,
			p.name,
			p.description,
			p.sku,
			p.price,
			c.name as category_name,
			ts_rank(
				to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '') || ' ' || coalesce(p.sku, '')),
				to_tsquery('english', $2)
			) as relevance
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.tenant_id = $1
		AND to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '') || ' ' || coalesce(p.sku, ''))
		@@ to_tsquery('english', $2)
		ORDER BY relevance DESC
		LIMIT $3
	`

	rows, err := h.db.Query(ctx, sqlQuery, tenantID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []SearchResult{}
	for rows.Next() {
		var (
			id           uuid.UUID
			name         string
			description  *string
			sku          *string
			price        float64
			categoryName *string
			relevance    float64
		)

		if err := rows.Scan(&id, &name, &description, &sku, &price, &categoryName, &relevance); err != nil {
			continue
		}

		desc := ""
		if description != nil {
			desc = *description
		}

		metadata := map[string]interface{}{
			"price": price,
		}
		if sku != nil {
			metadata["sku"] = *sku
		}
		if categoryName != nil {
			metadata["category"] = *categoryName
		}

		results = append(results, SearchResult{
			ID:          id.String(),
			Type:        "product",
			Title:       name,
			Description: desc,
			Metadata:    metadata,
			Relevance:   relevance,
		})
	}

	return results, nil
}

func (h *SearchHandlers) searchOrders(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]SearchResult, error) {
	sqlQuery := `
		SELECT 
			o.id,
			o.order_number,
			o.status,
			o.total_amount,
			o.created_at,
			ts_rank(
				to_tsvector('english', coalesce(o.order_number, '') || ' ' || coalesce(o.notes, '')),
				to_tsquery('english', $2)
			) as relevance
		FROM orders o
		WHERE o.tenant_id = $1
		AND to_tsvector('english', coalesce(o.order_number, '') || ' ' || coalesce(o.notes, ''))
		@@ to_tsquery('english', $2)
		ORDER BY relevance DESC
		LIMIT $3
	`

	rows, err := h.db.Query(ctx, sqlQuery, tenantID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []SearchResult{}
	for rows.Next() {
		var (
			id          uuid.UUID
			orderNumber string
			status      string
			totalAmount float64
			createdAt   string
			relevance   float64
		)

		if err := rows.Scan(&id, &orderNumber, &status, &totalAmount, &createdAt, &relevance); err != nil {
			continue
		}

		results = append(results, SearchResult{
			ID:          id.String(),
			Type:        "order",
			Title:       fmt.Sprintf("Order #%s", orderNumber),
			Description: fmt.Sprintf("Status: %s", status),
			Metadata: map[string]interface{}{
				"status":       status,
				"total_amount": totalAmount,
				"created_at":   createdAt,
			},
			Relevance: relevance,
		})
	}

	return results, nil
}

func (h *SearchHandlers) searchInvoices(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]SearchResult, error) {
	sqlQuery := `
		SELECT 
			i.id,
			i.invoice_number,
			i.status,
			i.total_amount,
			i.due_date,
			ts_rank(
				to_tsvector('english', coalesce(i.invoice_number, '') || ' ' || coalesce(i.notes, '')),
				to_tsquery('english', $2)
			) as relevance
		FROM invoices i
		WHERE i.tenant_id = $1
		AND to_tsvector('english', coalesce(i.invoice_number, '') || ' ' || coalesce(i.notes, ''))
		@@ to_tsquery('english', $2)
		ORDER BY relevance DESC
		LIMIT $3
	`

	rows, err := h.db.Query(ctx, sqlQuery, tenantID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []SearchResult{}
	for rows.Next() {
		var (
			id            uuid.UUID
			invoiceNumber string
			status        string
			totalAmount   float64
			dueDate       string
			relevance     float64
		)

		if err := rows.Scan(&id, &invoiceNumber, &status, &totalAmount, &dueDate, &relevance); err != nil {
			continue
		}

		results = append(results, SearchResult{
			ID:          id.String(),
			Type:        "invoice",
			Title:       fmt.Sprintf("Invoice #%s", invoiceNumber),
			Description: fmt.Sprintf("Status: %s", status),
			Metadata: map[string]interface{}{
				"status":       status,
				"total_amount": totalAmount,
				"due_date":     dueDate,
			},
			Relevance: relevance,
		})
	}

	return results, nil
}

func (h *SearchHandlers) searchCategories(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]SearchResult, error) {
	sqlQuery := `
		SELECT 
			c.id,
			c.name,
			c.description,
			ts_rank(
				to_tsvector('english', coalesce(c.name, '') || ' ' || coalesce(c.description, '')),
				to_tsquery('english', $2)
			) as relevance
		FROM categories c
		WHERE c.tenant_id = $1
		AND to_tsvector('english', coalesce(c.name, '') || ' ' || coalesce(c.description, ''))
		@@ to_tsquery('english', $2)
		ORDER BY relevance DESC
		LIMIT $3
	`

	rows, err := h.db.Query(ctx, sqlQuery, tenantID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []SearchResult{}
	for rows.Next() {
		var (
			id          uuid.UUID
			name        string
			description *string
			relevance   float64
		)

		if err := rows.Scan(&id, &name, &description, &relevance); err != nil {
			continue
		}

		desc := ""
		if description != nil {
			desc = *description
		}

		results = append(results, SearchResult{
			ID:          id.String(),
			Type:        "category",
			Title:       name,
			Description: desc,
			Metadata:    map[string]interface{}{},
			Relevance:   relevance,
		})
	}

	return results, nil
}

func (h *SearchHandlers) searchWarehouses(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]SearchResult, error) {
	sqlQuery := `
		SELECT 
			w.id,
			w.name,
			w.location,
			w.capacity,
			ts_rank(
				to_tsvector('english', coalesce(w.name, '') || ' ' || coalesce(w.location, '')),
				to_tsquery('english', $2)
			) as relevance
		FROM warehouses w
		WHERE w.tenant_id = $1
		AND to_tsvector('english', coalesce(w.name, '') || ' ' || coalesce(w.location, ''))
		@@ to_tsquery('english', $2)
		ORDER BY relevance DESC
		LIMIT $3
	`

	rows, err := h.db.Query(ctx, sqlQuery, tenantID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []SearchResult{}
	for rows.Next() {
		var (
			id        uuid.UUID
			name      string
			location  *string
			capacity  *int
			relevance float64
		)

		if err := rows.Scan(&id, &name, &location, &capacity, &relevance); err != nil {
			continue
		}

		loc := ""
		if location != nil {
			loc = *location
		}

		metadata := map[string]interface{}{}
		if capacity != nil {
			metadata["capacity"] = *capacity
		}

		results = append(results, SearchResult{
			ID:          id.String(),
			Type:        "warehouse",
			Title:       name,
			Description: loc,
			Metadata:    metadata,
			Relevance:   relevance,
		})
	}

	return results, nil
}
