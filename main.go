package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
)

var logger *log.Logger

func main() {

	ctx := context.Background()

	// Set Google Cloud Platform project ID
	projectID := "checkup-339803"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Sets the name of the log to write to.
	logName := "checkup-logs"

	logger = client.Logger(logName).StandardLogger(logging.Info)

	// Logs "hello world", log entry is visible at
	// Cloud Logs.
	logger.Println("hello world")

	router := gin.New()
	router.Use(
		gin.Recovery(),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Status": "OK",
		})
	})
	router.GET("/select", selectHandler)
	router.POST("/insert", insertHandler)

	if err = SetupDatabase(); err != nil {
		logger.Printf("could not setup database: %s", err.Error())
		return
	}

	logger.Print("Listening on port " + port)

	if err := router.Run(":" + port); err != nil {
		logger.Fatal(err)
	}
}

func selectHandler(c *gin.Context) {

	logger.Print("selectHandler called")

	var data BaseUsersData
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := GetDataForUsername(data.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"Status":   "OK",
		"Username": data.Username,
		"Data":     jsonData,
	})
}

func insertHandler(c *gin.Context) {

	logger.Print("insertHandler called")

	var data BaseUsersData
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	rows, err := SetDataForUsername(data.Username, data.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"Status": "OK", "Rows": rows})
}
