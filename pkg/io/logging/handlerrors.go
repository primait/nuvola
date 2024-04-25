package logging

import (
	"fmt"
	"log"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
)

func HandleAWSError(err *http.ResponseError, service string, operation string) {
	fmt.Println(runtime.Caller(1))
	switch err.Response.StatusCode {
	case 400:
		log.Fatalf("Service: %s, error: %v\n", service, err.Unwrap())
	case 403:
		log.Fatalf("Service: %s, Operation: %s, error: %s\n", service, operation, "Permission Denied")
	default:
		log.Fatalf("Service: %s, Operation: %s, StatusCode: %d, error: %v", service, operation, err.HTTPStatusCode(), err.ResponseError)
	}
}

func HandleError(err error, service string, operation string, exitonError ...bool) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("Error pointer: %s:%d\n", file, line)
	if len(exitonError) >= 1 && !exitonError[0] {
		log.Printf("Service: %s, Operation: %s, Error: %s\n", service, operation, err)
	} else {
		log.Fatalf("Service: %s, Operation: %s, Error: %s\n", service, operation, err)
	}
}
