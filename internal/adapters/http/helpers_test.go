package httpadapter

import (
	"github.com/gin-gonic/gin"

	"code/internal/application"
)

const (
	testBaseURL    = "https://short.example"
	testExampleURL = "https://example.com"
)

func newTestHandler(
	linkService application.LinkUseCase,
	visitService application.VisitUseCase,
) *Handler {
	return NewHandler(linkService, visitService, testBaseURL)
}

func newHandlerRouter(handler *Handler) *gin.Engine {
	router := gin.New()
	handler.RegisterRootRoutes(router)
	handler.RegisterAPIRoutes(router.Group(""))

	return router
}

func linksQuery(first, last int64, sort *application.Sort) application.ListLinksQuery {
	return application.ListLinksQuery{Range: application.Range{First: first, Last: last}, Sort: sort}
}

func visitsQuery(first, last int64, sort *application.Sort) application.ListLinkVisitsQuery {
	return application.ListLinkVisitsQuery{Range: application.Range{First: first, Last: last}, Sort: sort}
}

// firstPageLast ends the page a request without a range reads: "0-" cut to maxPageSize.
const firstPageLast = maxPageSize - 1
