package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test DB: %v", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("failed to auto migrate User table: %v", err)
	}

	return db
}

func TestRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	u := &User{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}

	// 1. Test Create
	err := repo.Create(ctx, u)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, u.ID)

	// 2. Test FindByID
	foundByID, err := repo.FindByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, u.Email, foundByID.Email)
	assert.Equal(t, u.FullName, foundByID.FullName)

	// 3. Test FindByEmail
	foundByEmail, err := repo.FindByEmail(ctx, u.Email)
	assert.NoError(t, err)
	assert.Equal(t, u.ID, foundByEmail.ID)

	// 4. Test FindAll
	allUsers, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, allUsers, 1)

	// 5. Test Update
	u.FullName = "Updated Name"
	err = repo.Update(ctx, u)
	assert.NoError(t, err)

	updatedUser, err := repo.FindByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.FullName)

	// 6. Test Delete
	err = repo.Delete(ctx, u.ID)
	assert.NoError(t, err)

	_, err = repo.FindByID(ctx, u.ID)
	assert.Error(t, err) // GORM First returns error for soft-deleted elements by default
}
