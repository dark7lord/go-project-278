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

// ListLinks handles listing all links.
func (h *Handler) ListLinks(c *gin.Context) {
	rangeParam := requestRange(c)

	sort, err := parseSortParam(c.Query("sort"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}
	if sort != nil {
		if rangeParam == "" {
			c.JSON(http.StatusBadRequest, errJSON(ErrSortWithoutRange.Error()))
			return
		}
		if _, ok := linksSortableFields[sort.Field]; !ok {
			c.JSON(http.StatusBadRequest, errJSON(application.ErrSortField.Error()))
			return
		}
	}

	if rangeParam == "" {
		links, err := h.linkService.ListLinks(c.Request.Context())
		if err != nil {
			writeServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, h.linkResponses(links))

		return
	}

	start, end, err := parseRangeParam(rangeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.linkService.PageLinks(
		c.Request.Context(),
		application.ListLinksQuery{Start: start, End: end, Sort: sort},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, "links", application.RangePage[linkResponse]{
		Items: h.linkResponses(page.Items),
		Start: page.Start,
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
