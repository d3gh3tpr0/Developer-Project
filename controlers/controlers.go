package controlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test/models"
)

// @BasePath /api/v1

// DevGet godoc
// @Summary ping dev
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success	200 {array} developer
// @Failure 404 {array} developer
// @Router /dev [get]
func GetDev(db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var req models.DevRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			query := "select * from developers"
			rows, err := db.Raw(query).Rows()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items := []models.Developer{}
			for rows.Next() {
				var item models.Developer
				if err := rows.Scan(&item.ID, &item.Name, &item.Language); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				items = append(items, item)
			}
			c.JSON(http.StatusOK, gin.H{"data_query": items})
			return
		}
		query := "select * from developers as d where d.language = ?"
		rows, err := db.Raw(query, req.Language).Rows()
		defer rows.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		items := []models.Developer{}
		for rows.Next() {
			var item models.Developer
			if err := rows.Scan(
				&item.ID,
				&item.Name,
				&item.Language,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, gin.H{"data_query": items})
	}
	return gin.HandlerFunc(fn)
}
