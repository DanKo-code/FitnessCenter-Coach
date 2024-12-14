package dtos

import (
	"github.com/DanKo-code/FitnessCenter-Coach/internal/models"
	serviceGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.service"
)

type CoachWithServices struct {
	Coach    *models.Coach                `db:"coach"`
	Services []*serviceGRPC.ServiceObject `db:"services"`
}
