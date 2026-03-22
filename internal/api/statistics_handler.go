package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	db "github.com/moklidia/go-project-278/internal/db"
)

type LinkVisitResponse struct {
	LinkID    int64  `json:"link_id"`
	Ip        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Status    int32 `json:"status"`
	Referer   string `json:"referer"`
}

func GetStatistics(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var linkVisits []db.LinkVisit
		var err error
		var requestedRange *rangeBounds
		paginationRange := c.Query("range")

		if paginationRange != "" {
			pagination, bounds, err := getLimitAndOffsetFromQuery(paginationRange)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			requestedRange = &bounds

			linkVisits, err = queries.ListLinkVisits(c.Request.Context(), db.ListLinkVisitsParams{
				Limit:  pagination.Limit,
				Offset: pagination.Offset,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			linkVisits, err = queries.ListAllLinkVisits(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		total, err := queries.CountLinkVisits(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		contentRange := buildContentRange("link_visits", requestedRange, len(linkVisits), total)
		c.Header("Content-Range", contentRange)

		response := make([]LinkVisitResponse, 0, len(linkVisits))
		for _, lv := range linkVisits {
			response = append(response, toLinkVisitResponse(lv))
		}

		c.JSON(http.StatusOK, response)
	}
}

func toLinkVisitResponse(lv db.LinkVisit) LinkVisitResponse {
	return LinkVisitResponse{
		LinkID:    lv.LinkID,
		Ip:        lv.Ip,
		UserAgent: lv.UserAgent,
		Status:    lv.Status,
		Referer:   lv.Referer,
	}
}
