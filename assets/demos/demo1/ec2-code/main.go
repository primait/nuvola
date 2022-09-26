package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello!\n")

	if len(r.URL.Query()["url"]) > 0 {
		io.WriteString(w, SSRF(r.URL.Query()["url"][0]))
	}
}

func SSRF(url string) (bodyString string) {
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	// localhost:3333?url=http://169.254.169.254/latest/meta-data/identity-credentials/ec2/security-credentials/ec2-instance
	// localhost:3333?url=file:///proc/self/environ
	var client = &http.Client{
		Timeout:   time.Duration(5 * time.Second),
		Transport: t,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		bodyString = string(bodyBytes)
	}
	return
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	var port string = "3333"
	var host string = "0.0.0.0"
	fmt.Printf("Started on %s:%s\n", host, port)

	err := http.ListenAndServe(host+":"+port, mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
