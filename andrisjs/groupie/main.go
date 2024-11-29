package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Band struct {
	Name      string   `json:"name"`
	Image     string   `json:"image"`
	Year      int      `json:"year"`
	Album     string   `json:"album"`
	Members   []string `json:"members"`
	Dates     []string `json:"dates"`
	Locations []string `json:"locations"`
}

func main() {
	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API endpoint to fetch band data
	http.HandleFunc("/api/bands", func(w http.ResponseWriter, r *http.Request) {
		bands := loadSampleBands()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bands)
	})

	// Start the server
	println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Load sample band data from JSON files in the "data" directory
func loadSampleBands() []Band {
	var bands []Band
	dataDir := "./data"

	// Read the files in the "data" directory
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatalf("Failed to read data directory: %v", err)
	}

	// Loop through all files in the "data" directory
	for _, file := range files {
		// Only process JSON files
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(dataDir, file.Name())

			// Read the JSON file content
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Failed to read file %s: %v", filePath, err)
				continue // Skip the current file and move to the next
			}

			// Parse the JSON into a Band struct
			var band Band
			if err := json.Unmarshal(fileData, &band); err != nil {
				log.Printf("Failed to parse JSON from file %s: %v", filePath, err)
				continue // Skip the file if it cannot be parsed
			}

			// Append the parsed band to the bands slice
			bands = append(bands, band)
		}
	}

	// Return the slice of bands
	return bands
}
