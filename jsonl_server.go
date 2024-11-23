package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "embed"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed schemas/OpenApi.schema.json
var openAPISpec embed.FS

// MyServer implements ServerInterface
type MyServer struct {
	Records      []Record
	TotalRecords int
}

// CountRecords returns the total count of records
func (s *MyServer) CountRecords(ctx echo.Context) error {
	return ctx.JSON(200, s.TotalRecords)
}

// ListRecords handles paginated listing of records
func (s *MyServer) ListRecords(ctx echo.Context, params ListRecordsParams) error {
	limit := 10 // Default limit
	offset := 0 // Default offset

	if params.Limit != nil {
		limit = *params.Limit
	}

	if params.Offset != nil {
		offset = *params.Offset
	}

	// Calculate bounds
	start := offset
	end := offset + limit
	if start >= s.TotalRecords {
		return ctx.JSON(200, []map[string]interface{}{})
	}
	if end > s.TotalRecords {
		end = s.TotalRecords
	}

	return ctx.JSON(200, s.Records[start:end])
}

// ShowRecordById retrieves a specific record by index
func (s *MyServer) ShowRecordById(ctx echo.Context, index int) error {
	if index < 0 || index >= s.TotalRecords {
		return ctx.JSON(404, nil)
	}
	return ctx.JSON(200, s.Records[index])
}

func RegisterSpecRoute(e *echo.Echo) {
	e.GET("/openapi.json", func(c echo.Context) error {
		spec, err := openAPISpec.ReadFile("schemas/OpenApi.schema.json")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load OpenAPI spec"})
		}
		return c.Blob(http.StatusOK, "application/json", spec)
	})
}

func ServeJsonl(config *Config) {
	// Load JSONL data
	jsonlSource := *config.tempThing.ServeJsonl.Source
	file, err := os.Open(jsonlSource)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	var records []map[string]interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			log.Printf("Error parsing JSON line: %v", err)
			continue
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	totalRecords := len(records)
	log.Printf("Loaded %d records", totalRecords)

	// Initialize Echo and server
	e := echo.New()
	server := &MyServer{
		Records:      records,
		TotalRecords: totalRecords,
	}

	// Register handlers using oapi-codegen
	RegisterHandlers(e, server)

	// Register OpenAPI spec route
	RegisterSpecRoute(e)

	// Start server
	port := 8080
	if config.tempThing.ServeJsonl.Port != 0 {
		port = config.tempThing.ServeJsonl.Port
	}
	// Add CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	log.Printf("Starting server on port %d", port)
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
