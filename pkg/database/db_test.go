package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestItem struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
}

func TestGormTxManager_RunInTx(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&TestItem{})
	assert.NoError(t, err)

	txManager := NewTransactionManager(db)
	ctx := context.Background()

	t.Run("commit success", func(t *testing.T) {
		err := txManager.RunInTx(ctx, func(txCtx context.Context) error {
			txDB := GetDB(txCtx, db)
			item := TestItem{Name: "commit-test"}
			return txDB.Create(&item).Error
		})
		assert.NoError(t, err)

		var count int64
		err = db.Model(&TestItem{}).Where("name = ?", "commit-test").Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("rollback on error", func(t *testing.T) {
		err := txManager.RunInTx(ctx, func(txCtx context.Context) error {
			txDB := GetDB(txCtx, db)
			item := TestItem{Name: "rollback-test"}
			if err := txDB.Create(&item).Error; err != nil {
				return err
			}
			return errors.New("forced rollback error")
		})
		assert.Error(t, err)
		assert.Equal(t, "forced rollback error", err.Error())

		var count int64
		err = db.Model(&TestItem{}).Where("name = ?", "rollback-test").Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("rollback on panic", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = txManager.RunInTx(ctx, func(txCtx context.Context) error {
				txDB := GetDB(txCtx, db)
				item := TestItem{Name: "panic-test"}
				_ = txDB.Create(&item)
				panic("forced panic")
			})
		})

		var count int64
		err = db.Model(&TestItem{}).Where("name = ?", "panic-test").Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("nested transaction reuse", func(t *testing.T) {
		err := txManager.RunInTx(ctx, func(txCtx context.Context) error {
			txDB := GetDB(txCtx, db)
			item1 := TestItem{Name: "nested-1"}
			if err := txDB.Create(&item1).Error; err != nil {
				return err
			}

			return txManager.RunInTx(txCtx, func(nestedCtx context.Context) error {
				nestedDB := GetDB(nestedCtx, db)
				item2 := TestItem{Name: "nested-2"}
				return nestedDB.Create(&item2).Error
			})
		})
		assert.NoError(t, err)

		var count int64
		err = db.Model(&TestItem{}).Where("name IN ?", []string{"nested-1", "nested-2"}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
