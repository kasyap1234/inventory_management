package repositories

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"agromart2/internal/models"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RoleRepoTestSuite struct {
	suite.Suite
	mock       pgxmock.PgxPoolIface
	repo       RoleRepository
	tenantID1  uuid.UUID
	tenantID2  uuid.UUID
	roleID     uuid.UUID
	context    context.Context
}

func (suite *RoleRepoTestSuite) SetupTest() {
	mock, err := pgxmock.NewPool()
	assert.NoError(suite.T(), err)
	suite.mock = mock

	suite.repo = NewRoleRepo(mock)
	suite.tenantID1 = uuid.New()
	suite.tenantID2 = uuid.New()
	suite.roleID = uuid.New()
	suite.context = context.Background()
}

func (suite *RoleRepoTestSuite) TearDownTest() {
	suite.mock.Close()
}

func TestRoleRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RoleRepoTestSuite))
}

func (suite *RoleRepoTestSuite) TestCreate_Success() {
	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    suite.tenantID1,
		Name:        "Admin",
		Description: stringPtr("Administrator role"),
	}

	suite.mock.ExpectExec(`
		INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
		VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
		ON CONFLICT \(tenant_id, name\) DO NOTHING
	`).WithArgs(role.ID, role.TenantID, role.Name, role.Description).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := suite.repo.Create(suite.context, role)
	assert.NoError(suite.T(), err)
}

func (suite *RoleRepoTestSuite) TestCreate_DuplicateNameInSameTenant() {
	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    suite.tenantID1,
		Name:        "Manager",
		Description: stringPtr("Manager role"),
	}

	suite.mock.ExpectExec(`
		INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
		VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
		ON CONFLICT \(tenant_id, name\) DO NOTHING
	`).WithArgs(role.ID, role.TenantID, role.Name, role.Description).
		WillReturnResult(pgxmock.NewResult("INSERT", 0)) // No rows affected due to conflict

	err := suite.repo.Create(suite.context, role)
	assert.NoError(suite.T(), err) // ON CONFLICT DO NOTHING doesn't error
}

func (suite *RoleRepoTestSuite) TestCreate_SameNameInDifferentTenant() {
	role1 := &models.Role{
		ID:          uuid.New(),
		TenantID:    suite.tenantID1,
		Name:        "User",
		Description: stringPtr("User role tenant 1"),
	}
	role2 := &models.Role{
		ID:          uuid.New(),
		TenantID:    suite.tenantID2,
		Name:        "User",
		Description: stringPtr("User role tenant 2"),
	}

	// Both should succeed (different tenant)
	suite.mock.ExpectExec(`
		INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
		VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
		ON CONFLICT \(tenant_id, name\) DO NOTHING
	`).WithArgs(role1.ID, role1.TenantID, role1.Name, role1.Description).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	suite.mock.ExpectExec(`
		INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
		VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
		ON CONFLICT \(tenant_id, name\) DO NOTHING
	`).WithArgs(role2.ID, role2.TenantID, role2.Name, role2.Description).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := suite.repo.Create(suite.context, role1)
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(suite.context, role2)
	assert.NoError(suite.T(), err)
}

func (suite *RoleRepoTestSuite) TestCreate_DatabaseError() {
	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    suite.tenantID1,
		Name:        "Editor",
		Description: stringPtr("Editor role"),
	}

	suite.mock.ExpectExec(`
		INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
		VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
		ON CONFLICT \(tenant_id, name\) DO NOTHING
	`).WithArgs(role.ID, role.TenantID, role.Name, role.Description).
		WillReturnError(errors.New("database connection failed"))

	err := suite.repo.Create(suite.context, role)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

func (suite *RoleRepoTestSuite) TestGetByID_Success() {
	role := &models.Role{
		ID:          suite.roleID,
		TenantID:    suite.tenantID1,
		Name:        "Viewer",
		Description: stringPtr("Viewer role"),
	}

	createdAt := "2023-01-01T00:00:00Z"
	updatedAt := "2023-01-01T00:00:00Z"

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1 AND id = \$2
	`).WithArgs(suite.tenantID1, suite.roleID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
			AddRow(suite.roleID, suite.tenantID1, "Viewer", "Viewer role", createdAt, updatedAt))

	result, err := suite.repo.GetByID(suite.context, suite.tenantID1, suite.roleID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), role.ID, result.ID)
	assert.Equal(suite.T(), role.TenantID, result.TenantID)
	assert.Equal(suite.T(), role.Name, result.Name)
	assert.Equal(suite.T(), *role.Description, *result.Description)
}

func (suite *RoleRepoTestSuite) TestGetByID_WrongTenant() {
	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1 AND id = \$2
	`).WithArgs(suite.tenantID2, suite.roleID).
		WillReturnError(pgx.ErrNoRows)

	result, err := suite.repo.GetByID(suite.context, suite.tenantID2, suite.roleID)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, pgx.ErrNoRows)
	assert.Nil(suite.T(), result)
}

func (suite *RoleRepoTestSuite) TestGetByID_NotFound() {
	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1 AND id = \$2
	`).WithArgs(suite.tenantID1, suite.roleID).
		WillReturnError(pgx.ErrNoRows)

	result, err := suite.repo.GetByID(suite.context, suite.tenantID1, suite.roleID)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, pgx.ErrNoRows)
	assert.Nil(suite.T(), result)
}

func (suite *RoleRepoTestSuite) TestGetByName_Success() {
	roleName := "Operator"
	expectedRole := &models.Role{
		ID:          suite.roleID,
		TenantID:    suite.tenantID1,
		Name:        roleName,
		Description: stringPtr("Operator role"),
	}

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1 AND name = \$2
	`).WithArgs(suite.tenantID1, roleName).
		WillReturnRows(pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
			AddRow(expectedRole.ID, expectedRole.TenantID, expectedRole.Name, *expectedRole.Description, "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z"))

	result, err := suite.repo.GetByName(suite.context, suite.tenantID1, roleName)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedRole.Name, result.Name)
	assert.Equal(suite.T(), expectedRole.TenantID, result.TenantID)
}

func (suite *RoleRepoTestSuite) TestGetByName_WrongTenant() {
	roleName := "Supervisor"

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1 AND name = \$2
	`).WithArgs(suite.tenantID2, roleName).
		WillReturnError(pgx.ErrNoRows)

	result, err := suite.repo.GetByName(suite.context, suite.tenantID2, roleName)
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, pgx.ErrNoRows)
	assert.Nil(suite.T(), result)
}

func (suite *RoleRepoTestSuite) TestUpdate_Success() {
	role := &models.Role{
		ID:          suite.roleID,
		TenantID:    suite.tenantID1,
		Name:        "Modified Admin",
		Description: stringPtr("Modified administrator role"),
	}

	suite.mock.ExpectExec(`
		UPDATE roles
		SET name = \$1, description = \$2, updated_at = NOW\(\)
		WHERE tenant_id = \$3 AND id = \$4
	`).WithArgs(role.Name, role.Description, role.TenantID, role.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := suite.repo.Update(suite.context, role)
	assert.NoError(suite.T(), err)
}

func (suite *RoleRepoTestSuite) TestUpdate_NoRowsAffected() {
	role := &models.Role{
		ID:          suite.roleID,
		TenantID:    suite.tenantID1,
		Name:        "Non-existent Role",
		Description: stringPtr("This won't be found"),
	}

	suite.mock.ExpectExec(`
		UPDATE roles
		SET name = \$1, description = \$2, updated_at = NOW\(\)
		WHERE tenant_id = \$3 AND id = \$4
	`).WithArgs(role.Name, role.Description, role.TenantID, role.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := suite.repo.Update(suite.context, role)
	assert.NoError(suite.T(), err) // Update doesn't error if no rows affected
}

func (suite *RoleRepoTestSuite) TestDelete_Success() {
	suite.mock.ExpectExec(`DELETE FROM roles WHERE tenant_id = \$1 AND id = \$2`).
		WithArgs(suite.tenantID1, suite.roleID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := suite.repo.Delete(suite.context, suite.tenantID1, suite.roleID)
	assert.NoError(suite.T(), err)
}

func (suite *RoleRepoTestSuite) TestDelete_WrongTenant() {
	suite.mock.ExpectExec(`DELETE FROM roles WHERE tenant_id = \$1 AND id = \$2`).
		WithArgs(suite.tenantID2, suite.roleID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := suite.repo.Delete(suite.context, suite.tenantID2, suite.roleID)
	assert.NoError(suite.T(), err) // Doesn't error even if no rows deleted
}

func (suite *RoleRepoTestSuite) TestList_Success() {
	limit, offset := 10, 0

	rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
		AddRow(uuid.New(), suite.tenantID1, "Role1", "Description1", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z").
		AddRow(uuid.New(), suite.tenantID1, "Role2", "Description2", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3
	`).WithArgs(suite.tenantID1, limit, offset).
		WillReturnRows(rows)

	result, err := suite.repo.List(suite.context, suite.tenantID1, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "Role1", result[0].Name)
	assert.Equal(suite.T(), "Role2", result[1].Name)
}

func (suite *RoleRepoTestSuite) TestList_EmptyResult() {
	limit, offset := 5, 0

	rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"})

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3
	`).WithArgs(suite.tenantID1, limit, offset).
		WillReturnRows(rows)

	result, err := suite.repo.List(suite.context, suite.tenantID1, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), result)
}

func (suite *RoleRepoTestSuite) TestList_WithOffset() {
	limit, offset := 5, 10

	rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
		AddRow(uuid.New(), suite.tenantID1, "Role10", "Description10", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3
	`).WithArgs(suite.tenantID1, limit, offset).
		WillReturnRows(rows)

	result, err := suite.repo.List(suite.context, suite.tenantID1, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "Role10", result[0].Name)
}

func (suite *RoleRepoTestSuite) TestList_DifferentTenants() {
	limit, offset := 10, 0

	// Tenant 1 has roles
	tenant1Rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
		AddRow(uuid.New(), suite.tenantID1, "Admin", "Admin role", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3
	`).WithArgs(suite.tenantID1, limit, offset).
		WillReturnRows(tenant1Rows)

	// Tenant 2 has different roles
	tenant2Rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
		AddRow(uuid.New(), suite.tenantID2, "User", "User role", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	suite.mock.ExpectQuery(`
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM roles
		WHERE tenant_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3
	`).WithArgs(suite.tenantID2, limit, offset).
		WillReturnRows(tenant2Rows)

	result1, err := suite.repo.List(suite.context, suite.tenantID1, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result1, 1)
	assert.Equal(suite.T(), "Admin", result1[0].Name)
	assert.Equal(suite.T(), suite.tenantID1, result1[0].TenantID)

	result2, err := suite.repo.List(suite.context, suite.tenantID2, limit, offset)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result2, 1)
	assert.Equal(suite.T(), "User", result2[0].Name)
	assert.Equal(suite.T(), suite.tenantID2, result2[0].TenantID)
}

func (suite *RoleRepoTestSuite) TestRoleHierarchyLogic() {
	// Test hierarchical role names
	roles := []models.Role{
		{ID: uuid.New(), TenantID: suite.tenantID1, Name: "SuperAdmin", Description: stringPtr("Top level admin")},
		{ID: uuid.New(), TenantID: suite.tenantID1, Name: "Admin", Description: stringPtr("Administrator")},
		{ID: uuid.New(), TenantID: suite.tenantID1, Name: "Manager", Description: stringPtr("Department manager")},
		{ID: uuid.New(), TenantID: suite.tenantID1, Name: "User", Description: stringPtr("Basic user")},
	}

	// Sort by hierarchy (though repository doesn't enforce this, test logic)
	hierarchyOrder := []string{"SuperAdmin", "Admin", "Manager", "User"}

	for _, role := range roles {
		suite.mock.ExpectExec(`
			INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
			VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
			ON CONFLICT \(tenant_id, name\) DO NOTHING
		`).WithArgs(role.ID, role.TenantID, role.Name, role.Description).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := suite.repo.Create(suite.context, &role)
		assert.NoError(suite.T(), err)
	}

	// Verify each role can be found by name within tenant
	for _, expectedName := range hierarchyOrder {
		row := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"}).
			AddRow(uuid.New(), suite.tenantID1, expectedName, "Test description", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

		suite.mock.ExpectQuery(`
			SELECT id, tenant_id, name, description, created_at, updated_at
			FROM roles
			WHERE tenant_id = \$1 AND name = \$2
		`).WithArgs(suite.tenantID1, expectedName).
			WillReturnRows(row)

		found, err := suite.repo.GetByName(suite.context, suite.tenantID1, expectedName)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), expectedName, found.Name)
	}
}

func (suite *RoleRepoTestSuite) TestComplexPagination() {
	limit, offset := 3, 6

	// Test with larger dataset
	expectations := []struct {
		name string
		args interface{}
	}{
		{"roles.page1", []interface{}{suite.tenantID1, 3, 0}},
		{"roles.page2", []interface{}{suite.tenantID1, 3, 3}},
		{"roles.page3", []interface{}{suite.tenantID1, 3, 6}},
	}

	for _, exp := range expectations {
		rows := pgxmock.NewRows([]string{"id", "tenant_id", "name", "description", "created_at", "updated_at"})
		for i := offset; i < offset+limit && i < 20; i++ {
			rows.AddRow(uuid.New(), suite.tenantID1, "Role"+strconv.Itoa(i), "Description"+strconv.Itoa(i), "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")
		}

		// Note: This would need proper SQL escaping in real tests
		suite.mock.ExpectQuery(`
			SELECT id, tenant_id, name, description, created_at, updated_at
			FROM roles
			WHERE tenant_id = \$1
			ORDER BY created_at DESC
			LIMIT \$2 OFFSET \$3
		`).WithArgs(exp.args.([]interface{})...).
			WillReturnRows(rows)
	}

	// This simplified test demonstrates pagination logic
	// In practice, you'd iterate through pages

	result, err := suite.repo.List(suite.context, suite.tenantID1, limit, offset)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	// Expected results would vary based on mock setup
}

func (suite *RoleRepoTestSuite) TestContextCancellation() {
	cancelledCtx, cancel := context.WithCancel(suite.context)
	cancel() // Cancel immediately

	role := &models.Role{
		ID:          suite.roleID,
		TenantID:    suite.tenantID1,
		Name:        "Cancelled",
		Description: stringPtr("Will be cancelled"),
	}

	suite.mock.ExpectExec(`
		INSERT INTO roles \(id, tenant_id, name, description, created_at, updated_at\)
		VALUES \(\$1, \$2, \$3, \$4, NOW\(\), NOW\(\)\)
		ON CONFLICT \(tenant_id, name\) DO NOTHING
	`).WithArgs(role.ID, role.TenantID, role.Name, role.Description).
		WillReturnError(context.Canceled)

	err := suite.repo.Create(cancelledCtx, role)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}