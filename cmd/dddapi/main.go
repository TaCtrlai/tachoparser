package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/traconiq/tachoparser/internal/pkg/certificates"
	"github.com/traconiq/tachoparser/pkg/decoder"
)

var (
	port = flag.String("port", "8080", "Port to listen on")
	addr = flag.String("addr", "", "Address to listen on (empty = all interfaces)")
)

func main() {
	flag.Parse()

	// Suppress certificate warnings in production (they're just informational)
	log.SetOutput(io.Discard)

	// Create Gin router
	r := gin.Default()

	// Add CORS middleware (useful for web clients)
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "ddd-parser-api",
		})
	})

	// Parse endpoint - accepts file upload or raw binary data
	r.POST("/parse", parseHandler)
	r.POST("/parse/card", parseCardHandler)
	r.POST("/parse/vu", parseVuHandler)

	// Root endpoint with API info
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "DDD Parser API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"POST /parse": "Auto-detect and parse Card or VU data",
				"POST /parse/card": "Parse as Card (TLV format)",
				"POST /parse/vu": "Parse as VU (TV format)",
				"GET /health": "Health check",
			},
		})
	})

	// Start server
	listenAddr := *addr + ":" + *port
	log.SetOutput(os.Stderr)
	log.Printf("DDD Parser API listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, r))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func parseHandler(c *gin.Context) {
	data, err := getDataFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try Card first
	var card decoder.Card
	verified, err := decoder.UnmarshalTLV(data, &card)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"type":     "card",
			"verified": verified,
			"data":     card,
		})
		return
	}

	// If Card parsing failed, try VU
	var vu decoder.Vu
	verified, err = decoder.UnmarshalTV(data, &vu)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse as either Card or VU",
			"card_error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "vu",
		"verified": verified,
		"data":     vu,
	})
}

func parseCardHandler(c *gin.Context) {
	data, err := getDataFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card decoder.Card
	verified, err := decoder.UnmarshalTLV(data, &card)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse as Card",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "card",
		"verified": verified,
		"data":     card,
	})
}

func parseVuHandler(c *gin.Context) {
	data, err := getDataFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var vu decoder.Vu
	verified, err := decoder.UnmarshalTV(data, &vu)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse as VU",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "vu",
		"verified": verified,
		"data":     vu,
	})
}

func getDataFromRequest(c *gin.Context) ([]byte, error) {
	// Try to get file from multipart form first
	file, _, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		return io.ReadAll(file)
	}

	// If no file, try reading from request body
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, gin.Error{Err: err, Meta: "No data provided. Send file as multipart/form-data with key 'file' or send raw binary in request body"}
	}

	return data, nil
}

