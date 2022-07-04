package models
import "fmt"

type Developer struct {
	ID       int
	Name     string
	Language string
}

type DevRequest struct {
	Language string `form:"language" binding:"required"`
}

type DevIDRequest struct {
	ID int `uri:"id" binding:"required,min=1"`
}
type DevCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Language string `json:"language" binding:"required"`
}
type DevCreateParams struct {
	Name     string `json:"name"`
	Language string `json:"language"`
}

func helloword(){
	fmt.Println("hello world")
}