package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func SSRF(url string) (bodyString string) {
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	var client = &http.Client{
		Timeout:   time.Duration(5 * time.Second),
		Transport: t,
	}

	resp, err := client.Get(url)
	if err != nil {
		bodyString = err.Error() + "\n"
		return bodyString
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			bodyString = err.Error() + "\n"
			return bodyString
		}
		bodyString = string(bodyBytes)
	}
	return
}

func LFI(path string) (fileContent string) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
	}
	fileContent = string(b)

	return
}

func Handler(req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	ipAllowed, err := os.LookupEnv("IP_ALLOWED")
	if err && ipAllowed == req.Headers["x-forwarded-for"] {
		url := req.QueryStringParameters["url"]
		path := req.QueryStringParameters["path"]

		headers := map[string]string{
			"Content-Type": "text/plain",
		}

		if url == "" && path == "" {
			return events.LambdaFunctionURLResponse{
				StatusCode: 200,
				Headers:    headers,
				Body:       string("Hello!\n"),
			}, nil
		} else {
			if url != "" {
				return events.LambdaFunctionURLResponse{
					StatusCode: 200,
					Headers:    headers,
					Body:       SSRF(url),
				}, nil
			}
			if path != "" {
				return events.LambdaFunctionURLResponse{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "appliation/octet-stream",
					},
					Body: LFI(path),
				}, nil
			}
		}
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: 500,
		Body:       "Nope!\n",
	}, nil
}

func main() {
	lambda.Start(Handler)
}
