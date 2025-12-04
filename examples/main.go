package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ashpect/revproxy/pkg/utils"
)

func main() {
	http.HandleFunc("/", basicResponse)
	http.HandleFunc("/stream", streamResponse)

	fmt.Println("Server listening on :9000")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func basicResponse(w http.ResponseWriter, r *http.Request) {
	utils.PrintRequest(r, "=== Basic Response Request ===")
	w.Write([]byte("Hello"))
}

func streamResponse(w http.ResponseWriter, r *http.Request) {
	utils.PrintRequest(r, "=== Stream Response Request ===")
	w.WriteHeader(http.StatusOK)

	for i := 0; i < 2; i++ {
		w.Write([]byte(fmt.Sprintf("Hello %d", i)))
		time.Sleep(1 * time.Second)
	}
}
