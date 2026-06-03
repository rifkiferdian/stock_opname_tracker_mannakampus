package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const stockCheckSessionDetailItemLimit = 0

func StockCheckCheckerSupplierIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	dayOfWeek, _ := strconv.Atoi(c.DefaultQuery("day_of_week", "0"))
	lastSODate := sanitizeQueryDate(c.Query("last_so_date"))
	if dayOfWeek < 1 || dayOfWeek > 7 {
		dayOfWeek = 0
	}

	renderStockCheckCheckerSupplierPage(c, buildSupplierService(), "", models.SupplierListFilter{
		Search:     c.Query("search"),
		SearchMode: "name_code",
		Status:     "active",
		DayOfWeek:  dayOfWeek,
		LastSODate: lastSODate,
		Sort:       "recent",
		Page:       page,
		Limit:      50,
	})
}

func StockCheckCheckerSupplierDetail(c *gin.Context) {
	supplierID, err := strconv.Atoi(c.Param("id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	renderStockCheckCheckerDetailPage(
		c,
		buildSupplierService(),
		buildStockCheckSessionService(),
		supplierID,
		models.StockCheckSessionListFilter{
			SupplierID: supplierID,
			Status:     c.Query("status"),
			Page:       parsePositiveInt(c.Query("page"), 1),
			Limit:      10,
		},
		c.Query("error"),
		c.Query("success"),
		"",
		models.StockCheckSession{},
		extractCurrentUserID(c),
	)
}

func StockCheckCheckerSessionCreatePage(c *gin.Context) {
	supplierID, err := strconv.Atoi(c.Param("id"))
	if err != nil || supplierID <= 0 {
		c.String(http.StatusBadRequest, "invalid supplier id")
		return
	}

	backURL := sanitizeRedirectTarget(c.Query("back_to"))
	if backURL == "" {
		backURL = buildStockCheckCheckerDefaultBackURL(supplierID)
	}

	renderStockCheckCheckerCreateSessionPage(
		c,
		buildSupplierService(),
		buildStockCheckSessionService(),
		supplierID,
		"",
		models.StockCheckSession{},
		extractCurrentUserID(c),
		backURL,
	)
}

func StockCheckCheckerSessionInput(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	renderStockCheckCheckerSessionInputPage(
		c,
		buildStockCheckSessionService(),
		sessionID,
		c.Query("success"),
		c.Query("error"),
		models.StockCheckSessionCheckerScanForm{
			Location: sanitizeStockCheckCheckerScanLocation(c.Query("location")),
		},
	)
}

func StockCheckCheckerSessionScanPage(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	barcode := strings.TrimSpace(c.Query("barcode"))
	if barcode == "" {
		c.Redirect(http.StatusSeeOther, buildStockCheckCheckerSessionInputPageURL(sessionID, sanitizeStockCheckCheckerScanLocation(c.Query("location")), ""))
		return
	}

	backURL := sanitizeRedirectTarget(c.Query("back_to"))
	if backURL == "" {
		backURL = buildStockCheckCheckerSessionInputPageURL(sessionID, sanitizeStockCheckCheckerScanLocation(c.Query("location")), "")
	}

	renderStockCheckCheckerSessionScanPage(
		c,
		buildStockCheckSessionService(),
		sessionID,
		barcode,
		sanitizeStockCheckCheckerScanLocation(c.Query("location")),
		c.Query("success"),
		c.Query("error"),
		models.StockCheckSessionCheckerScanForm{},
		backURL,
	)
}

func StockCheckCheckerSessionSuggest(c *gin.Context) {
	type stockCheckCheckerSuggestForm struct {
		ItemID        int    `form:"item_id" binding:"required"`
		SuggestCarton string `form:"suggest_carton" binding:"required"`
		SuggestBox    string `form:"suggest_box"`
		SuggestPcs    string `form:"suggest_pcs"`
		RedirectTo    string `form:"redirect_to"`
	}

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	var form stockCheckCheckerSuggestForm
	service := buildStockCheckSessionService()

	if err := c.ShouldBind(&form); err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Form suggest tidak lengkap"))
			return
		}
		c.String(http.StatusBadRequest, "form suggest tidak lengkap")
		return
	}

	suggestCarton, err := parseStockCheckNonNegativeInt(form.SuggestCarton)
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Suggest carton harus berupa angka bulat yang valid"))
			return
		}
		c.String(http.StatusBadRequest, "suggest carton harus berupa angka bulat yang valid")
		return
	}

	suggestBox, err := parseStockCheckNonNegativeInt(form.SuggestBox)
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Suggest box harus berupa angka bulat yang valid"))
			return
		}
		c.String(http.StatusBadRequest, "suggest box harus berupa angka bulat yang valid")
		return
	}

	suggestPcs, err := parseStockCheckNonNegativeInt(form.SuggestPcs)
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Suggest pcs harus berupa angka bulat yang valid"))
			return
		}
		c.String(http.StatusBadRequest, "suggest pcs harus berupa angka bulat yang valid")
		return
	}

	err = service.UpdateCheckerSuggest(models.StockCheckSessionCheckerSuggestInput{
		SessionID:     sessionID,
		ItemID:        form.ItemID,
		SuggestCarton: suggestCarton,
		SuggestBox:    suggestBox,
		SuggestPcs:    suggestPcs,
		UpdatedBy:     extractCurrentUserID(c),
	})
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", err.Error()))
			return
		}
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "success", "Suggest checker berhasil disimpan"))
		return
	}

	c.Redirect(http.StatusSeeOther, buildStockCheckCheckerSessionInputPageURL(sessionID, "store", "Suggest checker berhasil disimpan"))
}

func StockCheckCheckerSessionScan(c *gin.Context) {
	type stockCheckCheckerScanForm struct {
		Location   string `form:"location" binding:"required"`
		Barcode    string `form:"barcode" binding:"required"`
		QtyCarton  string `form:"qty_carton"`
		QtyBox     string `form:"qty_box"`
		QtyPcs     string `form:"qty_pcs"`
		RedirectTo string `form:"redirect_to"`
		BackTo     string `form:"back_to"`
	}

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	var form stockCheckCheckerScanForm
	service := buildStockCheckSessionService()
	redirectTo := ""
	backTo := ""

	renderScanPage := func(message string) bool {
		if isStockCheckCheckerSessionScanRedirectTarget(redirectTo, sessionID) {
			renderStockCheckCheckerSessionScanPage(
				c,
				service,
				sessionID,
				form.Barcode,
				sanitizeStockCheckCheckerScanLocation(form.Location),
				"",
				message,
				models.StockCheckSessionCheckerScanForm{
					Location:  sanitizeStockCheckCheckerScanLocation(form.Location),
					Barcode:   strings.TrimSpace(form.Barcode),
					QtyCarton: strings.TrimSpace(form.QtyCarton),
					QtyBox:    strings.TrimSpace(form.QtyBox),
					QtyPcs:    strings.TrimSpace(form.QtyPcs),
				},
				backTo,
			)
			return true
		}
		return false
	}

	if err := c.ShouldBind(&form); err != nil {
		redirectTo = sanitizeRedirectTarget(form.RedirectTo)
		backTo = sanitizeRedirectTarget(form.BackTo)
		if backTo == "" {
			backTo = buildStockCheckCheckerSessionInputPageURL(sessionID, sanitizeStockCheckCheckerScanLocation(form.Location), "")
		}
		if renderScanPage("Form scan item tidak lengkap") {
			return
		}
		renderStockCheckCheckerSessionInputPage(c, service, sessionID, "", "Form scan item tidak lengkap", models.StockCheckSessionCheckerScanForm{
			Location:  sanitizeStockCheckCheckerScanLocation(form.Location),
			Barcode:   strings.TrimSpace(form.Barcode),
			QtyCarton: strings.TrimSpace(form.QtyCarton),
			QtyBox:    strings.TrimSpace(form.QtyBox),
			QtyPcs:    strings.TrimSpace(form.QtyPcs),
		})
		return
	}

	redirectTo = sanitizeRedirectTarget(form.RedirectTo)
	backTo = sanitizeRedirectTarget(form.BackTo)
	if backTo == "" {
		backTo = buildStockCheckCheckerSessionInputPageURL(sessionID, sanitizeStockCheckCheckerScanLocation(form.Location), "")
	}

	qtyCarton, err := parseStockCheckNonNegativeInt(form.QtyCarton)
	if err != nil {
		if renderScanPage("Qty carton harus berupa angka bulat yang valid") {
			return
		}
		renderStockCheckCheckerSessionInputPage(c, service, sessionID, "", "Qty carton harus berupa angka bulat yang valid", models.StockCheckSessionCheckerScanForm{
			Location:  sanitizeStockCheckCheckerScanLocation(form.Location),
			Barcode:   strings.TrimSpace(form.Barcode),
			QtyCarton: strings.TrimSpace(form.QtyCarton),
			QtyBox:    strings.TrimSpace(form.QtyBox),
			QtyPcs:    strings.TrimSpace(form.QtyPcs),
		})
		return
	}

	qtyBox, err := parseStockCheckNonNegativeInt(form.QtyBox)
	if err != nil {
		if renderScanPage("Qty box harus berupa angka bulat yang valid") {
			return
		}
		renderStockCheckCheckerSessionInputPage(c, service, sessionID, "", "Qty box harus berupa angka bulat yang valid", models.StockCheckSessionCheckerScanForm{
			Location:  sanitizeStockCheckCheckerScanLocation(form.Location),
			Barcode:   strings.TrimSpace(form.Barcode),
			QtyCarton: strings.TrimSpace(form.QtyCarton),
			QtyBox:    strings.TrimSpace(form.QtyBox),
			QtyPcs:    strings.TrimSpace(form.QtyPcs),
		})
		return
	}

	qtyPcs, err := parseStockCheckNonNegativeInt(form.QtyPcs)
	if err != nil {
		if renderScanPage("Qty pcs harus berupa angka bulat yang valid") {
			return
		}
		renderStockCheckCheckerSessionInputPage(c, service, sessionID, "", "Qty pcs harus berupa angka bulat yang valid", models.StockCheckSessionCheckerScanForm{
			Location:  sanitizeStockCheckCheckerScanLocation(form.Location),
			Barcode:   strings.TrimSpace(form.Barcode),
			QtyCarton: strings.TrimSpace(form.QtyCarton),
			QtyBox:    strings.TrimSpace(form.QtyBox),
			QtyPcs:    strings.TrimSpace(form.QtyPcs),
		})
		return
	}

	_, err = service.RecordCheckerScan(models.StockCheckSessionCheckerScanInput{
		SessionID: sessionID,
		Location:  form.Location,
		Barcode:   form.Barcode,
		QtyCarton: qtyCarton,
		QtyBox:    qtyBox,
		QtyPcs:    qtyPcs,
		UpdatedBy: extractCurrentUserID(c),
	})
	if err != nil {
		if renderScanPage(err.Error()) {
			return
		}
		renderStockCheckCheckerSessionInputPage(c, service, sessionID, "", err.Error(), models.StockCheckSessionCheckerScanForm{
			Location:  sanitizeStockCheckCheckerScanLocation(form.Location),
			Barcode:   strings.TrimSpace(form.Barcode),
			QtyCarton: strings.TrimSpace(form.QtyCarton),
			QtyBox:    strings.TrimSpace(form.QtyBox),
			QtyPcs:    strings.TrimSpace(form.QtyPcs),
		})
		return
	}

	if backTo != "" {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(backTo, "success", "Qty item berhasil disimpan"))
		return
	}

	c.Redirect(http.StatusSeeOther, buildStockCheckCheckerSessionInputPageURL(sessionID, sanitizeStockCheckCheckerScanLocation(form.Location), "Qty item berhasil disimpan"))
}

func StockCheckSessionIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	storeID, _ := strconv.Atoi(c.DefaultQuery("store_id", "0"))
	supplierID, _ := strconv.Atoi(c.DefaultQuery("supplier_id", "0"))

	renderStockCheckSessionPage(c, buildStockCheckSessionService(), "", "", models.StockCheckSession{}, models.StockCheckSessionListFilter{
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		StoreID:    storeID,
		SupplierID: supplierID,
		Status:     c.Query("status"),
		Page:       page,
		Limit:      10,
	})
}

func StockCheckSessionDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	renderStockCheckSessionDetailPage(c, buildStockCheckSessionService(), id, c.Query("success"), "", models.StockCheckSessionReviewItemEditForm{})
}

type stockCheckSessionPOItemView struct {
	No                  int
	ProductName         string
	ProductCode         string
	UnitName            string
	QtyDisplay          string
	QtyBreakdownDisplay string
	QtyCarton           int
	QtyBox              int
	QtyPcs              int
	UnitPriceDisplay    string
	SubtotalDisplay     string
	GroupID             int
	GroupName           string
	GroupSortOrder      int
}

type stockCheckSessionPOGroupSection struct {
	GroupID     int
	GroupName   string
	GroupLabel  string
	SortOrder   int
	Ungrouped   bool
	ItemCount   int
	TotalCarton int
	TotalBox    int
	TotalPcs    int
	Items       []stockCheckSessionPOItemView
}

func StockCheckSessionPODetail(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}
	poSortBy := sanitizeStockCheckSessionPOSortBy(c.Query("sort_by"))

	sessionService := buildStockCheckSessionService()
	pageData, err := sessionService.GetSessionDetailPage(sessionID, 1, 0, models.StockCheckSessionDetailFilter{})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Stock check session tidak ditemukan",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, sessionService, pageData.Session.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"code_error": 3,
			"error":      "Anda Tidak punya Akses di Halaman ini",
		})
		return
	}

	supplier, err := buildSupplierService().GetSupplierByID(pageData.Session.SupplierID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	storeAddress := ""
	stores, err := buildStoreService().GetStores()
	if err == nil {
		for _, store := range stores {
			if store.StoreID == pageData.Session.StoreID {
				storeAddress = strings.TrimSpace(store.StoreAddress)
				break
			}
		}
	}

	buyerApproverName, err := sessionService.Repo.GetLatestBuyerApproverName(sessionID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if strings.TrimSpace(buyerApproverName) == "" {
		buyerApproverName = "-"
	}

	items := make([]stockCheckSessionPOItemView, 0, len(pageData.Items))
	subtotal := 0.0
	totalQty := 0.0
	totalCarton := 0
	totalBox := 0
	totalPcs := 0
	for _, item := range pageData.Items {
		if item.Status != "approved" && item.Status != "po_created" {
			continue
		}

		unitPrice := 0.0
		if item.ApprovedBuyQty > 0 {
			unitPrice = item.ApprovedLineValue / item.ApprovedBuyQty
		}
		lineSubtotal := item.ApprovedLineValue
		subtotal += lineSubtotal
		totalQty += item.ApprovedBuyQty
		totalCarton += item.ApprovedBuyCarton
		totalBox += item.ApprovedBuyBox
		totalPcs += item.ApprovedBuyPcs

		items = append(items, stockCheckSessionPOItemView{
			ProductName:         item.ProductName,
			ProductCode:         item.ProductCode,
			UnitName:            item.UnitName,
			QtyDisplay:          formatStockCheckPOWholeNumber(item.ApprovedBuyQty),
			QtyBreakdownDisplay: formatStockCheckPOBreakdown(item.ApprovedBuyCarton, item.ApprovedBuyBox, item.ApprovedBuyPcs),
			QtyCarton:           item.ApprovedBuyCarton,
			QtyBox:              item.ApprovedBuyBox,
			QtyPcs:              item.ApprovedBuyPcs,
			UnitPriceDisplay:    formatStockCheckPOCurrency(unitPrice),
			SubtotalDisplay:     formatStockCheckPOCurrency(lineSubtotal),
			GroupID:             item.SupplierProductGroupID,
			GroupName:           strings.TrimSpace(item.SupplierProductGroupName),
			GroupSortOrder:      item.SupplierProductGroupSortOrder,
		})
	}

	poSections := buildStockCheckSessionPOSections(items, poSortBy)

	shippingHandling := 0.0
	estimatedTax := subtotal * 0.08
	totalAmount := subtotal + shippingHandling + estimatedTax

	Render(c, "stock_check_session_po_detail.html", gin.H{
		"Title":                   "PO Detail " + pageData.Session.SessionNumber,
		"Page":                    "stock_check_po_recap",
		"CurrentRole":             extractCurrentUserRole(c),
		"Success":                 strings.TrimSpace(c.Query("success")),
		"Error":                   strings.TrimSpace(c.Query("error")),
		"CurrentDetailURL":        c.Request.URL.RequestURI(),
		"Session":                 pageData.Session,
		"Supplier":                supplier,
		"StoreAddress":            storeAddress,
		"BuyerApproverName":       buyerApproverName,
		"POItems":                 items,
		"POSections":              poSections,
		"POSortBy":                poSortBy,
		"POItemCount":             len(items),
		"POTotalQtyDisplay":       formatStockCheckPOWholeNumber(totalQty),
		"POTotalBreakdownDisplay": formatStockCheckPOBreakdown(totalCarton, totalBox, totalPcs),
		"POTotalCarton":           totalCarton,
		"POTotalBox":              totalBox,
		"POTotalPcs":              totalPcs,
		"SubtotalDisplay":         formatStockCheckPOCurrency(subtotal),
		"ShippingDisplay":         formatStockCheckPOCurrency(shippingHandling),
		"EstimatedTaxDisplay":     formatStockCheckPOCurrency(estimatedTax),
		"TotalAmountDisplay":      formatStockCheckPOCurrency(totalAmount),
		"PODateDisplay":           pageData.Session.SessionDateDisplay,
		"PONumberDisplay":         "PO-" + pageData.Session.SessionNumber,
		"BackToRecapPOPageURL": buildStockCheckPORecapPageURL(models.StockCheckSessionListFilter{
			DateFrom:     sanitizeQueryDate(c.Query("date_from")),
			DateTo:       sanitizeQueryDate(c.Query("date_to")),
			SupplierName: c.Query("supplier_name"),
		}, parsePositiveInt(c.Query("page"), 1)),
	})
}

func buildStockCheckSessionPOSections(items []stockCheckSessionPOItemView, poSortBy string) []stockCheckSessionPOGroupSection {
	if len(items) == 0 {
		return nil
	}

	sectionsByKey := make(map[string]*stockCheckSessionPOGroupSection)
	order := make([]string, 0, len(items))
	for _, item := range items {
		groupName := strings.TrimSpace(item.GroupName)
		key := strconv.Itoa(item.GroupID) + "|" + groupName
		if item.GroupID <= 0 || groupName == "" {
			key = "ungrouped"
		}

		section, exists := sectionsByKey[key]
		if !exists {
			section = &stockCheckSessionPOGroupSection{
				GroupID:    item.GroupID,
				GroupName:  groupName,
				GroupLabel: groupName,
				SortOrder:  item.GroupSortOrder,
				Ungrouped:  key == "ungrouped",
				Items:      make([]stockCheckSessionPOItemView, 0, 8),
			}
			if section.Ungrouped {
				section.GroupID = 0
				section.GroupName = ""
				section.GroupLabel = "Item Tanpa Group"
				section.SortOrder = 1 << 30
			}
			sectionsByKey[key] = section
			order = append(order, key)
		}

		section.Items = append(section.Items, item)
		section.ItemCount++
		section.TotalCarton += item.QtyCarton
		section.TotalBox += item.QtyBox
		section.TotalPcs += item.QtyPcs
	}

	sections := make([]stockCheckSessionPOGroupSection, 0, len(order))
	for _, key := range order {
		sections = append(sections, *sectionsByKey[key])
	}

	sort.SliceStable(sections, func(i, j int) bool {
		if sections[i].Ungrouped != sections[j].Ungrouped {
			return !sections[i].Ungrouped
		}
		if sections[i].SortOrder != sections[j].SortOrder {
			return sections[i].SortOrder < sections[j].SortOrder
		}

		leftName := strings.ToLower(strings.TrimSpace(sections[i].GroupLabel))
		rightName := strings.ToLower(strings.TrimSpace(sections[j].GroupLabel))
		if leftName != rightName {
			return leftName < rightName
		}

		return sections[i].GroupID < sections[j].GroupID
	})

	rowNo := 1
	for i := range sections {
		sort.SliceStable(sections[i].Items, func(leftIndex, rightIndex int) bool {
			leftName := strings.ToLower(strings.TrimSpace(sections[i].Items[leftIndex].ProductName))
			rightName := strings.ToLower(strings.TrimSpace(sections[i].Items[rightIndex].ProductName))
			if leftName != rightName {
				if poSortBy == "name_desc" {
					return leftName > rightName
				}
				return leftName < rightName
			}

			leftCode := strings.ToLower(strings.TrimSpace(sections[i].Items[leftIndex].ProductCode))
			rightCode := strings.ToLower(strings.TrimSpace(sections[i].Items[rightIndex].ProductCode))
			if leftCode != rightCode {
				if poSortBy == "name_desc" {
					return leftCode > rightCode
				}
				return leftCode < rightCode
			}

			return leftIndex < rightIndex
		})

		for itemIndex := range sections[i].Items {
			sections[i].Items[itemIndex].No = rowNo
			rowNo++
		}
	}

	return sections
}

func sanitizeStockCheckSessionPOSortBy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "name_desc", "desc":
		return "name_desc"
	default:
		return "name_asc"
	}
}

func StockCheckSessionReviewItemUpdate(c *gin.Context) {
	type stockCheckSessionReviewItemForm struct {
		ItemID            int    `form:"item_id" binding:"required"`
		ApprovedBuyCarton string `form:"approved_buy_carton"`
		ApprovedBuyBox    string `form:"approved_buy_box"`
		ApprovedBuyPcs    string `form:"approved_buy_pcs"`
		ApprovedBuyQty    string `form:"approved_buy_qty"`
		BuyerNotes        string `form:"buyer_notes"`
		RedirectTo        string `form:"redirect_to"`
	}

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	var form stockCheckSessionReviewItemForm
	service := buildStockCheckSessionService()

	session, err := service.Repo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.String(http.StatusNotFound, "stock check session tidak ditemukan")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, session.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		if redirectTo := sanitizeRedirectTarget(c.PostForm("redirect_to")); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Anda Tidak punya Akses di Halaman ini"))
			return
		}
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"code_error": 3,
			"error":      "Anda Tidak punya Akses di Halaman ini",
		})
		return
	}

	if err := c.ShouldBind(&form); err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Form edit item tidak lengkap"))
			return
		}
		renderStockCheckSessionDetailPage(c, service, sessionID, "", "Form edit item tidak lengkap", models.StockCheckSessionReviewItemEditForm{
			ItemID:            form.ItemID,
			ApprovedBuyCarton: form.ApprovedBuyCarton,
			ApprovedBuyBox:    form.ApprovedBuyBox,
			ApprovedBuyPcs:    form.ApprovedBuyPcs,
			ApprovedBuyQty:    form.ApprovedBuyQty,
			BuyerNotes:        form.BuyerNotes,
		})
		return
	}

	approvedBuyCarton, err := parseStockCheckNonNegativeInt(form.ApprovedBuyCarton)
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Final approve carton harus berupa angka bulat yang valid"))
			return
		}
		renderStockCheckSessionDetailPage(c, service, sessionID, "", "Final approve carton harus berupa angka bulat yang valid", models.StockCheckSessionReviewItemEditForm{
			ItemID:            form.ItemID,
			ApprovedBuyCarton: form.ApprovedBuyCarton,
			ApprovedBuyBox:    form.ApprovedBuyBox,
			ApprovedBuyPcs:    form.ApprovedBuyPcs,
			ApprovedBuyQty:    form.ApprovedBuyQty,
			BuyerNotes:        form.BuyerNotes,
		})
		return
	}

	approvedBuyBox, err := parseStockCheckNonNegativeInt(form.ApprovedBuyBox)
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Final approve box harus berupa angka bulat yang valid"))
			return
		}
		renderStockCheckSessionDetailPage(c, service, sessionID, "", "Final approve box harus berupa angka bulat yang valid", models.StockCheckSessionReviewItemEditForm{
			ItemID:            form.ItemID,
			ApprovedBuyCarton: form.ApprovedBuyCarton,
			ApprovedBuyBox:    form.ApprovedBuyBox,
			ApprovedBuyPcs:    form.ApprovedBuyPcs,
			ApprovedBuyQty:    form.ApprovedBuyQty,
			BuyerNotes:        form.BuyerNotes,
		})
		return
	}

	approvedBuyPcs, err := parseStockCheckNonNegativeInt(form.ApprovedBuyPcs)
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Final approve pcs harus berupa angka bulat yang valid"))
			return
		}
		renderStockCheckSessionDetailPage(c, service, sessionID, "", "Final approve pcs harus berupa angka bulat yang valid", models.StockCheckSessionReviewItemEditForm{
			ItemID:            form.ItemID,
			ApprovedBuyCarton: form.ApprovedBuyCarton,
			ApprovedBuyBox:    form.ApprovedBuyBox,
			ApprovedBuyPcs:    form.ApprovedBuyPcs,
			ApprovedBuyQty:    form.ApprovedBuyQty,
			BuyerNotes:        form.BuyerNotes,
		})
		return
	}

	// Backward compatibility for pages/forms still sending a single approved_buy_qty field.
	if strings.TrimSpace(form.ApprovedBuyCarton) == "" &&
		strings.TrimSpace(form.ApprovedBuyBox) == "" &&
		strings.TrimSpace(form.ApprovedBuyPcs) == "" &&
		strings.TrimSpace(form.ApprovedBuyQty) != "" {
		legacyApprovedQty, legacyErr := strconv.ParseFloat(strings.TrimSpace(form.ApprovedBuyQty), 64)
		if legacyErr != nil {
			if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
				c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Final approve harus berupa angka yang valid"))
				return
			}
			renderStockCheckSessionDetailPage(c, service, sessionID, "", "Final approve harus berupa angka yang valid", models.StockCheckSessionReviewItemEditForm{
				ItemID:            form.ItemID,
				ApprovedBuyQty:    form.ApprovedBuyQty,
				ApprovedBuyCarton: form.ApprovedBuyCarton,
				ApprovedBuyBox:    form.ApprovedBuyBox,
				ApprovedBuyPcs:    form.ApprovedBuyPcs,
				BuyerNotes:        form.BuyerNotes,
			})
			return
		}
		approvedBuyPcs = int(math.Round(legacyApprovedQty))
		if approvedBuyPcs < 0 {
			approvedBuyPcs = 0
		}
	}

	err = service.UpdateReviewItem(models.StockCheckSessionReviewItemUpdateInput{
		SessionID:         sessionID,
		ItemID:            form.ItemID,
		ApprovedBuyCarton: approvedBuyCarton,
		ApprovedBuyBox:    approvedBuyBox,
		ApprovedBuyPcs:    approvedBuyPcs,
		BuyerNotes:        form.BuyerNotes,
		ReviewedBy:        extractCurrentUserID(c),
		UpdatedBy:         extractCurrentUserID(c),
	})
	if err != nil {
		if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", err.Error()))
			return
		}
		renderStockCheckSessionDetailPage(c, service, sessionID, "", err.Error(), models.StockCheckSessionReviewItemEditForm{
			ItemID:            form.ItemID,
			ApprovedBuyCarton: form.ApprovedBuyCarton,
			ApprovedBuyBox:    form.ApprovedBuyBox,
			ApprovedBuyPcs:    form.ApprovedBuyPcs,
			ApprovedBuyQty:    form.ApprovedBuyQty,
			BuyerNotes:        form.BuyerNotes,
		})
		return
	}

	if redirectTo := sanitizeRedirectTarget(form.RedirectTo); redirectTo != "" {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "success", "Item review berhasil diperbarui"))
		return
	}

	c.Redirect(http.StatusSeeOther, buildStockCheckSessionDetailPageURL(sessionID, parsePositiveInt(c.Query("page"), 1), "Item review berhasil diperbarui"))
}

func StockCheckSessionApplyAllSubmitted(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil || sessionID <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	service := buildStockCheckSessionService()
	session, err := service.Repo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.String(http.StatusNotFound, "stock check session tidak ditemukan")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	redirectTo := sanitizeRedirectTarget(c.PostForm("redirect_to"))
	if redirectTo == "" {
		redirectTo = buildStockCheckSessionDetailPageURL(sessionID, parsePositiveInt(c.Query("page"), 1), "")
	}

	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, session.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", "Anda Tidak punya Akses di Halaman ini"))
		return
	}

	appliedCount, err := service.ApplyAllSubmittedBySession(sessionID, extractCurrentUserID(c))
	if err != nil {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "error", err.Error()))
		return
	}

	if appliedCount == 0 {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "success", "Tidak ada item submitted yang perlu di-apply"))
		return
	}

	c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTo, "success", fmt.Sprintf("%d item submitted berhasil di-apply", appliedCount)))
}

func StockCheckSessionStore(c *gin.Context) {
	type stockCheckSessionForm struct {
		SessionDate    string `form:"session_date" binding:"required"`
		StoreID        int    `form:"store_id" binding:"required"`
		SupplierID     int    `form:"supplier_id" binding:"required"`
		InitiationType string `form:"initiation_type" binding:"required"`
		Status         string `form:"status" binding:"required"`
		Notes          string `form:"notes"`
		ReturnTo       string `form:"return_to"`
		BackTo         string `form:"back_to"`
	}

	var form stockCheckSessionForm
	service := buildStockCheckSessionService()
	filter := buildStockCheckSessionFilter(c)
	formSession := models.StockCheckSession{}
	redirectTarget := ""
	backURL := ""

	if err := c.ShouldBind(&form); err != nil {
		formSession = models.StockCheckSession{
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}
		redirectTarget = sanitizeRedirectTarget(form.ReturnTo)
		backURL = sanitizeRedirectTarget(form.BackTo)
		if backURL == "" {
			backURL = buildStockCheckCheckerDefaultBackURL(form.SupplierID)
		}
		if isStockCheckCheckerDetailRedirectTarget(redirectTarget, form.SupplierID) {
			renderStockCheckCheckerDetailPage(
				c,
				buildSupplierService(),
				service,
				form.SupplierID,
				buildCheckerDetailFilterFromRedirectTarget(redirectTarget, form.SupplierID),
				"Form stock check session tidak lengkap",
				"",
				"create",
				formSession,
				extractCurrentUserID(c),
			)
			return
		}
		if isStockCheckCheckerCreateRedirectTarget(redirectTarget, form.SupplierID) {
			renderStockCheckCheckerCreateSessionPage(
				c,
				buildSupplierService(),
				service,
				form.SupplierID,
				"Form stock check session tidak lengkap",
				formSession,
				extractCurrentUserID(c),
				backURL,
			)
			return
		}
		renderStockCheckSessionPage(c, service, "Form stock check session tidak lengkap", "create", formSession, filter)
		return
	}

	formSession = models.StockCheckSession{
		SessionDate:    form.SessionDate,
		StoreID:        form.StoreID,
		SupplierID:     form.SupplierID,
		InitiationType: form.InitiationType,
		Status:         form.Status,
		Notes:          form.Notes,
	}
	redirectTarget = sanitizeRedirectTarget(form.ReturnTo)
	backURL = sanitizeRedirectTarget(form.BackTo)
	if backURL == "" {
		backURL = buildStockCheckCheckerDefaultBackURL(form.SupplierID)
	}

	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, form.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		if isStockCheckCheckerDetailRedirectTarget(redirectTarget, form.SupplierID) {
			renderStockCheckCheckerDetailPage(
				c,
				buildSupplierService(),
				service,
				form.SupplierID,
				buildCheckerDetailFilterFromRedirectTarget(redirectTarget, form.SupplierID),
				"Store tidak tersedia untuk user login",
				"",
				"create",
				formSession,
				extractCurrentUserID(c),
			)
			return
		}
		if isStockCheckCheckerCreateRedirectTarget(redirectTarget, form.SupplierID) {
			renderStockCheckCheckerCreateSessionPage(
				c,
				buildSupplierService(),
				service,
				form.SupplierID,
				"Store tidak tersedia untuk user login",
				formSession,
				extractCurrentUserID(c),
				backURL,
			)
			return
		}
		renderStockCheckSessionPage(c, service, "Store tidak tersedia untuk user login", "create", formSession, filter)
		return
	}

	sessionID, err := service.CreateSession(models.StockCheckSessionCreateInput{
		SessionDate:    form.SessionDate,
		StoreID:        form.StoreID,
		SupplierID:     form.SupplierID,
		InitiationType: form.InitiationType,
		Status:         form.Status,
		Notes:          form.Notes,
		CreatedBy:      extractCurrentUserID(c),
	})
	if err != nil {
		if isStockCheckCheckerDetailRedirectTarget(redirectTarget, form.SupplierID) {
			renderStockCheckCheckerDetailPage(
				c,
				buildSupplierService(),
				service,
				form.SupplierID,
				buildCheckerDetailFilterFromRedirectTarget(redirectTarget, form.SupplierID),
				err.Error(),
				"",
				"create",
				formSession,
				extractCurrentUserID(c),
			)
			return
		}
		if isStockCheckCheckerCreateRedirectTarget(redirectTarget, form.SupplierID) {
			renderStockCheckCheckerCreateSessionPage(
				c,
				buildSupplierService(),
				service,
				form.SupplierID,
				err.Error(),
				formSession,
				extractCurrentUserID(c),
				backURL,
			)
			return
		}
		renderStockCheckSessionPage(c, service, err.Error(), "create", formSession, filter)
		return
	}

	if isStockCheckCheckerDetailRedirectTarget(redirectTarget, form.SupplierID) || isStockCheckCheckerCreateRedirectTarget(redirectTarget, form.SupplierID) {
		c.Redirect(http.StatusSeeOther, buildStockCheckCheckerSessionInputPageURL(sessionID, "store", "Session berhasil dibuat. Silakan mulai input item."))
		return
	}

	c.Redirect(http.StatusSeeOther, "/stock-check-sessions")
}

func StockCheckSessionUpdate(c *gin.Context) {
	type stockCheckSessionForm struct {
		ID             int    `form:"id" binding:"required"`
		SessionNumber  string `form:"session_number_display"`
		SessionDate    string `form:"session_date" binding:"required"`
		StoreID        int    `form:"store_id" binding:"required"`
		SupplierID     int    `form:"supplier_id" binding:"required"`
		InitiationType string `form:"initiation_type" binding:"required"`
		Status         string `form:"status" binding:"required"`
		Notes          string `form:"notes"`
		ReturnTo       string `form:"return_to"`
	}

	var form stockCheckSessionForm
	service := buildStockCheckSessionService()
	filter := buildStockCheckSessionFilter(c)
	redirectTarget := ""

	if err := c.ShouldBind(&form); err != nil {
		redirectTarget = sanitizeRedirectTarget(form.ReturnTo)
		if redirectTarget != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTarget, "error", "Form edit stock check session tidak lengkap"))
			return
		}

		renderStockCheckSessionPage(c, service, "Form edit stock check session tidak lengkap", "edit", models.StockCheckSession{
			ID:             form.ID,
			SessionNumber:  form.SessionNumber,
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}
	redirectTarget = sanitizeRedirectTarget(form.ReturnTo)

	existingSession, err := service.Repo.GetByID(form.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if redirectTarget != "" {
				c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTarget, "error", "stock check session tidak ditemukan"))
				return
			}
			c.String(http.StatusNotFound, "stock check session tidak ditemukan")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	hasExistingStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, existingSession.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasExistingStoreAccess {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"code_error": 3,
			"error":      "Anda Tidak punya Akses di Halaman ini",
		})
		return
	}

	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, form.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		if redirectTarget != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTarget, "error", "Store tidak tersedia untuk user login"))
			return
		}

		renderStockCheckSessionPage(c, service, "Store tidak tersedia untuk user login", "edit", models.StockCheckSession{
			ID:             form.ID,
			SessionNumber:  form.SessionNumber,
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}

	err = service.UpdateSession(models.StockCheckSessionUpdateInput{
		ID:             form.ID,
		SessionDate:    form.SessionDate,
		StoreID:        form.StoreID,
		SupplierID:     form.SupplierID,
		InitiationType: form.InitiationType,
		Status:         form.Status,
		Notes:          form.Notes,
	})
	if err != nil {
		if redirectTarget != "" {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTarget, "error", err.Error()))
			return
		}

		renderStockCheckSessionPage(c, service, err.Error(), "edit", models.StockCheckSession{
			ID:             form.ID,
			SessionNumber:  form.SessionNumber,
			SessionDate:    form.SessionDate,
			StoreID:        form.StoreID,
			SupplierID:     form.SupplierID,
			InitiationType: form.InitiationType,
			Status:         form.Status,
			Notes:          form.Notes,
		}, filter)
		return
	}

	if redirectTarget != "" {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(redirectTarget, "success", "Status session berhasil diperbarui"))
		return
	}

	c.Redirect(http.StatusSeeOther, "/stock-check-sessions")
}

func StockCheckSessionDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid stock check session id")
		return
	}

	service := buildStockCheckSessionService()
	session, err := service.Repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.String(http.StatusNotFound, "stock check session tidak ditemukan")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, session.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"code_error": 3,
			"error":      "Anda Tidak punya Akses di Halaman ini",
		})
		return
	}

	if err := service.DeleteSession(id); err != nil {
		renderStockCheckSessionPage(c, service, err.Error(), "", models.StockCheckSession{}, buildStockCheckSessionFilter(c))
		return
	}

	c.Redirect(http.StatusSeeOther, "/stock-check-sessions")
}

func buildStockCheckSessionService() *services.StockCheckSessionService {
	repo := &repositories.StockCheckSessionRepository{DB: config.DB}
	return &services.StockCheckSessionService{Repo: repo}
}

func renderStockCheckSessionPage(c *gin.Context, service *services.StockCheckSessionService, message string, formMode string, formSession models.StockCheckSession, filter models.StockCheckSessionListFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	isSuperAdmin := currentUserHasRole(c, "super-admin")
	currentUserID := extractCurrentUserID(c)

	var stores []models.Store
	var err error
	if isSuperAdmin {
		stores, err = service.GetStoreOptions()
	} else {
		stores, err = service.GetStoreOptionsByUserID(currentUserID)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	allowedStoreSet := buildStoreAccessSet(stores)
	if !isSuperAdmin && filter.StoreID > 0 && !isStoreAccessible(allowedStoreSet, filter.StoreID) {
		filter.StoreID = 0
	}

	suppliers, err := service.GetSupplierOptions()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !isSuperAdmin {
		suppliers = filterSuppliersByStoreAccess(suppliers, allowedStoreSet)
	}

	sessions := []models.StockCheckSession{}
	totalItems := 0

	if isSuperAdmin || filter.StoreID > 0 {
		sessions, totalItems, err = service.GetSessions(filter)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		if !isSuperAdmin {
			filteredSessions := make([]models.StockCheckSession, 0, len(sessions))
			for _, session := range sessions {
				if !isStoreAccessible(allowedStoreSet, session.StoreID) {
					continue
				}
				filteredSessions = append(filteredSessions, session)
			}
			sessions = filteredSessions
			totalItems = len(filteredSessions)
		}
	} else {
		fetchFilter := filter
		fetchFilter.Page = 1
		fetchFilter.Limit = 2000

		sessionsRaw, _, err := service.GetSessions(fetchFilter)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		accessibleSessions := make([]models.StockCheckSession, 0, len(sessionsRaw))
		for _, session := range sessionsRaw {
			if !isStoreAccessible(allowedStoreSet, session.StoreID) {
				continue
			}
			accessibleSessions = append(accessibleSessions, session)
		}

		totalItems = len(accessibleSessions)
		totalPages := 0
		if totalItems > 0 {
			totalPages = (totalItems + filter.Limit - 1) / filter.Limit
			if filter.Page > totalPages {
				filter.Page = totalPages
			}
		}

		startIndex := (filter.Page - 1) * filter.Limit
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := startIndex + filter.Limit
		if endIndex > totalItems {
			endIndex = totalItems
		}

		if startIndex < endIndex {
			sessions = accessibleSessions[startIndex:endIndex]
		}
	}

	if formMode == "create" && !isSuperAdmin && formSession.StoreID > 0 && !isStoreAccessible(allowedStoreSet, formSession.StoreID) {
		formSession.StoreID = 0
	}
	if formMode == "edit" && !isSuperAdmin && formSession.StoreID > 0 && !isStoreAccessible(allowedStoreSet, formSession.StoreID) {
		formSession.StoreID = 0
	}

	pagination := buildStockCheckSessionPagination(filter, totalItems)

	Render(c, "stock_check_sessions.html", gin.H{
		"Title":       "Stock Check Sessions",
		"Page":        "stock_check_sessions",
		"Sessions":    sessions,
		"Stores":      stores,
		"Suppliers":   suppliers,
		"Filters":     filter,
		"Pagination":  pagination,
		"Error":       message,
		"FormMode":    formMode,
		"FormSession": formSession,
	})
}

func renderStockCheckCheckerSupplierPage(c *gin.Context, supplierService *services.SupplierService, message string, filter models.SupplierListFilter) {
	filter.Status = "active"
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Sort == "" {
		filter.Sort = "recent"
	}
	if filter.DayOfWeek < 1 || filter.DayOfWeek > 7 {
		filter.DayOfWeek = 0
	}
	filter.LastSODate = sanitizeQueryDate(filter.LastSODate)

	suppliers, totalItems, err := supplierService.GetSuppliers(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	pagination := buildStockCheckCheckerSupplierPagination(filter, totalItems)

	Render(c, "stock_check_checker_suppliers.html", gin.H{
		"Title":      "Input SO Checker",
		"Page":       "stock_check_checker",
		"Suppliers":  suppliers,
		"Filters":    filter,
		"Pagination": pagination,
		"TotalItems": totalItems,
		"Error":      message,
	})
}

func renderStockCheckCheckerDetailPage(c *gin.Context, supplierService *services.SupplierService, sessionService *services.StockCheckSessionService, supplierID int, filter models.StockCheckSessionListFilter, errorMessage string, successMessage string, formMode string, formSession models.StockCheckSession, currentUserID int) {
	supplier, err := supplierService.GetSupplierByID(supplierID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Supplier tidak ditemukan",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	filter.SupplierID = supplierID
	filter.Status = sanitizeStockCheckSessionStatusFilter(filter.Status)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	stores, err := sessionService.GetStoreOptionsByUserID(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	allowedStoreSet := buildStoreAccessSet(stores)

	fetchFilter := filter
	fetchFilter.Page = 1
	fetchFilter.Limit = 500
	sessionsRaw, _, err := sessionService.GetSessions(fetchFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	accessibleSessions := make([]models.StockCheckSession, 0, len(sessionsRaw))
	for _, session := range sessionsRaw {
		if !isStoreAccessible(allowedStoreSet, session.StoreID) {
			continue
		}
		accessibleSessions = append(accessibleSessions, session)
	}

	todayKey := time.Now().Format("2006-01-02")
	hasTodaySession := false
	todayFilter := models.StockCheckSessionListFilter{
		SupplierID: supplierID,
		DateFrom:   todayKey,
		DateTo:     todayKey,
		Page:       1,
		Limit:      200,
	}
	todaySessionsRaw, _, err := sessionService.GetSessions(todayFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	for _, session := range todaySessionsRaw {
		if !isStoreAccessible(allowedStoreSet, session.StoreID) {
			continue
		}
		hasTodaySession = true
		break
	}

	totalItems := len(accessibleSessions)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + filter.Limit - 1) / filter.Limit
		if filter.Page > totalPages {
			filter.Page = totalPages
		}
	}

	startIndex := (filter.Page - 1) * filter.Limit
	if startIndex < 0 {
		startIndex = 0
	}
	endIndex := startIndex + filter.Limit
	if endIndex > totalItems {
		endIndex = totalItems
	}

	sessions := make([]models.StockCheckSession, 0)
	if startIndex < endIndex {
		sessions = accessibleSessions[startIndex:endIndex]
	}

	if formMode != "create" {
		formSession = buildDefaultStockCheckCheckerCreateForm(supplierID, stores)
	} else {
		formSession = applyDefaultStockCheckCheckerCreateForm(formSession, supplierID, stores)
	}

	pagination := buildStockCheckCheckerSessionPagination(supplierID, filter, totalItems)
	currentURL := buildStockCheckCheckerDetailPageURL(supplierID, filter, filter.Page)

	Render(c, "stock_check_checker_detail.html", gin.H{
		"Title":           supplier.SupplierName,
		"Page":            "stock_check_checker",
		"Supplier":        supplier,
		"Sessions":        sessions,
		"Filters":         filter,
		"Pagination":      pagination,
		"TotalItems":      totalItems,
		"Stores":          stores,
		"Error":           errorMessage,
		"Success":         successMessage,
		"FormMode":        formMode,
		"FormSession":     formSession,
		"CurrentPath":     c.Request.URL.Path,
		"CurrentURL":      currentURL,
		"CreateURL":       buildStockCheckCheckerSessionCreatePageURL(supplierID, currentURL),
		"HasTodaySession": hasTodaySession,
	})
}

func renderStockCheckCheckerCreateSessionPage(c *gin.Context, supplierService *services.SupplierService, sessionService *services.StockCheckSessionService, supplierID int, errorMessage string, formSession models.StockCheckSession, currentUserID int, backURL string) {
	supplier, err := supplierService.GetSupplierByID(supplierID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Supplier tidak ditemukan",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	stores, err := sessionService.GetStoreOptionsByUserID(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if backURL == "" {
		backURL = buildStockCheckCheckerDefaultBackURL(supplierID)
	}

	formSession = applyDefaultStockCheckCheckerCreateForm(formSession, supplierID, stores)
	currentURL := buildStockCheckCheckerSessionCreatePageURL(supplierID, backURL)

	Render(c, "stock_check_checker_session_create.html", gin.H{
		"Title":       "Tambah Session Baru",
		"Page":        "stock_check_checker",
		"Supplier":    supplier,
		"Stores":      stores,
		"Error":       errorMessage,
		"FormSession": formSession,
		"BackURL":     backURL,
		"CurrentPath": c.Request.URL.Path,
		"CurrentURL":  currentURL,
	})
}

func renderStockCheckCheckerSessionInputPage(c *gin.Context, service *services.StockCheckSessionService, sessionID int, successMessage string, errorMessage string, scanForm models.StockCheckSessionCheckerScanForm) {
	pageData, err := service.GetCheckerInputPage(sessionID, extractCurrentUserID(c))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Stock check session tidak ditemukan",
			})
			return
		}
		if strings.Contains(strings.ToLower(err.Error()), "tidak tersedia untuk user login") {
			c.String(http.StatusForbidden, err.Error())
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if scanForm.Location == "" {
		scanForm.Location = "store"
	}

	Render(c, "stock_check_checker_session_input.html", gin.H{
		"Title":             pageData.Session.SessionNumber,
		"MobileHeaderTitle": "Check Stok",
		"Page":              "stock_check_checker",
		"Session":           pageData.Session,
		"Items":             pageData.Items,
		"Success":           successMessage,
		"Error":             errorMessage,
		"ScanForm":          scanForm,
		"CurrentURL":        buildStockCheckCheckerSessionInputPageURL(sessionID, scanForm.Location, ""),
	})
}

func renderStockCheckCheckerSessionScanPage(c *gin.Context, service *services.StockCheckSessionService, sessionID int, barcode string, location string, successMessage string, errorMessage string, scanForm models.StockCheckSessionCheckerScanForm, backURL string) {
	barcode = strings.TrimSpace(barcode)
	location = sanitizeStockCheckCheckerScanLocation(location)
	if location == "" {
		location = "store"
	}
	if backURL == "" {
		backURL = buildStockCheckCheckerSessionInputPageURL(sessionID, location, "")
	}
	if barcode == "" {
		c.Redirect(http.StatusSeeOther, appendRedirectMessage(backURL, "error", "Barcode wajib diisi"))
		return
	}

	pageData, err := service.GetCheckerScanPage(sessionID, extractCurrentUserID(c), barcode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Redirect(http.StatusSeeOther, appendRedirectMessage(buildStockCheckCheckerSessionInputPageURL(sessionID, location, ""), "error", "Barcode Item ini tidak ada di supplier ini."))
			return
		}
		if strings.Contains(strings.ToLower(err.Error()), "tidak tersedia untuk user login") {
			c.String(http.StatusForbidden, err.Error())
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if scanForm.Location == "" {
		scanForm.Location = location
	}
	if strings.TrimSpace(scanForm.Barcode) == "" {
		scanForm.Barcode = barcode
	}
	if location == "warehouse" {
		if strings.TrimSpace(scanForm.QtyCarton) == "" {
			scanForm.QtyCarton = strconv.Itoa(pageData.Item.QtyWarehouseCarton)
		}
		if strings.TrimSpace(scanForm.QtyBox) == "" {
			scanForm.QtyBox = strconv.Itoa(pageData.Item.QtyWarehouseBox)
		}
		if strings.TrimSpace(scanForm.QtyPcs) == "" {
			scanForm.QtyPcs = strconv.Itoa(pageData.Item.QtyWarehousePcs)
		}
	} else {
		if strings.TrimSpace(scanForm.QtyCarton) == "" {
			scanForm.QtyCarton = strconv.Itoa(pageData.Item.QtyStoreCarton)
		}
		if strings.TrimSpace(scanForm.QtyBox) == "" {
			scanForm.QtyBox = strconv.Itoa(pageData.Item.QtyStoreBox)
		}
		if strings.TrimSpace(scanForm.QtyPcs) == "" {
			scanForm.QtyPcs = strconv.Itoa(pageData.Item.QtyStorePcs)
		}
	}

	Render(c, "stock_check_checker_session_scan.html", gin.H{
		"Title":       "Input Qty Item",
		"Page":        "stock_check_checker",
		"Session":     pageData.Session,
		"Item":        pageData.Item,
		"Items":       pageData.Items,
		"Success":     successMessage,
		"Error":       errorMessage,
		"ScanForm":    scanForm,
		"BackURL":     backURL,
		"CurrentURL":  buildStockCheckCheckerSessionScanPageURL(sessionID, location, scanForm.Barcode, backURL),
		"CurrentPath": c.Request.URL.Path,
	})
}

func renderStockCheckSessionDetailPage(c *gin.Context, service *services.StockCheckSessionService, id int, successMessage string, errorMessage string, reviewForm models.StockCheckSessionReviewItemEditForm) {
	currentPage := 1
	pageData, err := service.GetSessionDetailPage(id, currentPage, stockCheckSessionDetailItemLimit, models.StockCheckSessionDetailFilter{
		SortBy: c.Query("sort_by"),
		Status: c.Query("item_status"),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"code_error": http.StatusNotFound,
				"error":      "Stock check session tidak ditemukan",
			})
			return
		}

		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	hasStoreAccess, err := currentUserCanAccessStockCheckStore(c, service, pageData.Session.StoreID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStoreAccess {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"code_error": 3,
			"error":      "Anda Tidak punya Akses di Halaman ini",
		})
		return
	}

	pageData.Pagination = buildStockCheckSessionDetailPagination(id, pageData.Pagination)

	Render(c, "stock_check_session_detail.html", gin.H{
		"Title":       pageData.Session.SessionNumber,
		"Page":        "stock_check_sessions",
		"Session":     pageData.Session,
		"Items":       pageData.Items,
		"Overview":    pageData.OverviewCards,
		"Pagination":  pageData.Pagination,
		"Filters":     pageData.Filters,
		"Success":     successMessage,
		"Error":       errorMessage,
		"ReviewForm":  reviewForm,
		"CurrentPath": c.Request.URL.Path,
		"CurrentURL":  c.Request.URL.RequestURI(),
	})
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseStockCheckNonNegativeInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid qty")
	}

	return parsed, nil
}

func buildStockCheckSessionDetailPagination(sessionID int, pagination models.Pagination) models.Pagination {
	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		if pagination.TotalItems > 0 {
			pagination.PageSize = pagination.TotalItems
		} else {
			pagination.PageSize = 1
		}
	}
	if pagination.TotalItems == 0 {
		return pagination
	}

	if pagination.TotalPages <= 0 {
		pagination.TotalPages = (pagination.TotalItems + pagination.PageSize - 1) / pagination.PageSize
	}
	if pagination.CurrentPage > pagination.TotalPages {
		pagination.CurrentPage = pagination.TotalPages
	}

	pagination.StartItem = ((pagination.CurrentPage - 1) * pagination.PageSize) + 1
	pagination.EndItem = pagination.StartItem + pagination.PageSize - 1
	if pagination.EndItem > pagination.TotalItems {
		pagination.EndItem = pagination.TotalItems
	}

	pagination.HasPrev = pagination.CurrentPage > 1
	pagination.HasNext = pagination.CurrentPage < pagination.TotalPages
	if pagination.HasPrev {
		pagination.PrevURL = buildStockCheckSessionDetailPageURL(sessionID, pagination.CurrentPage-1, "")
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckSessionDetailPageURL(sessionID, pagination.CurrentPage+1, "")
	}

	startPage := pagination.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 4
	if endPage > pagination.TotalPages {
		endPage = pagination.TotalPages
	}
	if endPage-startPage < 4 {
		startPage = endPage - 4
		if startPage < 1 {
			startPage = 1
		}
	}

	pagination.Pages = nil
	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockCheckSessionDetailPageURL(sessionID, page, ""),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckSessionDetailPageURL(sessionID int, page int, successMessage string) string {
	values := url.Values{}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	if successMessage != "" {
		values.Set("success", successMessage)
	}

	baseURL := fmt.Sprintf("/stock-check-sessions/%d", sessionID)
	encoded := values.Encode()
	if encoded == "" {
		return baseURL
	}
	return baseURL + "?" + encoded
}

func buildStockCheckCheckerSessionInputPageURL(sessionID int, location string, successMessage string) string {
	values := url.Values{}
	location = sanitizeStockCheckCheckerScanLocation(location)
	if location != "" {
		values.Set("location", location)
	}
	if successMessage != "" {
		values.Set("success", successMessage)
	}

	baseURL := fmt.Sprintf("/stock-checker/sessions/%d/input", sessionID)
	encoded := values.Encode()
	if encoded == "" {
		return baseURL
	}
	return baseURL + "?" + encoded
}

func buildStockCheckCheckerSessionScanPageURL(sessionID int, location string, barcode string, backURL string) string {
	values := url.Values{}
	location = sanitizeStockCheckCheckerScanLocation(location)
	if location != "" {
		values.Set("location", location)
	}
	barcode = strings.TrimSpace(barcode)
	if barcode != "" {
		values.Set("barcode", barcode)
	}
	backURL = sanitizeRedirectTarget(backURL)
	if backURL != "" {
		values.Set("back_to", backURL)
	}

	baseURL := fmt.Sprintf("/stock-checker/sessions/%d/scan", sessionID)
	encoded := values.Encode()
	if encoded == "" {
		return baseURL
	}
	return baseURL + "?" + encoded
}

func sanitizeRedirectTarget(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return ""
	}
	return value
}

func appendRedirectMessage(target string, key string, message string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return target
	}
	values := parsed.Query()
	values.Set(key, message)
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func isStockCheckCheckerSessionScanRedirectTarget(target string, sessionID int) bool {
	if sessionID <= 0 || target == "" {
		return false
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}

	return parsed.Path == fmt.Sprintf("/stock-checker/sessions/%d/scan", sessionID)
}

func isStockCheckCheckerDetailRedirectTarget(target string, supplierID int) bool {
	if supplierID <= 0 || target == "" {
		return false
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}

	return parsed.Path == fmt.Sprintf("/stock-checker/%d", supplierID)
}

func isStockCheckCheckerCreateRedirectTarget(target string, supplierID int) bool {
	if supplierID <= 0 || target == "" {
		return false
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}

	return parsed.Path == fmt.Sprintf("/stock-checker/%d/sessions/new", supplierID)
}

func buildCheckerDetailFilterFromRedirectTarget(target string, supplierID int) models.StockCheckSessionListFilter {
	filter := models.StockCheckSessionListFilter{
		SupplierID: supplierID,
		Page:       1,
		Limit:      10,
	}
	if target == "" {
		return filter
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return filter
	}

	values := parsed.Query()
	filter.Status = sanitizeStockCheckSessionStatusFilter(values.Get("status"))
	filter.Page = parsePositiveInt(values.Get("page"), 1)

	return filter
}

func sanitizeStockCheckCheckerScanLocation(value string) string {
	switch strings.TrimSpace(value) {
	case "store", "warehouse":
		return value
	default:
		return ""
	}
}

func buildDefaultStockCheckCheckerCreateForm(supplierID int, stores []models.Store) models.StockCheckSession {
	form := models.StockCheckSession{
		SessionDate:    time.Now().Format("2006-01-02"),
		SupplierID:     supplierID,
		InitiationType: "checker_initiative",
		Status:         "in_progress",
	}
	if len(stores) == 1 {
		form.StoreID = stores[0].StoreID
	}
	return form
}

func applyDefaultStockCheckCheckerCreateForm(form models.StockCheckSession, supplierID int, stores []models.Store) models.StockCheckSession {
	if strings.TrimSpace(form.SessionDate) == "" {
		form.SessionDate = time.Now().Format("2006-01-02")
	}
	if form.SupplierID <= 0 {
		form.SupplierID = supplierID
	}
	if strings.TrimSpace(form.InitiationType) == "" {
		form.InitiationType = "checker_initiative"
	}
	if strings.TrimSpace(form.Status) == "" {
		form.Status = "in_progress"
	}
	if form.StoreID <= 0 && len(stores) == 1 {
		form.StoreID = stores[0].StoreID
	}
	return form
}

func sanitizeStockCheckSessionStatusFilter(value string) string {
	switch strings.TrimSpace(value) {
	case "draft", "in_progress", "submitted", "reviewed", "closed", "po", "cancelled":
		return value
	default:
		return ""
	}
}

func buildStockCheckSessionFilter(c *gin.Context) models.StockCheckSessionListFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	storeID, _ := strconv.Atoi(c.DefaultQuery("store_id", "0"))
	supplierID, _ := strconv.Atoi(c.DefaultQuery("supplier_id", "0"))

	return models.StockCheckSessionListFilter{
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		StoreID:    storeID,
		SupplierID: supplierID,
		Status:     c.Query("status"),
		Page:       page,
		Limit:      10,
	}
}

func extractCurrentUserID(c *gin.Context) int {
	sess := sessions.Default(c)

	if raw := sess.Get("user_id"); raw != nil {
		switch id := raw.(type) {
		case int:
			return id
		case int64:
			return int(id)
		case float64:
			return int(id)
		}
	}

	if raw := sess.Get("user"); raw != nil {
		switch user := raw.(type) {
		case models.SessionUser:
			return user.UserID
		case map[string]interface{}:
			if id, ok := user["user_id"].(float64); ok {
				return int(id)
			}
		case gin.H:
			if id, ok := user["user_id"].(float64); ok {
				return int(id)
			}
		}
	}

	return 0
}

func buildStockCheckSessionPagination(filter models.StockCheckSessionListFilter, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: filter.Page,
		PageSize:    filter.Limit,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}
	if totalItems == 0 {
		return pagination
	}

	pagination.TotalPages = (totalItems + pagination.PageSize - 1) / pagination.PageSize
	if pagination.CurrentPage > pagination.TotalPages {
		pagination.CurrentPage = pagination.TotalPages
	}

	pagination.StartItem = ((pagination.CurrentPage - 1) * pagination.PageSize) + 1
	pagination.EndItem = pagination.StartItem + pagination.PageSize - 1
	if pagination.EndItem > totalItems {
		pagination.EndItem = totalItems
	}

	pagination.HasPrev = pagination.CurrentPage > 1
	pagination.HasNext = pagination.CurrentPage < pagination.TotalPages
	if pagination.HasPrev {
		pagination.PrevURL = buildStockCheckSessionPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckSessionPageURL(filter, pagination.CurrentPage+1)
	}

	startPage := pagination.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 2
	if endPage > pagination.TotalPages {
		endPage = pagination.TotalPages
	}
	if endPage-startPage < 2 {
		startPage = endPage - 2
		if startPage < 1 {
			startPage = 1
		}
	}

	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockCheckSessionPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckSessionPageURL(filter models.StockCheckSessionListFilter, page int) string {
	values := url.Values{}
	if filter.DateFrom != "" {
		values.Set("date_from", filter.DateFrom)
	}
	if filter.DateTo != "" {
		values.Set("date_to", filter.DateTo)
	}
	if filter.StoreID > 0 {
		values.Set("store_id", strconv.Itoa(filter.StoreID))
	}
	if filter.SupplierID > 0 {
		values.Set("supplier_id", strconv.Itoa(filter.SupplierID))
	}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}

	encoded := values.Encode()
	if encoded == "" {
		return "/stock-check-sessions"
	}
	return "/stock-check-sessions?" + encoded
}

func currentUserCanAccessStockCheckStore(c *gin.Context, service *services.StockCheckSessionService, storeID int) (bool, error) {
	if currentUserHasRole(c, "super-admin") {
		return true, nil
	}

	userID := extractCurrentUserID(c)
	if userID <= 0 {
		return false, nil
	}

	hasStoreAccess, err := service.Repo.UserHasStoreAccess(userID, storeID)
	if err != nil {
		return false, err
	}
	if hasStoreAccess {
		return true, nil
	}

	return service.Repo.UserHasRole(userID, "checker")
}

func filterSuppliersByStoreAccess(suppliers []models.Supplier, allowedStores map[int]struct{}) []models.Supplier {
	if len(suppliers) == 0 {
		return []models.Supplier{}
	}

	filtered := make([]models.Supplier, 0, len(suppliers))
	for _, supplier := range suppliers {
		if isStoreAccessible(allowedStores, supplier.StoreID) {
			filtered = append(filtered, supplier)
		}
	}

	return filtered
}

func buildStockCheckCheckerSupplierPagination(filter models.SupplierListFilter, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: filter.Page,
		PageSize:    filter.Limit,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 50
	}
	if totalItems == 0 {
		return pagination
	}

	pagination.TotalPages = (totalItems + pagination.PageSize - 1) / pagination.PageSize
	if pagination.CurrentPage > pagination.TotalPages {
		pagination.CurrentPage = pagination.TotalPages
	}

	pagination.StartItem = ((pagination.CurrentPage - 1) * pagination.PageSize) + 1
	pagination.EndItem = pagination.StartItem + pagination.PageSize - 1
	if pagination.EndItem > totalItems {
		pagination.EndItem = totalItems
	}

	pagination.HasPrev = pagination.CurrentPage > 1
	pagination.HasNext = pagination.CurrentPage < pagination.TotalPages
	if pagination.HasPrev {
		pagination.PrevURL = buildStockCheckCheckerSupplierPageURL(filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckCheckerSupplierPageURL(filter, pagination.CurrentPage+1)
	}

	startPage := pagination.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 4
	if endPage > pagination.TotalPages {
		endPage = pagination.TotalPages
	}
	if endPage-startPage < 4 {
		startPage = endPage - 4
		if startPage < 1 {
			startPage = 1
		}
	}

	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockCheckCheckerSupplierPageURL(filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckCheckerSupplierPageURL(filter models.SupplierListFilter, page int) string {
	values := url.Values{}
	if filter.Search != "" {
		values.Set("search", filter.Search)
	}
	if filter.DayOfWeek > 0 {
		values.Set("day_of_week", strconv.Itoa(filter.DayOfWeek))
	}
	if filter.LastSODate != "" {
		values.Set("last_so_date", filter.LastSODate)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}

	encoded := values.Encode()
	if encoded == "" {
		return "/stock-checker"
	}
	return "/stock-checker?" + encoded
}

func buildStockCheckCheckerSessionPagination(supplierID int, filter models.StockCheckSessionListFilter, totalItems int) models.Pagination {
	pagination := models.Pagination{
		CurrentPage: filter.Page,
		PageSize:    filter.Limit,
		TotalItems:  totalItems,
	}

	if pagination.CurrentPage <= 0 {
		pagination.CurrentPage = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}
	if totalItems == 0 {
		return pagination
	}

	pagination.TotalPages = (totalItems + pagination.PageSize - 1) / pagination.PageSize
	if pagination.CurrentPage > pagination.TotalPages {
		pagination.CurrentPage = pagination.TotalPages
	}

	pagination.StartItem = ((pagination.CurrentPage - 1) * pagination.PageSize) + 1
	pagination.EndItem = pagination.StartItem + pagination.PageSize - 1
	if pagination.EndItem > totalItems {
		pagination.EndItem = totalItems
	}

	pagination.HasPrev = pagination.CurrentPage > 1
	pagination.HasNext = pagination.CurrentPage < pagination.TotalPages
	if pagination.HasPrev {
		pagination.PrevURL = buildStockCheckCheckerDetailPageURL(supplierID, filter, pagination.CurrentPage-1)
	}
	if pagination.HasNext {
		pagination.NextURL = buildStockCheckCheckerDetailPageURL(supplierID, filter, pagination.CurrentPage+1)
	}

	startPage := pagination.CurrentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 4
	if endPage > pagination.TotalPages {
		endPage = pagination.TotalPages
	}
	if endPage-startPage < 4 {
		startPage = endPage - 4
		if startPage < 1 {
			startPage = 1
		}
	}

	for page := startPage; page <= endPage; page++ {
		pagination.Pages = append(pagination.Pages, models.PaginationLink{
			Number: page,
			URL:    buildStockCheckCheckerDetailPageURL(supplierID, filter, page),
			Active: page == pagination.CurrentPage,
		})
	}

	return pagination
}

func buildStockCheckCheckerDetailPageURL(supplierID int, filter models.StockCheckSessionListFilter, page int) string {
	values := url.Values{}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}

	baseURL := fmt.Sprintf("/stock-checker/%d", supplierID)
	encoded := values.Encode()
	if encoded == "" {
		return baseURL
	}
	return baseURL + "?" + encoded
}

func buildStockCheckCheckerSessionCreatePageURL(supplierID int, backURL string) string {
	values := url.Values{}
	backURL = sanitizeRedirectTarget(backURL)
	if backURL != "" {
		values.Set("back_to", backURL)
	}

	baseURL := fmt.Sprintf("/stock-checker/%d/sessions/new", supplierID)
	encoded := values.Encode()
	if encoded == "" {
		return baseURL
	}
	return baseURL + "?" + encoded
}

func buildStockCheckCheckerDefaultBackURL(supplierID int) string {
	return fmt.Sprintf("/stock-checker/%d", supplierID)
}

func formatStockCheckPOWholeNumber(value float64) string {
	return fmt.Sprintf("%.0f", value)
}

func formatStockCheckPOBreakdown(carton int, box int, pcs int) string {
	return fmt.Sprintf("%dc - %db - %dp", carton, box, pcs)
}

func formatStockCheckPOCurrency(value float64) string {
	negative := value < 0
	if negative {
		value = -value
	}

	raw := fmt.Sprintf("%.2f", value)
	parts := strings.SplitN(raw, ".", 2)
	integerPart := parts[0]
	decimalPart := "00"
	if len(parts) == 2 {
		decimalPart = parts[1]
	}

	var grouped strings.Builder
	for idx, char := range integerPart {
		if idx > 0 && (len(integerPart)-idx)%3 == 0 {
			grouped.WriteRune(',')
		}
		grouped.WriteRune(char)
	}

	prefix := ""
	if negative {
		prefix = "-"
	}

	return prefix + "$" + grouped.String() + "." + decimalPart
}
