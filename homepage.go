package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	_ "github.com/marcboeker/go-duckdb"
)

func main() {

	http.Handle("/", http.FileServer(http.Dir("./")))

	http.HandleFunc("/getServices", fetchServices)
	http.HandleFunc("/createService", createService)
	http.HandleFunc("/deleteService", deleteService)
	http.HandleFunc("/getBerryServices", fetchBerryServices)

	log.Println("Listening on :80...")
	err := http.ListenAndServe(":80", nil)
	handleError(err, "")
}

// handleError ...
func handleError(err error, str string) {
	if err != nil {
		log.Fatalf(str+"%v", err)
	}
}

type Service struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	ID       int    `json:"id"`
	Category string `json:"category"`
}

// fetchBerryServices ...
func fetchBerryServices(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("duckdb", "duck.db")
	handleError(err, "")
	defer db.Close()

	rows, err := db.Query(`SELECT id, name, address FROM services where category = 'Berry'`)
	handleError(err, "")
	defer rows.Close()

	services := make([]Service, 0)
	for rows.Next() {
		var id int
		var name, address string
		err = rows.Scan(&id, &name, &address)
		handleError(err, "")

		services = append(services, Service{
			ID:    id,
			Title: name,
			Url:   address,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(services)
	handleError(err, "")

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// fetchServices ...
func fetchServices(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("duckdb", "duck.db")
	handleError(err, "")
	defer db.Close()

	rows, err := db.Query(`SELECT id, name, address, category FROM services`)
	handleError(err, "")
	defer rows.Close()

	services := make([]Service, 0)
	for rows.Next() {
		var id int
		var name, address, category string
		err = rows.Scan(&id, &name, &address, &category)
		handleError(err, "")

		services = append(services, Service{
			ID:       id,
			Title:    name,
			Url:      address,
			Category: category,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(services)
	handleError(err, "")

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// createService ...
func createService(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("duckdb", "duck.db")
	handleError(err, "")
	defer db.Close()

	var id int64 = rand.Int63n(100)
	sqlStatement := `INSERT INTO services (id, name, address, category) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(sqlStatement, id, r.FormValue("title"), r.FormValue("url"), r.FormValue("category"))
	handleError(err, "")

	// fmt.Fprintf(w, "Successfully inserted service: %s", r.FormValue("title"))
	http.Redirect(w, r, "/index.html", http.StatusCreated)
}

// deleteService ...
func deleteService(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("duckdb", "duck.db")
	handleError(err, "")
	defer db.Close()

	sqlStatement := `DELETE FROM services WHERE id = $1`
	_, err = db.Exec(sqlStatement, r.FormValue("id"))
	handleError(err, "Failed to delete service")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully deleted service: %s", r.FormValue("id"))
}
