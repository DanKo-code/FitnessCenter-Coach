package repository

import (
	"User/internal/models"
	"context"
	"github.com/google/uuid"
)

type CoachRepository interface {
	CreateCoach(ctx context.Context, coach *models.Coach) (*models.Coach, error)
	GetCoaches(ctx context.Context) ([]*models.Coach, error)
	DeleteCoachById(ctx context.Context, id uuid.UUID) error
	UpdateCoach(ctx context.Context, coach *models.Coach) error
	GetCoachById(ctx context.Context, id uuid.UUID) (*models.Coach, error)
}
