package main

import (
	"log"
	"net/http"

	"github.com/PornchaiSakulsrimontri/gofinal/task"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Handler function
func setupRouter() *gin.Engine {
	r := gin.Default()

	//add middleware authen via Authorize header
	r.Use(authMiddleware)

	//custmer inject DB-> Handler via Struct
	r.GET("/customers", task.GetCustomersHandler)
	r.GET("/customers/:id", task.GetCustomerByIdHandler)

	r.POST("/customers", task.CreateCustomersHandler)

	r.PUT("/customers/:id", task.UpdateCustomerByIdHandler)
	r.DELETE("/customers/:id", task.DeleteCustomerHandler)

	return r
}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "November 10, 2009" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you don't have the permission!!"})
		c.Abort()
		return
	}
	c.Next()
}

func main() {
	//create table if not exist
	task.InitialCustomers()

	r := setupRouter()
	r.Run(":2009")

	log.Println("customer service")
}
