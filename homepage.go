package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	_ "github.com/marcboeker/go-duckdb"
)

// next steps
// 0. DONE - fix deprecated code above
// 1. DONE - create api for single function ... get services ... we'll deploy to all servers
// 2. create call to "hosted" get services api endpoints on local network
// 3. put services retrieved from servers into html on website
// 4. update Tim website template to have new format for hosting services!!
// 5. add ping endpoint to server apis ... (1) Uptime (2) CPU utilization (3) disk space remaining (4) containers running
// 6. pull in variables from yaml file ... server file locations

// AJAX --> https://www.horilla.com/blogs/how-to-set-up-ajax-ultimate-guide/

func main() {

	http.Handle("/", http.FileServer(http.Dir("./")))

	// http.HandleFunc("/getServices", servicesHandler)
	// http.HandleFunc("/cpuLoad", cpuLoadHandler)
	// http.HandleFunc("/systemUptime", systemUptimeHandler)

	log.Println("Listening on :3859...")
	err := http.ListenAndServe(":3859", nil)
	handleError(err, "")
}

type Service struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	ID       int    `json:"id"`
	Category string `json:"category"`
}

func functionName() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// _, err = db.Exec(`INSERT INTO people VALUES (42, 'John')`)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// var id int
	// var name string
	rows, err := db.Query(`SELECT id, name, address, category FROM services`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var services []struct {
		ID       int
		Name     string
		Address  string
		Category string
	}

	for rows.Next() {
		var id int
		var name string
		var address string
		var category string
		err = rows.Scan(&id, &name, &address, &category)
		if err != nil {
			log.Fatal(err)
		}

		services = append(services, struct {
			ID       int
			Name     string
			Address  string
			Category string
		}{
			ID:       id,
			Name:     name,
			Address:  address,
			Category: category,
		})
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range services {
		fmt.Printf("id: %d, name: %s\n", service.ID, service.Name)
	}

}

// insertService ...
func insertService(w http.ResponseWriter, r *http.Request) {
	// open sqlite databse
	db, err := sql.Open("sqlite3", "./data/services.db")
	handleError(err, "")
	defer db.Close()
	// insert service into database
	stmt, err := db.Prepare("INSERT INTO services (url, title) VALUES (?, ?)")
	handleError(err, "")
	_, err = stmt.Exec(r.FormValue("url"), r.FormValue("title"))
	handleError(err, "")
	// return success message to client
	fmt.Fprintf(w, "Successfully inserted service: %s", r.FormValue("title"))
}

func deleteService(w http.ResponseWriter, r *http.Request) {
	// open sqlite database
	db, err := sql.Open("sqlite3", "./data/services.db")
	handleError(err, "")
	defer db.Close()
	// parse the request body to extract the service ID
	id := r.FormValue("id")
	// delete the service from the database
	stmt, err := db.Prepare("DELETE FROM services WHERE id=?")
	handleError(err, "")
	_, err = stmt.Exec(id)
	handleError(err, "")
	// return success message to client
	fmt.Fprintf(w, "Successfully deleted service: %s", id)
}

// golang function to read url+title from a table in a sqlite database
func getServices(w http.ResponseWriter, r *http.Request) {
	// open sqlite databse
	db, err := sql.Open("sqlite3", "./data/services.db")
	handleError(err, "")
	defer db.Close()

	// read url+title from table
	rows, err := db.Query("SELECT url, title FROM services")
	handleError(err, "")
	defer rows.Close()

	// parse rows into json and return as response
	var services []Service
	for rows.Next() {
		var url string
		var title string
		err = rows.Scan(&url, &title)
		handleError(err, "")
		services = append(services, Service{url, title})
	}

	jsonData, err := json.Marshal(services)
	handleError(err, "")

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// cpuLoadHandler
func cpuLoadHandler(w http.ResponseWriter, r *http.Request) {
	cpuLoad, err := getCpuLoad()
	if err != nil {
		http.Error(w, "Failed to fetch CPU load", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(cpuLoad)
	if err != nil {
		http.Error(w, "Failed to marshal JSON data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// systemUptimeHandler
func systemUptimeHandler(w http.ResponseWriter, r *http.Request) {
	uptime, err := getSystemUptime()
	if err != nil {
		http.Error(w, "Failed to fetch system uptime", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(uptime)
	if err != nil {
		http.Error(w, "Failed to marshal JSON data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// servicesHandler
func servicesHandler(w http.ResponseWriter, r *http.Request) {

	services := get_docker_services()
	jupyterLabURL := get_pine_jupyterlab_url()
	for i, service := range services {
		if strings.Contains(service[0], "jupyter") {
			services[i] = append(services[i], jupyterLabURL)
		}
	}

	jsonData, err := json.Marshal(services)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func get_docker_services() [][]string {

	// list docker containers
	cli, err := client.NewClientWithOpts(client.FromEnv)
	handleError(err, "Error creating Docker client")
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	handleError(err, "")

	// get local ip
	conn, err := net.Dial("udp", "8.8.8.8:80")
	handleError(err, "")
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	conn.Close()

	var services [][]string
	for _, container := range containers {
		serviceURL := fmt.Sprintf("http://%v:%v\n", localAddr.IP, container.Ports[0].PublicPort)
		newService := []string{container.Image, serviceURL}
		services = append(services, newService)
	}

	return services
}

func get_pine_jupyterlab_url() string {
	data, err := os.ReadFile("/home/syran/sandbox/docker/jupyter_url.txt")
	if err != nil {
		return ""
	}

	return string(data)
}

func handleError(err error, str string) {
	if err != nil {
		log.Fatalf(str+"%v", err)
	}
}

type CpuLoad struct {
	Load float64 `json:"load"`
}

type SystemUptime struct {
	Uptime string `json:"uptime"`
}

func getSystemUptime() (*SystemUptime, error) {
	output, err := exec.Command("uptime").Output()
	if err != nil {
		return nil, err
	}

	uptimeString := strings.Split(string(output[:len(output)-1]), " ")[4]
	uptime := strings.Trim(uptimeString, ",")

	return &SystemUptime{Uptime: uptime}, nil

}

func getCpuLoad() (*CpuLoad, error) {
	output, err := exec.Command("cat", "/proc/loadavg").Output()
	if err != nil {
		return nil, err
	}

	fields := strings.Split(string(output), " ")
	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return nil, err
	}

	return &CpuLoad{Load: load}, nil
}
