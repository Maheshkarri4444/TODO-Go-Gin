package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	router.POST("/signup")
	router.POST("/login")
	router.GET("/todo")

	router.Run(":4000")
}
