package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintln(w,"Hello, World!")
		w.Write([]byte("Hello"))
	})

	fmt.Println("Server listening on :9000")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
