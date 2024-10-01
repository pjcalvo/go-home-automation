package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

type cmdresult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func homepage(writer http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(writer, "Go Home Simple Rest Api Server")
}
func getdate(writer http.ResponseWriter, _ *http.Request) {
	var result cmdresult

	out, err := exec.Command("date").Output()
	if err == nil {
		result.Success = true
		result.Message = fmt.Sprintf("The date is %s", string(out))
	} else {
		result.Message = fmt.Sprintf("Errored out %v", err)
	}

	json.NewEncoder(writer).Encode(result)
}

func main() {
	http.HandleFunc("/", homepage)
	http.HandleFunc("/api/v1/getdate", getdate)
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
