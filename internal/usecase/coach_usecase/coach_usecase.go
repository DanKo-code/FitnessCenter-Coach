package user_usecase

import (
	"context"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/dtos"
	customErrors "github.com/DanKo-code/FitnessCenter-Coach/internal/errors"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/models"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/repository"
	"github.com/DanKo-code/FitnessCenter-Coach/pkg/logger"
	serviceGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.service"
	"github.com/google/uuid"
	"time"
)

type CoachUseCase struct {
	coachRepo     repository.CoachRepository
	serviceClient *serviceGRPC.ServiceClient
}

func NewCoachUseCase(coachRepo repository.CoachRepository, serviceClient *serviceGRPC.ServiceClient) *CoachUseCase {
	return &CoachUseCase{coachRepo: coachRepo, serviceClient: serviceClient}
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

func (c *CoachUseCase) GetCoachesWithServices(
	ctx context.Context,
) ([]*dtos.CoachWithServices, error) {
	coaches, err := c.coachRepo.GetCoaches(ctx)
	if err != nil {
		logger.ErrorLogger.Printf("Failed GetCoaches: %s", err)
		return nil, err
	}

	getCoachesServicesRequest := &serviceGRPC.GetCoachesServicesRequest{}

	for _, coach := range coaches {
		getCoachesServicesRequest.CoachIds =
			append(
				getCoachesServicesRequest.CoachIds,
				coach.Id.String(),
			)
	}

	getCoachesServicesResponse, err := (*c.serviceClient).GetCoachesServices(ctx, getCoachesServicesRequest)
	if err != nil {
		logger.ErrorLogger.Printf("Failed GetCoachesServices: %s", err)
		return nil, err
	}

	var coachWithServices []*dtos.CoachWithServices

	//add coaches
	for _, coach := range coaches {

		aws := &dtos.CoachWithServices{
			Coach:    coach,
			Services: nil,
		}

		coachWithServices = append(coachWithServices, aws)
	}

	//add services
	for _, extValue := range getCoachesServicesResponse.CoachIdsWithServices {

		coachId := extValue.CoachId

		for key, value := range coachWithServices {
			if value.Coach.Id.String() == coachId {
				coachWithServices[key].Services = append(coachWithServices[key].Services, extValue.ServiceObjects...)
			}
		}
	}

	return coachWithServices, nil
}
