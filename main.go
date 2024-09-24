package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Read(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Todo struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	rows, err := db.Query("SELECT * FROM t_todos")
	if err != nil {
		http.Error(w, "Error running the DB query", http.StatusInternalServerError)
	}

	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Name); err != nil {
			http.Error(w, "Error scanning the returned rows", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func Create(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	payload := struct {
		Name string `json:"name"`
	}{}

	type res struct {
		Res bool `json:"res"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		panic(err)
	}

	query := fmt.Sprintf("INSERT INTO t_todos (name) VALUES ('%s')", payload.Name)
	if _, err := db.Exec(query); err != nil {
		http.Error(w, "Error when executing query", http.StatusInternalServerError)
	}

	result := []res{
		{Res: true},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		log.Println("Error encoding JSON response:", err)
		return
	}

}

func Delete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type res struct {
		Res bool `json:"res"`
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		response := map[string]string{"error": err.Error()}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	query := fmt.Sprintf("DELETE FROM t_todos WHERE id = %d", id)
	if _, err := db.Exec(query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		response := map[string]string{"error": err.Error()}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	result := []res{
		{Res: true},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		response := map[string]string{"error": err.Error()}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}
}

func main() {

	mux := http.NewServeMux()

	connStr := "postgresql://<username>:<password>@<database_ip:port/database_name>?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connecting to database: ", err)
		log.Fatal(err)
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Database is not reachable", err)
	}

	wrappedMux := cors(mux)

	mux.HandleFunc("GET /todos/read", func(w http.ResponseWriter, r *http.Request) {
		Read(w, r, db)
	})
	mux.HandleFunc("POST /todos/create", func(w http.ResponseWriter, r *http.Request) {
		Create(w, r, db)
	})
	mux.HandleFunc("DELETE /todos/delete/{id}", func(w http.ResponseWriter, r *http.Request) {
		Delete(w, r, db)
	})

	fmt.Println("Server running at port :5000")

	if err := http.ListenAndServe(":5000", wrappedMux); err != nil {
		fmt.Println("Error starting server: ", err)
	}

}
