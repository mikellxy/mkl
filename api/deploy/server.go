package deploy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mikellxy/mkl/api/deploy/models"
	"github.com/mikellxy/mkl/api/deploy/service"
	"github.com/mikellxy/mkl/config"
	"github.com/mikellxy/mkl/pkg/database"
	"github.com/mikellxy/mkl/pkg/docker_api"
	"net/http"
)

func GetServer() *gin.Engine {
	r := gin.Default()
	p := r.Group("/deploy")
	p.POST("/projects/:id/deployments", deploy)
	p.POST("/projects", createProject)
	p.DELETE("projects/:id/services", removeService)
	p.POST("/projects/:id/deployments_switch", deploymentSwitch)

	p.POST("/packages", createPackage)
	return r
}

type createProjectBody struct {
	Name      string `json:"name" binding:"required"`
	BlueIP    string `json:"blue_ip" binding:"required"`
	BluePort  uint32 `json:"blue_port" binding:"required"`
	GreenIP   string `json:"green_ip" binding:"required"`
	GreenPort uint32 `json:"green_port" binding:"required"`
}

func createProject(c *gin.Context) {
	body := createProjectBody{}
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	p := &models.Project{}
	p, err := p.Create(body.Name, body.BlueIP, body.BluePort, body.GreenIP, body.GreenPort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, p)
}

type IntID struct {
	ID int `uri:"id" binding:"required"`
}

func deploy(c *gin.Context) {
	var ip string
	var port uint32

	param := IntID{}
	if err := c.BindUri(&param); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	// 获取project
	project := &models.Project{}
	project, err := project.FindOneByID(uint(param.ID), true)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	// 选择要部署的环境
	deployment, err := service.GetStaging(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	// 根据color确定项目部署的ip和port
	if deployment.Color == "green" {
		ip, port = project.GreenIP, project.GreenPort
	} else {
		ip, port = project.BlueIP, project.BluePort
	}
	// 获取项目最新的docker image版本
	pkg := &models.Package{}
	pkg, err = pkg.FindOneByProjectID(project.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	// 获取连接相应部署环境的docker client，使用docker api进行部署
	conf := config.Conf.DockerClient
	dockerClient, err := docker_api.NewDockerClient(fmt.Sprintf(conf.Host, ip))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	err = dockerClient.CreateSwarmService(fmt.Sprintf("%s_%s", project.Name, deployment.Color),
		pkg.Tag, 4, map[uint32]uint32{pkg.Port: port})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	// 部署成功之后，把deployment的stage改成running
	db, _ := database.GetDB()
	deployment.Stage = "running"
	deployment.PackageTag = pkg.Tag
	db.Save(deployment)
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, db.Error.Error())
		return
	}

	c.JSON(http.StatusCreated, deployment)
}

func removeService(c *gin.Context) {
	param := IntID{}
	if err := c.BindUri(&param); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	project := &models.Project{}
	project, err := project.FindOneByID(uint(param.ID), true)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	conf := config.Conf.DockerClient
	for _, d := range project.Deployments {
		var ip, name string
		if d.Color == "green" {
			ip = project.GreenIP
			name = fmt.Sprintf("%s_%s", project.Name, d.Color)
		}
		dockerClient, err := docker_api.NewDockerClient(fmt.Sprintf(conf.Host, ip))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		err = dockerClient.RemoveSwarmService(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
	c.JSON(http.StatusNoContent, "ok")
}

func deploymentSwitch(c *gin.Context) {
	param := IntID{}
	if err := c.BindUri(&param); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	// 获取project
	project := &models.Project{}
	project, err := project.FindOneByID(uint(param.ID), true)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	// 获取staging和production(第一次上线，还没有production) deployment
	var staging, other *models.Deployment
	for _, d := range project.Deployments {
		if d.Status == "staging" && d.Stage == "running" {
			staging = d
		} else {
			other = d
		}
	}
	if staging == nil {
		c.JSON(http.StatusInternalServerError, "no staging project is running")
		return
	}
	// 把staging的身份转换成staging，把原先的production的身份转换成staging，合适的时候可以停用老版本代码(调用docker api删除service，并把stage改成stop)
	db, _ := database.GetDB()
	staging.Status = "production"
	if other != nil {
		other.Status = "staging"
		other.Stage = "stop"
		db.Save(other)
	}
	db.Save(staging)
	c.JSON(http.StatusOK, project)
}

type createPackageBody struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	Tag       string `json:"tag" binding:"required"`
	Port      uint32 `json:"port" binding:"required"`
}

func createPackage(c *gin.Context) {
	body := createPackageBody{}
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	p := &models.Package{}
	p, err := p.Create(body.ProjectID, body.Tag, body.port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, p)
}
