package awsconnector

import (
	"strings"
	"time"

	req "github.com/imroc/req/v3"
	"github.com/itchyny/gojq"
	"github.com/ohler55/ojg/oj"
	"github.com/primait/nuvola/pkg/io/logging"
)

func SetActions() {
	logger := logging.GetLogManager()
	URL := "https://awspolicygen.s3.amazonaws.com/js/policies.js"
	client := req.C().SetBaseURL(URL).SetTimeout(30 * time.Second).SetUserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:103.0) Gecko/20100101 Firefox/103.0")

	response, err := client.R().
		SetHeader("Connection", "keep-alive").
		SetHeader("Pragma", "no-cache").
		SetHeader("Cache-Control", "no-cache").
		Get(URL)
	if err != nil {
		logger.Error("Error on calling HTTP endpoint", "err", err)
	}

	resString := strings.Replace(response.String(), "app.PolicyEditorConfig=", "", 1)
	obj, err := oj.ParseString(resString)
	if err != nil {
		logger.Error("Error on parsing output string", "err", err)
	}
	query, err := gojq.Parse(`.serviceMap[] | .StringPrefix as $prefix | .Actions[] | "\($prefix):\(.)"`)
	if err != nil {
		logger.Error("Error on mapping string to object", "err", err)
	}

	iter := query.Run(obj)
	ActionsMap = make(map[string][]string)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			logger.Error("Error on iterating over objects", "err", err)
		}

		ActionsList = append(ActionsList, v.(string))
		split := strings.Split(v.(string), ":")
		ActionsMap[split[0]] = append(ActionsMap[split[0]], split[1])
	}

	ActionsList = unique(ActionsList)
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if !keys[entry] {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
