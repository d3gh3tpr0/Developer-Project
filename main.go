package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type developer struct {
	ID       int
	Name     string
	Language string
}
type devRequest struct {
	Language string `form:"language" binding:"required"`
}
type devIDRequest struct {
	ID int `uri:"id" binding:"required,min=1"`
}
type devCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Language string `json:"language" binding:"required"`
}
type devCreateParams struct {
	Name     string `json:"name"`
	Language string `json:"language"`
}


func main() {
	dsn := "root:example@tcp(127.0.0.1:3306)/manage_inventories?charset=utf8mb4"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	router := gin.Default()

	router.Static("/static_file", "./upload")
	router.POST("/upload", func(c *gin.Context) {
		// Single file
		file, _ := c.FormFile("conmemay")
		log.Println(file.Filename)

		// Upload the file to specific dst.
		c.SaveUploadedFile(file, "./upload/"+file.Filename)

		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})
	router.POST("/upload_multi", func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["files"]

		for _, file := range files {
			log.Println(file.Filename)

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, "./upload/"+file.Filename)
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	})
	router.GET("/dev", func(c *gin.Context) {
		var req devRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			query := "select * from developers"
			rows, err := db.Raw(query).Rows()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items := []developer{}
			for rows.Next() {
				var item developer
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

		items := []developer{}
		for rows.Next() {
			var item developer
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
	})

	router.GET("/dev/:id", func(ctx *gin.Context) {
		var req devIDRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "select * from developers where id = ?"
		row := db.Raw(query, req.ID).Row()
		var res developer
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
		var req devCreateRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		queryFirst := "insert into developers (name, language) values (?, ?)"
		querySecond := "select * from developers where id = last_insert_id()"
		var res developer
		arg := devCreateParams{
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
	router.DELETE("/dev/:id", func(ctx *gin.Context){
		var reqID devIDRequest
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

	router.PATCH("/dev/:id", func(ctx *gin.Context){
		var reqID devIDRequest
		if err := ctx.ShouldBindUri(&reqID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} 
		var req devRequest
		if err := ctx.ShouldBindQuery(&req); err != nil  {
			ctx.JSON(http.StatusBadRequest,gin.H{"error": err.Error()})
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


	router.Run(":7070")

}
