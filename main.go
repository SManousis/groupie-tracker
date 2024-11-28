package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Artist struct {
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	Members      []string          `json:"members"`
	CreationDate int               `json:"creationDate"`
	FirstAlbum   string            `json:"firstAlbum"`
	Locations    []string          `json:"-"`
	Dates        []string          `json:"-"`
	Relation     map[string]string `json:"-"`
}

const baseUrl = "https://groupietrackers.herokuapp.com/api/"

var (
	endpoints    = []string{"locations/", "dates/", "relation/"}
	artistsCache []Artist // Added so that the handler can get the complete artist list
)

func fetchArtists(artists *[]Artist) error {
	response, err := http.Get(baseUrl + "artists")
	if err != nil {
		return err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(artists)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for counter := range *artists {
		for _, endpoint := range endpoints {
			wg.Add(1)
			go func(artist *Artist, endpoint string) {
				defer wg.Done()
				if err := updateArtistStruct(artist, endpoint); err != nil {
					fmt.Println("Error updating artist:", err)
				}
			}(&(*artists)[counter], endpoint)
		}
	}
	wg.Wait()
	return nil
}
func updateArtistStruct(artist *Artist, endpoint string) error {
	response, err := http.Get(baseUrl + endpoint + strconv.Itoa(artist.ID))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var rawData map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&rawData); err != nil {
		return err
	}
	switch endpoint {
	case "locations/":
		if dates, exists := rawData["dates"].([]interface{}); exists {
			artist.Dates = make([]string, len(dates))
			for i, v := range dates {
				artist.Dates[i] = fmt.Sprint(v)
			}
		}
	case "dates/":
		if locations, exists := rawData["locations"].([]interface{}); exists {
			artist.Locations = make([]string, len(locations))
			for i, v := range locations {
				artist.Locations[i] = fmt.Sprint(v)
			}
		}
	case "relation/":
		if datesLocations, exists := rawData["datesLocations"].(map[string]interface{}); exists {
			artist.Relation = make(map[string]string)
			for k, v := range datesLocations {
				artist.Relation[k] = fmt.Sprint(v)
			}
		}
	}
	return nil
}

// search handler to create a json output
func searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(artistsCache); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	start := time.Now()
	artists := []Artist{}
	err := fetchArtists(&artists)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	for _, artist := range artists {
		fmt.Println(artist)
	}
	artistsCache = artists
	elapsed := time.Since(start)
	fmt.Printf("Execution time: %s\n", elapsed)
	http.HandleFunc("/api/search", searchHandler) // Route for the API of the JS
	http.ListenAndServe("localhost:8080", nil)    // initialize server
}
