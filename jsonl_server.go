/*
package main

import (

	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"

)

	func ServeJsonl(config *Config) {
		jsonlSource := *config.tempThing.ServeJsonl.Source
		fmt.Println("serve jsonl", jsonlSource)
		// Read the JSONL file into memory
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

		http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			json.NewEncoder(w).Encode(map[string]int{"count": totalRecords})
		})

		http.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			// Check if index parameter is provided for single record retrieval
			if idx := r.URL.Query().Get("index"); idx != "" {
				index, err := strconv.Atoi(idx)
				if err != nil {
					http.Error(w, "Invalid index parameter", http.StatusBadRequest)
					return
				}

				// Return latest record if index is -1
				if index == -1 {
					index = totalRecords - 1
				}

				if index < 0 || index >= totalRecords {
					json.NewEncoder(w).Encode(nil)
					return
				}

				json.NewEncoder(w).Encode(records[index])
				return
			}

			// Handle limit/offset pagination
			limit := 10 // Default limit
			offset := 0 // Default offset

			if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
				if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
					limit = l
				}
			}

			if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
				if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
					offset = o
				}
			}

			// If no parameters provided, return last page
			if r.URL.Query().Get("limit") == "" && r.URL.Query().Get("offset") == "" {
				offset = totalRecords - limit
				if offset < 0 {
					offset = 0
				}
			}

			end := offset + limit
			if end > totalRecords {
				end = totalRecords
			}

			if offset >= totalRecords {
				json.NewEncoder(w).Encode([]map[string]interface{}{})
				return
			}

			json.NewEncoder(w).Encode(records[offset:end])
		})

		port := 8080
		if config.tempThing.ServeJsonl.Port != 0 {
			port = config.tempThing.ServeJsonl.Port
		}

		log.Printf("Starting server on port %d", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Fatalf("Server error: %v", err)
		}

}
*/
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
	return ctx.JSON(200, map[string]int{"count": s.TotalRecords})
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

	log.Printf("Starting server on port %d", port)
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
