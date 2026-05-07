package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Patient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	CPF  string `json:"cpf"`
}

var db *sql.DB

func main() {
	var err error

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("DB_PASSWORD não definida")
	}

	connStr := fmt.Sprintf(
		"host=postgres-homelab port=5432 user=admin password=%s dbname=hospital sslmode=disable",
		password,
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("erro ao abrir conexão:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	log.Println("✅ Conectado ao banco com sucesso")

	createTable()

	http.HandleFunc("/patients", patientsHandler)

	log.Println("🚀 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS patients (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		cpf TEXT NOT NULL UNIQUE
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("erro ao criar tabela:", err)
	}

	log.Println("📦 tabela patients pronta")
}

func patientsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {

	// ================= CREATE =================
	case http.MethodPost:
		var patient Patient

		err := json.NewDecoder(r.Body).Decode(&patient)
		if err != nil {
			log.Println("❌ erro ao decodificar JSON:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = db.QueryRow(
			"INSERT INTO patients(name, cpf) VALUES($1, $2) RETURNING id",
			patient.Name,
			patient.CPF,
		).Scan(&patient.ID)

		if err != nil {
			log.Println("❌ erro ao inserir paciente:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("✅ paciente criado: id=%d name=%s cpf=%s\n",
			patient.ID, patient.Name, patient.CPF)

		json.NewEncoder(w).Encode(patient)

	// ================= LIST =================
	case http.MethodGet:

		rows, err := db.Query("SELECT id, name, cpf FROM patients")
		if err != nil {
			log.Println("❌ erro no SELECT:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		patients := make([]Patient, 0)

		for rows.Next() {
			var patient Patient

			err := rows.Scan(&patient.ID, &patient.Name, &patient.CPF)
			if err != nil {
				log.Println("❌ erro no scan:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			patients = append(patients, patient)
		}

		if err := rows.Err(); err != nil {
			log.Println("❌ erro nas rows:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("📄 listagem de pacientes retornada (%d registros)\n", len(patients))

		json.NewEncoder(w).Encode(patients)

	default:
		log.Println("⚠️ método não permitido:", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
