package logging

import (
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
)

func HandleAWSError(err *http.ResponseError, service string, operation string) {
	logger.Warn(runtime.Caller(1))
	switch err.Response.StatusCode {
	case 403:
		logger.Warn("service", service, "operation", operation, "status", err.HTTPStatusCode(), "err", "permission denied")
	default:
		logger.Warn("service", service, "operation", operation, "status", err.HTTPStatusCode(), "error", err.ResponseError, "err", err.Unwrap())
	}
}

func HandleError(err error, service string, operation string, exitonError ...bool) {
	_, file, line, _ := runtime.Caller(1)
	logger.Warn("Error pointer: %s:%d\n", file, line)
	if len(exitonError) >= 1 && !exitonError[0] {
		logger.Warn("service", service, "operation", operation, "err", err)
	} else {
		logger.Error("service", service, "operation", operation, "err", err)
	}
}
