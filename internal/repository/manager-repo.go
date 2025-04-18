package repository

import (
	"context"
	"gorm.io/gorm"
)

type ManagerRepo struct {
	db *gorm.DB
}

func NewManagerRepo(db *gorm.DB) *ManagerRepo {
	return &ManagerRepo{db: db}
}

func (mr *ManagerRepo) DeleteRecord(ctx context.Context, recordID int) error {
	return nil
}

func (mr *ManagerRepo) CreateRecord(ctx context.Context, recordID int) error {
	return nil
}
