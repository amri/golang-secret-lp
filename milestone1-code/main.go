package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	envVar := "DATA_FILE_PATH"
	filePath, ok := os.LookupEnv(envVar)
	fmt.Println(os.Getenv(envVar))
	if !ok {
		log.Fatalf("No environment variable %s specified", envVar)
	}

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		os.Create(filePath)
	}

	http.HandleFunc("/healthcheck", healthCheckHandler)
	http.HandleFunc("/", secretHandler)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalf("Error: %s", err)
		return
	}

}

type SecretStruct struct {
	PlainText string `json:"plain_text"`
}

func secretHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		_, err := fmt.Fprintf(writer, "ok")
		if err != nil {
			return
		}
	case "POST":
		//dec := json.NewDecoder(request.Body)
		body, err := ioutil.ReadAll(request.Body)
		log.Println(body)
		if err != nil {
			return
		}
		var t SecretStruct
		err = json.Unmarshal(body, &t)
		log.Println(t)
		if err != nil {
			log.Fatalf("error: ", err)
			return
		}
		_, err = fmt.Fprintf(writer, t.PlainText)
		if err != nil {
			log.Fatalf("error: ", err)
			return
		}
	}

}

func healthCheckHandler(writer http.ResponseWriter, request *http.Request) {
	_, err := fmt.Fprintf(writer, "ok")
	if err != nil {
		return
	}
}
