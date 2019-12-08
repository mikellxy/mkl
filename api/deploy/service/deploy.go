package service

import (
	"github.com/mikellxy/mkl/api/deploy/models"
	"github.com/pkg/errors"
)

const (
	GREEN         = "green"
	BLUE          = "blue"
	STATUS_PROD   = "production"
	STATUS_STAG   = "staging"
	STAGE_STOP    = "stop"
	STAGE_PENDING = "pending"
	STAGE_RUNNING = "running"
)

func GetStaging(project *models.Project) (*models.Deployment, error) {
	var blue, green *models.Deployment
	for _, d := range project.Deployments {
		if d.Color == GREEN {
			green = d
		} else {
			blue = d
		}
	}

	if blue != nil {
		// blue可用于staging
		if blue.Status == STATUS_STAG {
			if blue.Stage != STAGE_PENDING {
				return blue, nil
			}
			// 正在部署，返回错误
			return nil, errors.New("deploying")
		}

		if green != nil {
			if green.Stage != STAGE_PENDING {
				return green, nil
			}
			return nil, errors.New("deploying")
		}
		// 创建用于staging的deployment
		green, err := (&models.Deployment{}).Create(project.ID, GREEN)
		if err != nil {
			return nil, err
		}
		return green, nil
	} else if green != nil {
		// blue可用于staging
		if green.Status == STATUS_STAG {
			if green.Stage != STAGE_PENDING {
				return green, nil
			}
			return nil, errors.New("deploying")
		}

		blue, err := (&models.Deployment{}).Create(project.ID, BLUE)
		if err != nil {
			return nil, err
		}
		return blue, nil
	} else {
		blue, err := (&models.Deployment{}).Create(project.ID, BLUE)
		if err != nil {
			return nil, err
		}
		return blue, nil
	}
}
