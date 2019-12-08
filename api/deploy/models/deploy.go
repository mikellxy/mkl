package models

import (
	"github.com/mikellxy/mkl/pkg/database"
	"time"
)

type Model struct {
	ID        uint       `gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index"`
}

// 项目基础配置
type Project struct {
	Model
	Name        string        `gorm:"not null;unique_index:uix_project_name;type:varchar(64)" json:"name" binding:"required"`
	BlueIP      string        `json:"blue_ip"`
	BluePort    uint32        `json:"blue_port"`
	GreenIP     string        `json:"green_ip"`
	GreenPort   uint32        `json:"green_port"`
	Deployments []*Deployment `gorm:"association foreign:project_id"`
}

func (p *Project) FindOneByID(pid uint, joinedLoadDeployment bool) (*Project, error) {
	db, _ := database.GetDB()
	if joinedLoadDeployment {
		db.Preload("Deployments").Where("project.id=?", pid).First(p)
	} else {
		db.Where("project.id=?", pid).First(p)
	}
	if db.Error != nil {
		return nil, db.Error
	}
	return p, nil
}

func (p *Project) List() ([]*Project, error) {
	db, _ := database.GetDB()
	var ret []*Project
	db.Find(&ret)
	if db.Error != nil {
		return ret, db.Error
	}
	return ret, nil
}

func (p *Project) Create(name string, blueIP string, bluePort uint32, greenIP string, greenPort uint32) (*Project, error) {
	db, _ := database.GetDB()
	p.Name = name
	p.BlueIP = blueIP
	p.BluePort = bluePort
	p.GreenIP = greenIP
	p.GreenPort = greenPort
	db.Save(p)
	if db.Error != nil {
		return nil, db.Error
	}
	return p, nil
}

// 项目部署情况
type Deployment struct {
	Model
	ProjectID uint   `gorm:"not null;unique_index:uix_deployment_project_id_color" json:"project_id"`
	Color     string `gorm:"type:varchar(16);not null;unique_index:uix_deployment_project_id_color" json:"color"`
	// production or staging
	Status string `gorm:"type:varchar(32);not null;default:'staging'" json:"status"`
	// stop, pending or running
	Stage      string `gorm:"type:varchar(32);not null;default:'stop'" json:"stage"`
	PackageTag string `json:"package_tag"`
}

func (d *Deployment) Create(projectID uint, color string) (*Deployment, error) {
	db, _ := database.GetDB()
	d.ProjectID = projectID
	d.Color = color
	db.Create(d)
	if db.Error != nil {
		return nil, db.Error
	}
	return d, nil
}

type Package struct {
	Model
	ProjectID uint   `gorm:"not null" json:"project_id"`
	Tag       string `gorm:"not null;unique;" json:"tag"`
	Port      uint32 `gorm:"not null" json:"port"`
}

func (p *Package) FindOneByProjectID(projectID uint) (*Package, error) {
	db, _ := database.GetDB()
	db.Where("project_id=?", projectID).Order("id desc").First(p)
	if db.Error != nil {
		return nil, db.Error
	}
	return p, nil
}

func (p *Package) Create(projectID uint, tag string, port uint32) (*Package, error) {
	db, _ := database.GetDB()
	p.ProjectID = projectID
	p.Tag = tag
	p.Port = port
	db.Save(p)
	if db.Error != nil {
		return nil, db.Error
	}
	return p, nil
}
