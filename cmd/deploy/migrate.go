package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/mikellxy/mkl/api/deploy/models"
	"github.com/mikellxy/mkl/config"
	"github.com/mikellxy/mkl/pkg/database"
)

type helper struct {
	errs []error
	db *gorm.DB
}

func (h *helper) autoMigrate(values ...interface{}) {
	db := h.db.AutoMigrate(values...)
	if db.Error != nil {
		h.errs = append(h.errs, db.Error)
	}
}

func (h *helper) printError() {
	for _, err := range h.errs {
		fmt.Println(err)
	}
}

func main() {
	config.LoadConfig()
	db, err := database.GetDB()
	if err != nil {
		panic(err)
	}

	h := &helper{db:db}
	h.autoMigrate(&models.Project{})
	h.autoMigrate(&models.Deployment{})
	h.autoMigrate(&models.Package{})

	h.printError()
}
