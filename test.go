package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb"
)

type Service struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	ID       int    `json:"id"`
	Category string `json:"category"`
}

func main() {
	db, err := sql.Open("duckdb", "")
	handleError(err, "")
	defer db.Close()

	//
	rows, err := db.Query(`SELECT id, name, address, category FROM 'homepage.parquet';`)
	handleError(err, "")
	defer rows.Close()

	var id int
	var name, address, category string
	var services []Service
	for rows.Next() {
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
	handleError(err, "no")

	for _, service := range services {
		fmt.Printf("id: %d, name: %s\n", service.ID, service.Title)
	}

}

func fetchservices() []Service {
	db, err := sql.Open("duckdb", "")
	handleError(err, "")
	defer db.Close()

	//
	rows, err := db.Query(`SELECT id, name, address, category FROM 'homepage.parquet';`)
	handleError(err, "")
	defer rows.Close()

	var services []Service
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
	handleError(err, "")

	return services

}

func handleError(err error, str string) {
	if err != nil {
		log.Fatalf(str+"%v", err)
	}
}
