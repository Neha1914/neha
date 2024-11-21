package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int    `json:"year"`
}

var (
	movies = make(map[int]Movie)
	nextID = 1
	mu     sync.Mutex
)

func main() {
	http.HandleFunc("/movies", moviesHandler)
	http.HandleFunc("/movies/", movieHandler) // For specific movie actions
	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMovies(w)
	case http.MethodPost:
		createMovie(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func movieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getMovie(w, id)
	case http.MethodPut:
		updateMovie(w, r, id)
	case http.MethodDelete:
		deleteMovie(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getMovies(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()

	movieList := make([]Movie, 0, len(movies))
	for _, movie := range movies {
		movieList = append(movieList, movie)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movieList)
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	movie.ID = nextID
	nextID++
	movies[movie.ID] = movie
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func getMovie(w http.ResponseWriter, id int) {
	mu.Lock()
	defer mu.Unlock()

	movie, found := movies[id]
	if !found {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func updateMovie(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	defer mu.Unlock()

	movie, found := movies[id]
	if !found {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	movies[id] = movie
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func deleteMovie(w http.ResponseWriter, id int) {
	mu.Lock()
	defer mu.Unlock()

	if _, found := movies[id]; !found {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	delete(movies, id)
	w.WriteHeader(http.StatusNoContent)
}

func parseID(path string) (int, error) {
	var id int
	_, err := fmt.Sscanf(path, "/movies/%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid movie ID")
	}
	return id, nil
}
