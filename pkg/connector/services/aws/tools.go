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
	URL := "https://awspolicygen.s3.amazonaws.com/js/policies.js"
	client := req.C().SetBaseURL(URL).SetTimeout(30 * time.Second).SetUserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:103.0) Gecko/20100101 Firefox/103.0")

	response := client.Get().
		SetHeader("Connection", "keep-alive").
		SetHeader("Pragma", "no-cache").
		SetHeader("Cache-Control", "no-cache").
		Do()
	if response.Err != nil {
		logging.HandleError(response.Err, "AWS - SetActions", "Error on calling HTTP endpoint")
	}

	resString := strings.Replace(response.String(), "app.PolicyEditorConfig=", "", 1)
	obj, err := oj.ParseString(resString)
	if err != nil {
		logging.HandleError(err, "AWS - SetActions", "Error on parsing output string")
	}
	query, err := gojq.Parse(`.serviceMap[] | .StringPrefix as $prefix | .Actions[] | "\($prefix):\(.)"`)
	if err != nil {
		logging.HandleError(err, "AWS - SetActions", "Error on mapping string to object")
	}

	iter := query.Run(obj)
	ActionsMap = make(map[string][]string, 0)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			logging.HandleError(err, "AWS - SetActions", "Error on itering over objects")
		}

		ActionsList = append(ActionsList, v.(string))
		split := strings.Split(v.(string), ":")
		ActionsMap[split[0]] = append(ActionsMap[split[0]], split[1])
	}

	ActionsList = unique(ActionsList)
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
