package user_usecase

import (
	"context"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/dtos"
	customErrors "github.com/DanKo-code/FitnessCenter-Coach/internal/errors"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/models"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/repository"
	"github.com/google/uuid"
	"time"
)

type CoachUseCase struct {
	coachRepo repository.CoachRepository
}

func NewCoachUseCase(coachRepo repository.CoachRepository) *CoachUseCase {
	return &CoachUseCase{coachRepo: coachRepo}
}

func (c *CoachUseCase) CreateCoach(
	ctx context.Context,
	cmd *dtos.CreateCoachCommand,
) (*models.Coach, error) {

	coach := &models.Coach{
		Id:          cmd.Id,
		Name:        cmd.Name,
		Description: cmd.Description,
		Photo:       cmd.Photo,
		UpdatedTime: time.Now(),
		CreatedTime: time.Now(),
	}

	createdCoach, err := c.coachRepo.CreateCoach(ctx, coach)
	if err != nil {
		return nil, err
	}

	return createdCoach, nil
}

func (c *CoachUseCase) GetCoachById(
	ctx context.Context,
	id uuid.UUID,
) (*models.Coach, error) {
	coach, err := c.coachRepo.GetCoachById(ctx, id)
	if err != nil {
		return nil, err
	}

	return coach, nil
}

func (c *CoachUseCase) UpdateCoach(
	ctx context.Context,
	cmd *dtos.UpdateCoachCommand,
) (*models.Coach, error) {

	err := c.coachRepo.UpdateCoach(ctx, cmd)
	if err != nil {
		return nil, err
	}

	coach, err := c.coachRepo.GetCoachById(ctx, cmd.Id)
	if err != nil {
		return nil, err
	}

	return coach, nil
}

func (c *CoachUseCase) DeleteCoachById(
	ctx context.Context,
	id uuid.UUID,
) (*models.Coach, error) {
	coach, err := c.coachRepo.GetCoachById(ctx, id)
	if err != nil {
		return nil, customErrors.CoachNotFound
	}

	err = c.coachRepo.DeleteCoachById(ctx, id)
	if err != nil {
		return nil, err
	}

	return coach, nil
}

func (c *CoachUseCase) GetCoaches(
	ctx context.Context,
) ([]*models.Coach, error) {

	coaches, err := c.coachRepo.GetCoaches(ctx)
	if err != nil {
		return nil, err
	}

	return coaches, nil
}
