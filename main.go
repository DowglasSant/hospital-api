package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Patient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	CPF  string `json:"cpf"`
}

var db *sql.DB

func main() {
	var err error

	db, err = sql.Open("sqlite3", "./hospital.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable()

	http.HandleFunc("/patients", patientsHandler)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS patients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		cpf TEXT NOT NULL UNIQUE
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func patientsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:
		var patient Patient

		err := json.NewDecoder(r.Body).Decode(&patient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec(
			"INSERT INTO patients(name, cpf) VALUES(?, ?)",
			patient.Name,
			patient.CPF,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		patient.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patient)

	case http.MethodGet:

		rows, err := db.Query("SELECT id, name, cpf FROM patients")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var patients []Patient

		for rows.Next() {
			var patient Patient

			rows.Scan(&patient.ID, &patient.Name, &patient.CPF)

			patients = append(patients, patient)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patients)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
