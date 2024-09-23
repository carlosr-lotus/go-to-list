package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func Read(w http.ResponseWriter, r *http.Request) {
	type todo struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	response := []todo{
		{ID: 1, Name: "Listen to Lady Gaga!"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "New todo created!")
}

func Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintln(w, "Todo deleted: ", id)
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /todos/read", Read)
	mux.HandleFunc("POST /todos/create", Create)
	mux.HandleFunc("DELETE /todos/delete/{id}", Delete)

	wrappedMux := cors(mux)

	fmt.Println("Server running at port :5000")
	err := http.ListenAndServe(":5000", wrappedMux)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
