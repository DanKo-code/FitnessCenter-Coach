package grpc

import (
	"context"
	"errors"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/dtos"
	customErrors "github.com/DanKo-code/FitnessCenter-Coach/internal/errors"
	"github.com/DanKo-code/FitnessCenter-Coach/internal/usecase"
	"github.com/DanKo-code/FitnessCenter-Coach/pkg/logger"
	coachProtobuf "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.coach"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"time"
)

var _ coachProtobuf.CoachServer = (*CoachgRPC)(nil)

type CoachgRPC struct {
	coachProtobuf.UnimplementedCoachServer

	coachUseCase usecase.CoachUseCase
	cloudUseCase usecase.CloudUseCase
}

func RegisterCoachServer(gRPC *grpc.Server, coachUseCase usecase.CoachUseCase, cloudUseCase usecase.CloudUseCase) {
	coachProtobuf.RegisterCoachServer(gRPC, &CoachgRPC{coachUseCase: coachUseCase, cloudUseCase: cloudUseCase})
}

func (c *CoachgRPC) CreateCoach(g grpc.ClientStreamingServer[coachProtobuf.CreateCoachRequest, coachProtobuf.CreateCoachResponse]) error {

	coachData, coachPhoto, err := GetObjectData(
		&g,
		func(chunk *coachProtobuf.CreateCoachRequest) interface{} {
			return chunk.GetCoachDataForCreate()
		},
		func(chunk *coachProtobuf.CreateCoachRequest) []byte {
			return chunk.GetCoachPhoto()
		},
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request data")
	}

	if coachData == nil {
		logger.ErrorLogger.Printf("coach data is empty")
		return status.Error(codes.InvalidArgument, "coach data is empty")
	}

	castedCoachData, ok := coachData.(*coachProtobuf.CoachDataForCreate)
	if !ok {
		logger.ErrorLogger.Printf("coach data is not of type CoachProtobuf.CoachDataForCreate")
		return status.Error(codes.InvalidArgument, "coach data is not of type CoachProtobuf.CoachDataForCreate")
	}

	cmd := &dtos.CreateCoachCommand{
		Id:          uuid.New(),
		Name:        castedCoachData.Name,
		Description: castedCoachData.Description,
		Photo:       "",
	}

	var photoURL string
	if coachPhoto != nil {
		url, err := c.cloudUseCase.PutObject(context.TODO(), coachPhoto, "coach/"+cmd.Id.String())
		photoURL = url
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create coach photo in cloud: %v", err)
			return status.Error(codes.Internal, "Failed to create coach photo in cloud")
		}
	}

	cmd.Photo = photoURL

	coach, err := c.coachUseCase.CreateCoach(context.TODO(), cmd)
	if err != nil {

		if photoURL == "" {
			err := c.cloudUseCase.DeleteObject(context.TODO(), "coach/"+cmd.Id.String())
			if err != nil {
				logger.ErrorLogger.Printf("Failed to delete coach photo from cloud: %v", err)
				return status.Error(codes.Internal, "Failed to delete coach photo in cloud")
			}
		}

		return status.Error(codes.Internal, "Failed to create coach")
	}

	coachObject := &coachProtobuf.CoachObject{
		Id:          coach.Id.String(),
		Name:        coach.Name,
		Description: coach.Description,
		Photo:       coach.Photo,
		CreatedTime: coach.CreatedTime.String(),
		UpdatedTime: coach.UpdatedTime.String(),
	}

	response := &coachProtobuf.CreateCoachResponse{
		CoachObject: coachObject,
	}

	err = g.SendAndClose(response)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to send coach create response: %v", err)
		return status.Error(codes.Internal, "Failed to send coach create response")
	}

	return nil
}

func (c *CoachgRPC) GetCoachById(ctx context.Context, request *coachProtobuf.GetCoachByIdRequest) (*coachProtobuf.GetCoachByIdResponse, error) {

	coach, err := c.coachUseCase.GetCoachById(ctx, uuid.MustParse(request.Id))
	if err != nil {

		if errors.Is(err, customErrors.CoachNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, err
	}

	coachObject := &coachProtobuf.CoachObject{
		Id:          coach.Id.String(),
		Name:        coach.Name,
		Description: coach.Description,
		Photo:       coach.Photo,
		CreatedTime: coach.CreatedTime.String(),
		UpdatedTime: coach.UpdatedTime.String(),
	}

	response := &coachProtobuf.GetCoachByIdResponse{
		CoachObject: coachObject,
	}

	return response, nil
}

func (c *CoachgRPC) UpdateCoach(g grpc.ClientStreamingServer[coachProtobuf.UpdateCoachRequest, coachProtobuf.UpdateCoachResponse]) error {
	coachData, coachPhoto, err := GetObjectData(
		&g,
		func(chunk *coachProtobuf.UpdateCoachRequest) interface{} {
			return chunk.GetCoachDataForUpdate()
		},
		func(chunk *coachProtobuf.UpdateCoachRequest) []byte {
			return chunk.GetCoachPhoto()
		},
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request data")
	}

	if coachData == nil {
		logger.ErrorLogger.Printf("coach data is empty")
		return status.Error(codes.InvalidArgument, "coach data is empty")
	}

	castedCoachData, ok := coachData.(coachProtobuf.CoachDataForUpdate)
	if !ok {
		logger.ErrorLogger.Printf("coach data is not of type CoachProtobuf.CoachDataForCreate")
		return status.Error(codes.InvalidArgument, "coach data is not of type CoachProtobuf.CoachDataForCreate")
	}

	cmd := &dtos.UpdateCoachCommand{
		Id:          uuid.MustParse(castedCoachData.Id),
		Name:        castedCoachData.Name,
		Description: castedCoachData.Description,
		UpdatedTime: time.Now(),
	}

	_, err = c.coachUseCase.GetCoachById(context.TODO(), uuid.MustParse(castedCoachData.Id))
	if err != nil {
		return status.Error(codes.NotFound, "coach not found")
	}

	var photoURL string
	var previousPhoto []byte
	if coachPhoto != nil {
		previousPhoto, err = c.cloudUseCase.GetObjectByName(context.TODO(), "coach/"+cmd.Id.String())
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get previos photo from cloud: %v", err)
			return err
		}

		url, err := c.cloudUseCase.PutObject(context.TODO(), coachPhoto, "coach/"+cmd.Id.String())
		photoURL = url
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create coach photo in cloud: %v", err)
			return status.Error(codes.Internal, "Failed to create coach photo in cloud")
		}
	}

	cmd.Photo = photoURL

	coach, err := c.coachUseCase.UpdateCoach(context.TODO(), cmd)
	if err != nil {

		_, err := c.cloudUseCase.PutObject(context.TODO(), previousPhoto, "coach/"+cmd.Id.String())
		if err != nil {
			logger.ErrorLogger.Printf("Failed to set previous photo in cloud: %v", err)
			return status.Error(codes.Internal, "Failed to create coach photo in cloud")
		}

		return status.Error(codes.Internal, "Failed to create coach")
	}

	coachObject := &coachProtobuf.CoachObject{
		Id:          coach.Id.String(),
		Name:        coach.Name,
		Description: coach.Description,
		Photo:       coach.Photo,
		CreatedTime: coach.CreatedTime.String(),
		UpdatedTime: coach.UpdatedTime.String(),
	}

	updateCoachResponse := &coachProtobuf.UpdateCoachResponse{
		CoachObject: coachObject,
	}

	err = g.SendAndClose(updateCoachResponse)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to send coach update response: %v", err)
		return err
	}

	return nil
}

func (c *CoachgRPC) DeleteCoachById(ctx context.Context, request *coachProtobuf.DeleteCoachByIdRequest) (*coachProtobuf.DeleteCoachByIdResponse, error) {
	deletedCoach, err := c.coachUseCase.DeleteCoachById(ctx, uuid.MustParse(request.Id))
	if err != nil {
		return nil, err
	}

	coachObject := &coachProtobuf.CoachObject{
		Id:          deletedCoach.Id.String(),
		Name:        deletedCoach.Name,
		Description: deletedCoach.Description,
		Photo:       deletedCoach.Photo,
		CreatedTime: deletedCoach.CreatedTime.String(),
		UpdatedTime: deletedCoach.UpdatedTime.String(),
	}

	deleteCoachByIdResponse := &coachProtobuf.DeleteCoachByIdResponse{
		CoachObject: coachObject,
	}

	return deleteCoachByIdResponse, nil
}

func (c *CoachgRPC) GetCoaches(ctx context.Context, _ *emptypb.Empty) (*coachProtobuf.GetCoachesResponse, error) {

	coaches, err := c.coachUseCase.GetCoaches(ctx)
	if err != nil {
		return nil, err
	}

	var coachObjects []*coachProtobuf.CoachObject

	for _, coach := range coaches {

		coachObject := &coachProtobuf.CoachObject{
			Id:          coach.Id.String(),
			Name:        coach.Name,
			Description: coach.Description,
			Photo:       coach.Photo,
			CreatedTime: coach.CreatedTime.String(),
			UpdatedTime: coach.UpdatedTime.String(),
		}

		coachObjects = append(coachObjects, coachObject)
	}

	response := &coachProtobuf.GetCoachesResponse{CoachObjects: coachObjects}

	return response, nil
}

func GetObjectData[T any, R any](
	g *grpc.ClientStreamingServer[T, R],
	extractObjectData func(chunk *T) interface{},
	extractObjectPhoto func(chunk *T) []byte,
) (interface{},
	[]byte,
	error,
) {
	var objectData interface{}
	var objectPhoto []byte

	for {
		chunk, err := (*g).Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.ErrorLogger.Printf("Error getting chunk: %v", err)
			return nil, nil, err
		}

		if ud := extractObjectData(chunk); ud != nil {
			objectData = ud
		}

		if uf := extractObjectPhoto(chunk); uf != nil {
			objectPhoto = append(objectPhoto, uf...)
		}
	}

	return objectData, objectPhoto, nil
}
