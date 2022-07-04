package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	mydocs "test/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"test/controlers"
	"test/models"
	// swagger embed files
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:7070
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

func main() {
	dsn := "root:example@tcp(127.0.0.1:3306)/manage_inventories?charset=utf8mb4"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	router := gin.Default()

	router.Static("/static_file", "./upload")
	router.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("conmemay")
		log.Println(file.Filename)

		c.SaveUploadedFile(file, "./upload/"+file.Filename)

		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})
	router.POST("/upload_multi", func(c *gin.Context) {

		form, _ := c.MultipartForm()
		files := form.File["files"]

		for _, file := range files {
			log.Println(file.Filename)

			c.SaveUploadedFile(file, "./upload/"+file.Filename)
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	})

	router.GET("/dev", controlers.GetDev(db))

	router.GET("/dev/:id", func(ctx *gin.Context) {
		var req models.DevIDRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "select * from developers where id = ?"
		row := db.Raw(query, req.ID).Row()
		var res models.Developer
		if err := row.Scan(
			&res.ID,
			&res.Name,
			&res.Language,
		); err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data_query": res})
	})
	router.POST("/dev", func(ctx *gin.Context) {
		var req models.DevCreateRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		queryFirst := "insert into developers (name, language) values (?, ?)"
		querySecond := "select * from developers where id = last_insert_id()"
		var res models.Developer
		arg := models.DevCreateParams{
			Name:     req.Name,
			Language: req.Language,
		}

		row := db.Exec(queryFirst, arg.Name, arg.Language).Raw(querySecond).Row()
		err := row.Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := row.Scan(
			&res.ID,
			&res.Name,
			&res.Language,
		); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"dev": res})

	})
	router.DELETE("/dev/:id", func(ctx *gin.Context) {
		var reqID models.DevIDRequest
		if err := ctx.ShouldBindUri(&reqID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "delete from developers where id = ?"
		err := db.Exec(query, reqID.ID).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	router.PATCH("/dev/:id", func(ctx *gin.Context) {
		var reqID models.DevIDRequest
		if err := ctx.ShouldBindUri(&reqID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var req models.DevRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "update developers set language = ? where id = ?"
		err := db.Exec(query, req.Language, reqID.ID).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result := fmt.Sprintf("ID %v has change language to %v successfully", reqID.ID, req.Language)
		ctx.JSON(http.StatusOK, gin.H{"status": result})
	})

	mydocs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		eg := v1.Group("/")
		{
			eg.GET("/dev", controlers.GetDev(db))
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))


	router.Run(":7070")

}
