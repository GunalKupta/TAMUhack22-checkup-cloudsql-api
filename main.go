package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.Use(
		gin.Recovery(),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.GET("/", func(c *gin.Context) {
		fmt.Println("Health check")
		c.JSON(200, gin.H{
			"Status": "OK",
		})
	})
	router.GET("/select", selectHandler)
	router.POST("/insert", insertHandler)

	var err error
	if err = SetupDatabase(); err != nil {
		fmt.Println("Could not setup database: " + err.Error())
	}

	fmt.Println("Listening on port " + port)

	if err = router.Run(":" + port); err != nil {
		panic(err)
	}
}

// selectHandler handles the /select endpoint by querying
// the db for the given username
func selectHandler(c *gin.Context) {

	username, ok := c.GetQuery("username")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username not provided"})
		return
	}

	fmt.Println("selectHandler called username: " + username)

	data, err := GetDataForUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"Status":   "OK",
		"Username": username,
		"Data":     data,
	})
}

// insertHandler handles the /insert endpoint by inserting
// the given username and data into the db
func insertHandler(c *gin.Context) {

	fmt.Println("insertHandler called")

	var data BaseUsersData
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	rows, err := SetDataForUsername(data.Username, data.Data)
	if err != nil {
		fmt.Printf("setdata error: %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	fmt.Printf("inserted data: %#v\n", data)

	c.JSON(200, gin.H{"Status": "OK", "Rows": rows})
}
