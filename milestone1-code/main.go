package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	FILEPATH     = "DATA_FILE_PATH"
	mu           sync.Mutex
	mappedSecret = make(map[string]string)
)

type SecretRequest struct {
	PlainText string `json:"plain_text"`
}

type SecretPostResponse struct {
	Id string `json:"id"`
}

type SecretGetResponse struct {
	Data string `json:"data"`
}

func main() {
	filePath, ok := os.LookupEnv(FILEPATH)
	fmt.Println(os.Getenv(FILEPATH))
	if !ok {
		log.Fatalf("No environment variable %s specified", FILEPATH)
	}

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		os.Create(filePath)
	}

	//Load persisted data into map
	readFileIntoMap()

	http.HandleFunc("/healthcheck", healthCheckHandler)
	http.HandleFunc("/", secretHandler)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalf("Error: %s", err)
		return
	}

}

func secretHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		secretGetHandler(writer, request)
	case "POST":
		secretPostHandler(writer, request)
	}
}

func readFileIntoMap() {
	filePath, ok := os.LookupEnv(FILEPATH)
	if ok {
		fd, err := os.Open(filePath)
		if err != nil {
			return
		}
		defer fd.Close()
		reader := bufio.NewReader(fd)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
			datas := strings.Split(line, "|")
			mappedSecret[datas[0]] = strings.TrimSpace(datas[1])
			log.Println(mappedSecret)
		}
	}
}

func secretGetHandler(writer http.ResponseWriter, request *http.Request) {
	pathsRaw := request.RequestURI
	paths := strings.Split(strings.TrimSpace(pathsRaw), "/")
	response := &SecretGetResponse{
		Data: "",
	}
	if len(paths) == 2 {
		hashedId := paths[1]
		if val, ok := mappedSecret[hashedId]; ok {
			response.Data = val
			body, _ := json.Marshal(response)
			fmt.Fprintf(writer, string(body))
			return
		}
		body, _ := json.Marshal(response)
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, string(body))
	} else {
		body, _ := json.Marshal(response)
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, string(body))
	}
}

func secretPostHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	var t SecretRequest
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Fatalf("error: ", err)
		return
	}

	toBeHashed := t.PlainText
	hash := md5.New()
	_, err = io.WriteString(hash, toBeHashed)
	if err != nil {
		return
	}
	hashed := hex.EncodeToString(hash.Sum(nil))
	mappedSecret[hashed] = toBeHashed

	writeToFile(mappedSecret)

	response := &SecretPostResponse{
		Id: hashed,
	}
	body, err = json.Marshal(response)
	_, err = fmt.Fprintf(writer, "%s", string(body))
	if err != nil {
		log.Fatalf("error: ", err)
		return
	}
}

func writeToFile(secrets map[string]string) {
	filePath, ok := os.LookupEnv(FILEPATH)
	if ok {
		mu.Lock()
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0660)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		for k, v := range secrets {
			fmt.Fprintf(file, "%s|%s\n", k, v)
		}
		mu.Unlock()
	}
	return
}

func healthCheckHandler(writer http.ResponseWriter, request *http.Request) {
	_, err := fmt.Fprintf(writer, "ok")
	if err != nil {
		return
	}
}
