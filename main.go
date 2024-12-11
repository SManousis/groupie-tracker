package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"text/template"
)

type Artist struct {
	ID           int               `json:"id"`
	Image        string            `json:"image"`
	Name         string            `json:"name"`
	Members      []string          `json:"members"`
	CreationDate int               `json:"creationDate"`
	FirstAlbum   string            `json:"firstAlbum"`
	Locations    []string          `json:"-"`
	Dates        []string          `json:"-"`
	Relation     map[string]string `json:"-"`
}

const baseUrl = "https://groupietrackers.herokuapp.com/api/"

var endpoints = []string{"locations/", "dates/", "relation/"}

// create a var FetchArtist so that it points to the actual implementation when used in unit testing
var FetchArtists = fetchArtists

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
		if locations, exists := rawData["locations"].([]interface{}); exists {
			artist.Locations = make([]string, len(locations))
			for i, v := range locations {
				artist.Locations[i] = fmt.Sprint(v)
			}
		}
	case "dates/":
		if dates, exists := rawData["dates"].([]interface{}); exists {
			artist.Dates = make([]string, len(dates))
			for i, v := range dates {
				artist.Dates[i] = fmt.Sprint(v)
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

func serveHome(w http.ResponseWriter, r *http.Request) {
	// Fetch the artists and dynamically update them
	artists := []Artist{}
	err := FetchArtists(&artists)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching artists: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the data to be passed to the template
	data := struct {
		Artists []Artist
	}{
		Artists: artists,
	}
	if r.URL.Path != "/" {
		http.ServeFile(w, r, "templates/error.html")
		return
	}
	// Prepare the template with embedded JavaScript
	tmpl, err := template.New("home").Parse(`
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Groupie Tracker</title>
        <link rel="stylesheet" href="/static/styles.css">
    </head>
    <body>
        <header>
  			<img src="/static/assets/header.png" alt="Groupie Tracker" />
		</header>
        <main id="band-container">
            <!-- Cards will be dynamically injected here -->
        </main>
		<script>
			window.artistData = [
			{{range $index, $artist := .Artists}}{{if $index}}, {{end}}{
				name: "{{.Name}}",
				image: "{{.Image}}",
				firstAlbum: "{{.FirstAlbum}}",
				creationDate: {{.CreationDate}},
				members: [{{range $i, $member := .Members}}{{if $i}}, {{end}}{{printf "%q" $member}}{{end}}],
				dates: [{{range $i, $date := .Dates}}{{if $i}}, {{end}}{{printf "%q" $date}}{{end}}],
				locations: [{{range $i, $location := .Locations}}{{if $i}}, {{end}}{{printf "%q" $location}}{{end}}],
				events: {
					{{- $first := true -}}
					{{- range $key, $value := .Relation -}}
						{{- if not $first}}, {{end -}}
						{{- $first = false -}}
						{{printf "%q: %q" $key $value}}
					{{- end -}}
				}
			}{{end}}
			];
		</script>
        <script src="/static/app.js" defer></script>
        <footer>
            <p>Groupie Tracker - All Rights Reserved &copy; 2024</p>
        </footer>
    </body>
    </html>
    `)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	// Execute the template with artist data
	//err = tmpl.Execute(w, data)
	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve the homepage on /
	http.HandleFunc("/", serveHome)

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
