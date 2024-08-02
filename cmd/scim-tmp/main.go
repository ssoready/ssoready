package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/Users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)
		fmt.Println(r.URL)

		if err := json.NewEncoder(w).Encode(scimListUsersResponse{}); err != nil {
			panic(err)
		}
	}).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("fallback")
		fmt.Println(r.Header)
		fmt.Println(r.URL)
	})

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}
}

type scimListUsersResponse struct {
	Schemas      []string `json:"schemas"`
	TotalResults int      `json:"totalResults"`
	Resources    []struct {
		Id       string `json:"id"`
		UserName string `json:"userName"`
	} `json:"Resources"`
}
