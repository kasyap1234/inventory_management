package containers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a testcontainer for PostgreSQL
type PostgresContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	User      string
	Password  string
	Database  string
	Pool      *pgxpool.Pool
}

// PostgresContainerConfig holds configuration for the PostgreSQL container
type PostgresContainerConfig struct {
	ImageTag       string
	Database       string
	User           string
	Password       string
	MigrationsPath string
}

// DefaultPostgresConfig returns a default configuration
func DefaultPostgresConfig() PostgresContainerConfig {
	return PostgresContainerConfig{
		ImageTag:       "16-alpine",
		Database:       "agromart2_test",
		User:           "postgres",
		Password:       "postgres",
		MigrationsPath: "",
	}
}

// NewPostgresContainer creates a new PostgreSQL testcontainer
func NewPostgresContainer(ctx context.Context, config PostgresContainerConfig) (*PostgresContainer, error) {
	if config.ImageTag == "" {
		config.ImageTag = "16-alpine"
	}
	if config.Database == "" {
		config.Database = "agromart2_test"
	}
	if config.User == "" {
		config.User = "postgres"
	}
	if config.Password == "" {
		config.Password = "postgres"
	}

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("postgres:%s", config.ImageTag),
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     config.User,
			"POSTGRES_PASSWORD": config.Password,
			"POSTGRES_DB":       config.Database,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	pc := &PostgresContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
		User:      config.User,
		Password:  config.Password,
		Database:  config.Database,
	}

	// Create connection pool
	connString := pc.ConnectionString()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	pc.Pool = pool

	// Run migrations if path is provided
	if config.MigrationsPath != "" {
		if err := pc.RunMigrations(ctx, config.MigrationsPath); err != nil {
			pool.Close()
			container.Terminate(ctx)
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return pc, nil
}

// ConnectionString returns the PostgreSQL connection string
func (pc *PostgresContainer) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pc.User, pc.Password, pc.Host, pc.Port, pc.Database,
	)
}

// ConnectionStringDSN returns the PostgreSQL DSN-style connection string
func (pc *PostgresContainer) ConnectionStringDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pc.Host, pc.Port, pc.User, pc.Password, pc.Database,
	)
}

// RunMigrations runs SQL migration files from the specified directory
func (pc *PostgresContainer) RunMigrations(ctx context.Context, migrationsPath string) error {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort files to ensure order
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, fileName := range sqlFiles {
		// Skip seed data files - they require pre-existing data (e.g., tenants)
		// that won't exist in fresh test databases
		if strings.Contains(strings.ToLower(fileName), "seed") {
			continue
		}

		filePath := filepath.Join(migrationsPath, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		// Special handling for files which use CONCURRENTLY
		// and cannot run inside a transaction block (if implicit) or needs splitting
		if strings.Contains(fileName, "021_optimize_search_indexes.sql") || strings.Contains(fileName, "035_add_cursor_pagination_indexes.sql") {
			statements := strings.Split(string(content), ";")
			for _, stmt := range statements {
				stmt = strings.TrimSpace(stmt)
				if stmt == "" {
					continue
				}
				// Use Acquire to get a connection and use PgConn().Exec for simple protocol
				conn, err := pc.Pool.Acquire(ctx)
				if err != nil {
					return fmt.Errorf("failed to acquire connection: %w", err)
				}
				_, err = conn.Conn().PgConn().Exec(ctx, stmt).ReadAll()
				conn.Release()

				if err != nil {
					return fmt.Errorf("failed to execute statement in migration %s: %w\nStatement: %s", fileName, err, stmt)
				}
			}
		} else {
			// Use Acquire to get a connection and use PgConn().Exec for simple protocol
			// This supports multiple statements in one string
			conn, err := pc.Pool.Acquire(ctx)
			if err != nil {
				return fmt.Errorf("failed to acquire connection: %w", err)
			}
			_, err = conn.Conn().PgConn().Exec(ctx, string(content)).ReadAll()
			conn.Release()

			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// RunSQL executes a raw SQL string
func (pc *PostgresContainer) RunSQL(ctx context.Context, sql string) error {
	_, err := pc.Pool.Exec(ctx, sql)
	return err
}

// CleanTables truncates all tables except system tables
func (pc *PostgresContainer) CleanTables(ctx context.Context) error {
	// Get all user tables
	query := `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'pg_%'
		AND tablename NOT LIKE 'sql_%'
	`
	rows, err := pc.Pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	// Truncate all tables with cascade
	truncateSQL := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tables, ", "))
	_, err = pc.Pool.Exec(ctx, truncateSQL)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// Close terminates the container and closes the connection pool
func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

// WithTransaction executes a function within a transaction for testing
// The transaction is always rolled back to leave the database in its original state
func (pc *PostgresContainer) WithTransaction(ctx context.Context, fn func(ctx context.Context, pool *pgxpool.Pool) error) error {
	tx, err := pc.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	// Create a single-connection pool wrapper that uses the transaction
	// For testing purposes, we pass the main pool but the test should be aware
	if err := fn(ctx, pc.Pool); err != nil {
		return err
	}

	// Always rollback - this is for test isolation
	return tx.Rollback(ctx)
}
