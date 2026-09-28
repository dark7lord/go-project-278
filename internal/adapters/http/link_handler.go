package httpadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// redirectStatus is the single source for the response status and for the
// status recorded on the visit, so the two can never drift apart.
const redirectStatus = http.StatusFound

// CreateLink handles link creation.
func (h *Handler) CreateLink(c *gin.Context) {
	req, ok := bindLinkRequest(c)
	if !ok {
		return
	}

	cmd := application.CreateLinkCommand{OriginalURL: req.OriginalURL, ShortName: req.ShortName}
	link, err := h.linkService.CreateLink(c.Request.Context(), cmd)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusCreated, h.linkResponse(link))
}

// GetLink handles link retrieval by ID.
func (h *Handler) GetLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	link, err := h.linkService.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusOK, h.linkResponse(link))
}

// ListLinks handles listing a page of links.
func (h *Handler) ListLinks(c *gin.Context) {
	sort, err := parseSortParam(c.Query("sort"), linksSortFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	pageRange, fromHeader, err := requestRange(c, linksUnit)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.linkService.PageLinks(
		c.Request.Context(),
		application.ListLinksQuery{Range: pageRange, Sort: sort},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, linksUnit, fromHeader, application.RangePage[linkResponse]{
		Items: h.linkResponses(page.Items),
		First: page.First,
		Total: page.Total,
	})
}

// UpdateLink handles link updates.
func (h *Handler) UpdateLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	req, ok := bindLinkRequest(c)
	if !ok {
		return
	}

	cmd := application.UpdateLinkCommand{OriginalURL: req.OriginalURL, ShortName: req.ShortName}
	updated, err := h.linkService.UpdateLink(c.Request.Context(), id, cmd)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusOK, h.linkResponse(updated))
}

// DeleteLink handles link deletion.
func (h *Handler) DeleteLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	_, err = h.linkService.DeleteLink(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.Status(http.StatusNoContent)
}

// Redirect handles redirecting a short name to its original URL.
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	ctx := c.Request.Context()
	var referer *string
	if ref := c.Request.Referer(); ref != "" {
		referer = &ref
	}

	link, err := h.linkService.Redirect(ctx, application.RedirectCommand{
		ShortName: code,
		VisitMeta: application.VisitMeta{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Referer:   referer,
		},
		Status: int32(redirectStatus),
	})

	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.Redirect(redirectStatus, link.OriginalURL)
}
