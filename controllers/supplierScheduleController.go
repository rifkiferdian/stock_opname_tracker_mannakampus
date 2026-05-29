package controllers

import (
	"database/sql"
	"errors"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func SupplierScheduleIndex(c *gin.Context) {
	scheduleService := buildSupplierScheduleService()
	renderSupplierSchedulePage(c, scheduleService, c.Query("error"), c.Query("success"), parseSupplierScheduleFilter(c))
}

func SupplierScheduleStore(c *gin.Context) {
	type supplierScheduleForm struct {
		StoreID    int    `form:"store_id" binding:"required"`
		SupplierID int    `form:"supplier_id" binding:"required"`
		DayOfWeek  int    `form:"day_of_week" binding:"required"`
		SOTime     string `form:"so_time"`
		SequenceNo int    `form:"sequence_no"`
		IsActive   int    `form:"is_active"`
		Notes      string `form:"notes"`
	}

	scheduleService := buildSupplierScheduleService()
	var form supplierScheduleForm

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierSchedulePage(c, scheduleService, "Form jadwal tidak lengkap", "", parseSupplierScheduleFilter(c))
		return
	}

	hasStoreAccess, err := currentUserHasStoreAccess(c, form.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		renderSupplierSchedulePage(c, scheduleService, "Store tidak tersedia untuk user login", "", parseSupplierScheduleFilter(c))
		return
	}

	err = scheduleService.CreateSchedule(models.SupplierScheduleCreateInput{
		StoreID:    form.StoreID,
		SupplierID: form.SupplierID,
		DayOfWeek:  form.DayOfWeek,
		SOTime:     form.SOTime,
		SequenceNo: form.SequenceNo,
		IsActive:   form.IsActive == 1,
		Notes:      form.Notes,
	})
	if err != nil {
		renderSupplierSchedulePage(c, scheduleService, err.Error(), "", parseSupplierScheduleFilter(c))
		return
	}

	c.Redirect(http.StatusSeeOther, buildSupplierScheduleURL("Jadwal supplier berhasil ditambahkan"))
}

func SupplierScheduleUpdate(c *gin.Context) {
	type supplierScheduleForm struct {
		ID         int    `form:"id" binding:"required"`
		StoreID    int    `form:"store_id" binding:"required"`
		SupplierID int    `form:"supplier_id" binding:"required"`
		DayOfWeek  int    `form:"day_of_week" binding:"required"`
		SOTime     string `form:"so_time"`
		SequenceNo int    `form:"sequence_no"`
		IsActive   int    `form:"is_active"`
		Notes      string `form:"notes"`
	}

	scheduleService := buildSupplierScheduleService()
	var form supplierScheduleForm

	if err := c.ShouldBind(&form); err != nil {
		renderSupplierSchedulePage(c, scheduleService, "Form jadwal tidak lengkap", "", parseSupplierScheduleFilter(c))
		return
	}

	hasStoreAccess, err := currentUserHasStoreAccess(c, form.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		renderSupplierSchedulePage(c, scheduleService, "Store tidak tersedia untuk user login", "", parseSupplierScheduleFilter(c))
		return
	}

	err = scheduleService.UpdateSchedule(models.SupplierScheduleUpdateInput{
		ID:         form.ID,
		StoreID:    form.StoreID,
		SupplierID: form.SupplierID,
		DayOfWeek:  form.DayOfWeek,
		SOTime:     form.SOTime,
		SequenceNo: form.SequenceNo,
		IsActive:   form.IsActive == 1,
		Notes:      form.Notes,
	})
	if err != nil {
		renderSupplierSchedulePage(c, scheduleService, err.Error(), "", parseSupplierScheduleFilter(c))
		return
	}

	c.Redirect(http.StatusSeeOther, buildSupplierScheduleURL("Jadwal supplier berhasil diperbarui"))
}

func SupplierScheduleDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid schedule id")
		return
	}

	scheduleService := buildSupplierScheduleService()
	schedule, err := scheduleService.GetScheduleByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Redirect(http.StatusSeeOther, buildSupplierScheduleErrorURL("Jadwal tidak ditemukan"))
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	hasStoreAccess, err := currentUserHasStoreAccess(c, schedule.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		c.Redirect(http.StatusSeeOther, buildSupplierScheduleErrorURL("Store tidak tersedia untuk user login"))
		return
	}

	if err := scheduleService.DeleteSchedule(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, buildSupplierScheduleURL("Jadwal supplier berhasil dihapus"))
}

func buildSupplierScheduleService() *services.SupplierScheduleService {
	repo := &repositories.SupplierScheduleRepository{DB: config.DB}
	return &services.SupplierScheduleService{Repo: repo}
}

func renderSupplierSchedulePage(c *gin.Context, scheduleService *services.SupplierScheduleService, errorMessage string, successMessage string, filter models.SupplierScheduleListFilter) {
	stores, err := getStoreOptionsForCurrentUser(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	allowedStoreSet := buildStoreAccessSet(stores)
	if !currentUserHasRole(c, "super-admin") && filter.StoreID > 0 && !isStoreAccessible(allowedStoreSet, filter.StoreID) {
		filter.StoreID = 0
		errorMessage = "Store filter tidak tersedia untuk user login"
	}

	allowedStoreIDs := make([]int, 0, len(stores))
	for _, store := range stores {
		if store.StoreID > 0 {
			allowedStoreIDs = append(allowedStoreIDs, store.StoreID)
		}
	}

	schedules, err := scheduleService.GetSchedules(filter, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stats, err := scheduleService.GetStats(allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	suppliers, err := scheduleService.GetSupplierOptions(allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "supplier_schedule.html", gin.H{
		"Title":     "Jadwal Supplier",
		"Page":      "schedule",
		"Stats":     stats,
		"Schedules": schedules,
		"Stores":    stores,
		"Suppliers": suppliers,
		"Filters":   filter,
		"Error":     strings.TrimSpace(errorMessage),
		"Success":   strings.TrimSpace(successMessage),
	})
}

func parseSupplierScheduleFilter(c *gin.Context) models.SupplierScheduleListFilter {
	storeID, _ := strconv.Atoi(c.DefaultQuery("store_id", "0"))
	dayOfWeek, _ := strconv.Atoi(c.DefaultQuery("day_of_week", "0"))

	return models.SupplierScheduleListFilter{
		Search:    c.Query("search"),
		StoreID:   storeID,
		DayOfWeek: dayOfWeek,
		Status:    c.Query("status"),
	}
}

func buildSupplierScheduleURL(successMessage string) string {
	values := url.Values{}
	if strings.TrimSpace(successMessage) != "" {
		values.Set("success", strings.TrimSpace(successMessage))
	}

	if encoded := values.Encode(); encoded != "" {
		return "/schedule?" + encoded
	}
	return "/schedule"
}

func buildSupplierScheduleErrorURL(errorMessage string) string {
	values := url.Values{}
	if strings.TrimSpace(errorMessage) != "" {
		values.Set("error", strings.TrimSpace(errorMessage))
	}
	if encoded := values.Encode(); encoded != "" {
		return "/schedule?" + encoded
	}
	return "/schedule"
}
