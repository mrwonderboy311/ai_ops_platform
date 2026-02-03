// Package handler provides integration tests for API endpoints
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"myops.k8s.io/backend/pkg/model"
)

// IntegrationTestSuite provides a test suite for integration testing
type IntegrationTestSuite struct {
	DB      *gorm.DB
	Router  *gin.Engine
	Handler *RBACHandler
}

// SetupIntegrationTest creates a new test suite with database and router
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Create in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// Migrate all models
	err = db.AutoMigrate(
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
		&model.User{},
		&model.UserRole{},
		&model.ResourceAccessPolicy{},
		&model.AnomalyDetectionRule{},
		&model.AnomalyEvent{},
		&model.LLMConversation{},
		&model.LLMMessage{},
	)
	require.NoError(t, err)

	// Seed default data
	model.SeedDefaultPermissions(db)
	model.SeedDefaultRoles(db)

	// Create Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create handler
	handler := NewRBACHandler(db)

	return &IntegrationTestSuite{
		DB:      db,
		Router:  router,
		Handler: handler,
	}
}

// TestRBACWorkflow tests the complete RBAC workflow
func TestRBACWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	// Setup routes
	suite.Router.POST("/api/v1/rbac/roles", suite.Handler.CreateRole)
	suite.Router.GET("/api/v1/rbac/roles", suite.Handler.ListRoles)
	suite.Router.GET("/api/v1/rbac/roles/:id", suite.Handler.GetRole)
	suite.Router.PATCH("/api/v1/rbac/roles/:id", suite.Handler.UpdateRole)
	suite.Router.DELETE("/api/v1/rbac/roles/:id", suite.Handler.DeleteRole)
	suite.Router.POST("/api/v1/rbac/roles/:id/permissions", suite.Handler.AssignRolePermissions)
	suite.Router.GET("/api/v1/rbac/roles/:id/permissions", suite.Handler.ListRolePermissions)

	// Test: Seed default roles and verify
	t.Run("SeedDefaultRoles", func(t *testing.T) {
		// Get super_admin role
		var superAdmin model.Role
		err := suite.DB.Where("name = ?", "super_admin").First(&superAdmin).Error
		require.NoError(t, err)
		assert.True(t, superAdmin.IsSystem)
	})

	// Test: Create a custom role
	t.Run("CreateCustomRole", func(t *testing.T) {
		newRole := CreateRoleRequest{
			Name:        "custom_operator",
			DisplayName: "Custom Operator",
			Description: "A custom operator role for testing",
			IsDefault:   false,
		}

		body, _ := json.Marshal(newRole)
		req, _ := http.NewRequest("POST", "/api/v1/rbac/roles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response model.Role
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "custom_operator", response.Name)
		assert.False(t, response.IsSystem)
	})

	// Test: List roles and verify custom role exists
	t.Run("ListRoles", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/rbac/roles?page=1&pageSize=10", nil)
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data     []model.Role `json:"data"`
			Total    int          `json:"total"`
			Page     int          `json:"page"`
			PageSize int          `json:"pageSize"`
		}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response.Total, 5) // At least 4 default + 1 custom
	})

	// Test: Assign permissions to role
	t.Run("AssignPermissionsToRole", func(t *testing.T) {
		// Get the custom role we created
		var customRole model.Role
		err := suite.DB.Where("name = ?", "custom_operator").First(&customRole).Error
		require.NoError(t, err)

		// Get some permissions
		var permissions []model.Permission
		err = suite.DB.Limit(3).Find(&permissions).Error
		require.NoError(t, err)

		permIDs := make([]uuid.UUID, len(permissions))
		for i, perm := range permissions {
			permIDs[i] = perm.ID
		}

		assignReq := AssignPermissionsRequest{
			PermissionIDs: permIDs,
			Override:      true,
		}
		body, _ := json.Marshal(assignReq)
		url := fmt.Sprintf("/api/v1/rbac/roles/%s/permissions", customRole.ID.String())
		req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Permissions assigned successfully", response["message"])
	})

	// Test: Get role permissions
	t.Run("GetRolePermissions", func(t *testing.T) {
		var customRole model.Role
		err := suite.DB.Where("name = ?", "custom_operator").First(&customRole).Error
		require.NoError(t, err)

		url := fmt.Sprintf("/api/v1/rbac/roles/%s/permissions", customRole.ID.String())
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data  []model.Permission `json:"data"`
			Total int               `json:"total"`
		}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response.Total, 3)
	})

	// Test: Update role
	t.Run("UpdateRole", func(t *testing.T) {
		var customRole model.Role
		err := suite.DB.Where("name = ?", "custom_operator").First(&customRole).Error
		require.NoError(t, err)

		updateData := map[string]interface{}{
			"description": "Updated description",
		}
		body, _ := json.Marshal(updateData)
		url := fmt.Sprintf("/api/v1/rbac/roles/%s", customRole.ID.String())
		req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response model.Role
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", response.Description)
	})

	// Test: Delete custom role
	t.Run("DeleteCustomRole", func(t *testing.T) {
		var customRole model.Role
		err := suite.DB.Where("name = ?", "custom_operator").First(&customRole).Error
		require.NoError(t, err)

		url := fmt.Sprintf("/api/v1/rbac/roles/%s", customRole.ID.String())
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		var count int64
		suite.DB.Model(&model.Role{}).Where("id = ?", customRole.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	// Test: Cannot delete system role
	t.Run("CannotDeleteSystemRole", func(t *testing.T) {
		var superAdmin model.Role
		err := suite.DB.Where("name = ?", "super_admin").First(&superAdmin).Error
		require.NoError(t, err)

		url := fmt.Sprintf("/api/v1/rbac/roles/%s", superAdmin.ID.String())
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestUserPermissionWorkflow tests user permission checking workflow
func TestUserPermissionWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	// Setup routes
	suite.Router.POST("/api/v1/rbac/users/:userId/roles", suite.Handler.AssignUserRole)
	suite.Router.GET("/api/v1/rbac/users/:userId/roles", suite.Handler.ListUserRoles)
	suite.Router.GET("/api/v1/rbac/users/:userId/permissions", suite.Handler.GetUserPermissions)
	suite.Router.POST("/api/v1/rbac/users/:userId/check-permission", suite.Handler.CheckPermission)

	// Test: Create a test user
	t.Run("CreateTestUser", func(t *testing.T) {
		user := model.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // hashed "password"
		}
		err := suite.DB.Create(&user).Error
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
	})

	// Test: Assign viewer role to user
	t.Run("AssignViewerRoleToUser", func(t *testing.T) {
		var user model.User
		err := suite.DB.Where("username = ?", "testuser").First(&user).Error
		require.NoError(t, err)

		var viewerRole model.Role
		err = suite.DB.Where("name = ?", "viewer").First(&viewerRole).Error
		require.NoError(t, err)

		assignReq := AssignUserRoleRequest{
			RoleID: viewerRole.ID,
		}
		body, _ := json.Marshal(assignReq)
		url := fmt.Sprintf("/api/v1/rbac/users/%s/roles", user.ID.String())
		req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response model.UserRole
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, viewerRole.ID, response.RoleID)
	})

	// Test: Check user has permissions
	t.Run("CheckUserHasPermissions", func(t *testing.T) {
		var user model.User
		err := suite.DB.Where("username = ?", "testuser").First(&user).Error
		require.NoError(t, err)

		url := fmt.Sprintf("/api/v1/rbac/users/%s/permissions", user.ID.String())
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Permissions []model.Permission        `json:"permissions"`
			Grouped     map[string][]model.Permission `json:"grouped"`
			Total       int                       `json:"total"`
		}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response.Total, 1)
		assert.NotEmpty(t, response.Grouped)
	})

	// Test: Check specific permission
	t.Run("CheckSpecificPermission", func(t *testing.T) {
		var user model.User
		err := suite.DB.Where("username = ?", "testuser").First(&user).Error
		require.NoError(t, err)

		checkReq := CheckPermissionRequest{
			Resource: "hosts",
			Action:   "read",
		}
		body, _ := json.Marshal(checkReq)
		url := fmt.Sprintf("/api/v1/rbac/users/%s/check-permission", user.ID.String())
		req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response["allowed"].(bool))
	})
}

// TestPermissionEnforcement tests permission enforcement
func TestPermissionEnforcement(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
		&model.User{},
		&model.UserRole{},
	)
	require.NoError(t, err)

	model.SeedDefaultPermissions(db)
	model.SeedDefaultRoles(db)

	// Create test user
	user := model.User{
		Username: "permission_test_user",
		Email:    "permission@example.com",
		Password: "hashedpassword",
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	// Assign viewer role (limited permissions)
	var viewerRole model.Role
	err = db.Where("name = ?", "viewer").First(&viewerRole).Error
	require.NoError(t, err)

	userRole := model.UserRole{
		UserID: user.ID,
		RoleID: viewerRole.ID,
	}
	err = db.Create(&userRole).Error
	require.NoError(t, err)

	// Test: User should have read permissions but not write
	t.Run("ViewerHasReadNotWrite", func(t *testing.T) {
		result := model.UserHasPermission(db, user.ID, "hosts", "read", nil, "")
		assert.True(t, result.Allowed, "Viewer should have hosts.read permission")

		result = model.UserHasPermission(db, user.ID, "hosts", "update", nil, "")
		assert.False(t, result.Allowed, "Viewer should NOT have hosts.update permission")
	})

	// Test: User does not have delete permissions
	t.Run("ViewerCannotDelete", func(t *testing.T) {
		result := model.UserHasPermission(db, user.ID, "hosts", "delete", nil, "")
		assert.False(t, result.Allowed, "Viewer should NOT have hosts.delete permission")
	})

	// Test: Assign admin role and verify permissions
	t.Run("AdminHasAllPermissions", func(t *testing.T) {
		var adminRole model.Role
		err = db.Where("name = ?", "admin").First(&adminRole).Error
		require.NoError(t, err)

		// Remove old role
		db.Where("user_id = ?", user.ID).Delete(&model.UserRole{})

		// Assign admin role
		userRole = model.UserRole{
			UserID: user.ID,
			RoleID: adminRole.ID,
		}
		err = db.Create(&userRole).Error
		require.NoError(t, err)

		// Should now have write permissions
		result := model.UserHasPermission(db, user.ID, "hosts", "update", nil, "")
		assert.True(t, result.Allowed, "Admin should have hosts.update permission")

		result = model.UserHasPermission(db, user.ID, "hosts", "delete", nil, "")
		assert.True(t, result.Allowed, "Admin should have hosts.delete permission")
	})
}

// BenchmarkPermissionCheck benchmarks permission checking performance
func BenchmarkPermissionCheck(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	err = db.AutoMigrate(
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
		&model.User{},
		&model.UserRole{},
	)
	if err != nil {
		b.Fatal(err)
	}

	model.SeedDefaultPermissions(db)
	model.SeedDefaultRoles(db)

	user := model.User{
		Username: "bench_user",
		Email:    "bench@example.com",
		Password: "hashedpassword",
	}
	db.Create(&user)

	var viewerRole model.Role
	db.Where("name = ?", "viewer").First(&viewerRole)

	userRole := model.UserRole{
		UserID: user.ID,
		RoleID: viewerRole.ID,
	}
	db.Create(&userRole)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.UserHasPermission(db, user.ID, "hosts", "read", nil, "")
	}
}

// GetTestDataDir returns the test data directory
func GetTestDataDir(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(dir, "..", "..", "testdata")
}
