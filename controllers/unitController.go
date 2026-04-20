package controllers

import (
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UnitIndex(c *gin.Context) {
	unitRepo := &repositories.UnitRepository{DB: config.DB}
	unitService := &services.UnitService{Repo: unitRepo}

	renderUnitPage(c, unitService, "")
}

func UnitStore(c *gin.Context) {
	type unitForm struct {
		UnitCode    string `form:"unit_code" binding:"required"`
		UnitName    string `form:"unit_name" binding:"required"`
		Description string `form:"description"`
	}

	var (
		form        unitForm
		unitRepo    = &repositories.UnitRepository{DB: config.DB}
		unitService = &services.UnitService{Repo: unitRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderUnitPage(c, unitService, "Form tidak lengkap")
		return
	}

	input := models.UnitCreateInput{
		UnitCode:    form.UnitCode,
		UnitName:    form.UnitName,
		Description: form.Description,
	}

	if err := unitService.CreateUnit(input); err != nil {
		renderUnitPage(c, unitService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/units")
}

func UnitUpdate(c *gin.Context) {
	type unitForm struct {
		ID          int    `form:"id" binding:"required"`
		UnitCode    string `form:"unit_code" binding:"required"`
		UnitName    string `form:"unit_name" binding:"required"`
		Description string `form:"description"`
	}

	var (
		form        unitForm
		unitRepo    = &repositories.UnitRepository{DB: config.DB}
		unitService = &services.UnitService{Repo: unitRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderUnitPage(c, unitService, "Form tidak lengkap")
		return
	}

	input := models.UnitUpdateInput{
		ID:          form.ID,
		UnitCode:    form.UnitCode,
		UnitName:    form.UnitName,
		Description: form.Description,
	}

	if err := unitService.UpdateUnit(input); err != nil {
		renderUnitPage(c, unitService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/units")
}

func UnitDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid unit id")
		return
	}

	unitRepo := &repositories.UnitRepository{DB: config.DB}
	unitService := &services.UnitService{Repo: unitRepo}

	if err := unitService.DeleteUnit(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/units")
}

func renderUnitPage(c *gin.Context, unitService *services.UnitService, message string) {
	units, err := unitService.GetUnits()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "unit.html", gin.H{
		"Title": "Units Master",
		"Page":  "unit",
		"units": units,
		"Error": message,
	})
}
