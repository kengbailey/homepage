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
	http.HandleFunc("/editService", editService)

	log.Println("Listening on :1111 ...")
	err := http.ListenAndServe(":1111", nil)
	handleError(err, "")
}

// editService ...
func editService(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("duckdb", "duck.db")
	handleError(err, "")
	defer db.Close()

	sqlStatement := `UPDATE services SET name = $1, address = $2, category = $3 WHERE id = $4`
	_, err = db.Exec(sqlStatement, r.FormValue("title"), r.FormValue("url"), r.FormValue("category"), r.FormValue("id"))
	handleError(err, "Failed to edit service")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully edited service: %s", r.FormValue("id"))
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

	var newID int64 = rand.Int63n(100)

	// Check if the generated ID already exists in the database
	for {
		idExists := checkIfIDExists(db, newID)
		if !idExists {
			break
		}
		newID = rand.Int63n(100) // If it does, generate a new ID and try again
	}
	sqlStatement := `INSERT INTO services (id, name, address, category) VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(sqlStatement, newID, r.FormValue("title"), r.FormValue("url"), r.FormValue("category"))
	handleError(err, "")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully inserted service: %v", newID)
}

// checkIfIDExists ...
func checkIfIDExists(db *sql.DB, id int64) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)", id).Scan(&exists)
	handleError(err, "")
	return exists
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
