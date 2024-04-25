package yamler

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/primait/nuvola/tools/filesystem/files"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type Conf struct {
	Description string                   `yaml:"description,omitempty"`
	Name        string                   `yaml:"name"`
	Services    []string                 `yaml:"services,omitempty"`
	Properties  []map[string]interface{} `yaml:"properties,omitempty"`
	Return      []string                 `yaml:"return"`
	Enabled     bool                     `yaml:"enabled"`
	Find        Find                     `yaml:"find,omitempty"`
}

type Find struct {
	Who    []string
	To     []string
	With   []string            `yaml:"with,omitempty"`
	Target []map[string]string `yaml:"target,omitempty"`
}

func GetConf(file string) (c *Conf) {
	c = &Conf{}
	yamlFile, err := os.ReadFile(files.NormalizePath(file))
	if err != nil {
		logging.HandleError(err, "Yamler - GetConf", "Error on reading rule file")
	}
	c.Enabled = true // Default value is: Enabled
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logging.HandleError(err, "Yamler - GetConf", "Umarshalling yamlFile")
	}

	return c
}

func PrepareQuery(config *Conf) (query string, arguments map[string]interface{}) {
	arguments = make(map[string]interface{}, 0)
	if config.Services != nil {
		// Direct access to properties
		query = prepareService(config.Services, arguments) +
			prepareProperties(config.Properties, arguments) +
			prepareResults(config.Return, arguments)
	} else if config.Find.To != nil || config.Find.Who != nil || config.Find.With != nil {
		if config.Find.Target != nil {
			query = prepareQueryPrivEsc(config, arguments)
		} else {
			query = preparePathQuery(config, arguments)
		}
	} else {
		logging.HandleError(nil, "Yamler - PrepareQuery", "Malformed rule!")
	}
	return query, arguments
}

func preparePathQuery(rule *Conf, arguments map[string]interface{}) (query string) {
	template := "MATCH m%d = (who)-[:HAS_POLICY]->(:Policy)-[:ALLOWS]->(:Action {Service: $service%d, Action: $action%d}) \n"

	matchQueries := ""
	whereFilters := ""
	returnValues := ""

	for i, perm := range rule.Find.With {
		withSplit := strings.Split(perm, ":")
		service := withSplit[0]
		action := withSplit[1]
		arguments["action"+strconv.Itoa(i)] = action
		arguments["service"+strconv.Itoa(i)] = service

		matchQueries += fmt.Sprintf(template, i, i, i)
		returnValues += fmt.Sprintf("NODES(m%d) + ", i)
	}
	returnValues = strings.TrimRight(returnValues, "+ ")
	query = matchQueries

	if len(rule.Find.Who) > 0 {
		whereFilters = "WHERE ("
	}
	for i, who := range rule.Find.Who {
		block := fmt.Sprintf(`$who%d IN LABELS(who)`, i)
		arguments["who"+strconv.Itoa(i)] = cases.Title(language.Und).String(who)
		whereFilters += fmt.Sprintf("%s OR ", block)
	}
	if len(rule.Find.Who) > 0 {
		whereFilters = strings.TrimRight(whereFilters, "OR ")
		whereFilters += ") "
	}
	query += whereFilters
	query += fmt.Sprintf("\nWITH %s AS nds UNWIND nds as nd RETURN DISTINCT nd", returnValues)
	return
}

func prepareQueryPrivEsc(rule *Conf, arguments map[string]interface{}) (query string) {
	template := "MATCH m%d = (who)-[:HAS_POLICY]->(:Policy)-[:ALLOWS]->(:Action {Service: $service%d, Action: $action%d}) \n"

	matchQueries := ""
	whereFilters := ""
	returnValues := ""
	shortestPath := ""

	for i, perm := range rule.Find.With {
		withSplit := strings.Split(perm, ":")
		service := withSplit[0]
		action := withSplit[1]
		arguments["action"+strconv.Itoa(i)] = action
		arguments["service"+strconv.Itoa(i)] = service

		matchQueries += fmt.Sprintf(template, i, i, i)
		returnValues += fmt.Sprintf("NODES(m%d) + ", i)
	}
	returnValues = strings.TrimRight(returnValues, "+ ")
	query = matchQueries

	if len(rule.Find.Who) > 0 {
		whereFilters = "WHERE ("
	}
	for i, who := range rule.Find.Who {
		block := fmt.Sprintf(`$who%d IN LABELS(who)`, i)
		arguments["who"+strconv.Itoa(i)] = cases.Title(language.Und).String(who)
		whereFilters += fmt.Sprintf("%s OR ", block)
	}
	if len(rule.Find.Who) > 0 {
		whereFilters = strings.TrimRight(whereFilters, "OR ")
		whereFilters += ") "
	}

	if len(rule.Find.Target) > 0 {
		targetWhereFilter := "WHERE (%s) AND (%s)"
		targetWhereLabelFilters := ""
		targetWherePropertyFilters := ""
		for i, target := range rule.Find.Target {
			for what, id := range target {
				what = cases.Title(language.Und).String(what)

				blockLabel := fmt.Sprintf(`$targetType%d IN LABELS(target)`, i)
				arguments["targetType"+strconv.Itoa(i)] = what
				targetWhereLabelFilters += fmt.Sprintf("%s OR ", blockLabel)

				blockProperty := ""
				switch what {
				case "Policy":
					blockProperty = fmt.Sprintf(`target.Name = $target%d`, i)
				case "Role":
					blockProperty = fmt.Sprintf(`target.RoleName = $target%d`, i)
				case "Action":
					blockProperty = fmt.Sprintf(`target.Action = $target%d`, i)
				case "Group":
					blockProperty = fmt.Sprintf(`target.GroupName = $target%d`, i)
				case "User":
					blockProperty = fmt.Sprintf(`target.UserName = $target%d`, i)
				}
				arguments["target"+strconv.Itoa(i)] = id
				targetWherePropertyFilters += fmt.Sprintf("%s OR ", blockProperty)
			}
		}
		targetWhereLabelFilters = strings.TrimRight(targetWhereLabelFilters, "OR ")
		targetWherePropertyFilters = strings.TrimRight(targetWherePropertyFilters, "OR ")
		targetWhereFilter = fmt.Sprintf(targetWhereFilter, targetWhereLabelFilters, targetWherePropertyFilters)
		shortestPath = fmt.Sprintf("\nMATCH p0 = allShortestPaths((who)-[*1..10]->(target))\n%s", targetWhereFilter)
		returnValues += fmt.Sprintf(" + NODES(p%d)", 0)
	}
	query += whereFilters
	query += shortestPath
	query += fmt.Sprintf("\nWITH %s AS nds UNWIND nds as nd RETURN DISTINCT nd", returnValues)
	return
}

func prepareService(services []string, arguments map[string]interface{}) string {
	var query string
	query = "MATCH (s:Service)\nWHERE"
	for i, service := range services {
		service = cases.Title(language.Und).String(service)
		if i < len(services)-1 {
			query = fmt.Sprintf("%s $name%d IN LABELS(s) OR", query, i)
		} else {
			query = fmt.Sprintf("%s $name%d IN LABELS(s)\n", query, i)
		}
		arguments["name"+strconv.Itoa(i)] = service
	}
	query += "WITH s\n"
	return query
}

func prepareProperties(props []map[string]interface{}, arguments map[string]interface{}) string {
	var query = "MATCH (s)\n"
	var separator = "_"
	var count = 0

	if len(props) > 0 {
		query += "WHERE"
	}

	for _, prop := range props {
		subprops := walk(reflect.ValueOf(prop), separator)
		for _, p := range subprops {
			// Split the input in Key/Value
			key := strings.Split(p, "__")[0]
			value := strings.Split(p, "__")[1]
			arguments["key"+strconv.Itoa(count)] = key
			// Boolean must be used on queries
			if b, err := strconv.ParseBool(value); err == nil {
				arguments["value"+strconv.Itoa(count)] = b
			} else {
				arguments["value"+strconv.Itoa(count)] = value
			}
			query = fmt.Sprintf(`%s any(prop in keys(s) where toLower(prop) STARTS WITH toLower($key%d) AND s[prop] = $value%d) AND`,
				query, count, count)
			count++
		}
	}

	query = strings.TrimRight(query, " AND") + "\n"
	return query
}

//nolint:unused
func prepareExactResults(results []string, arguments map[string]interface{}) string {
	var query = "RETURN "

	for i, result := range results {
		if i < len(results)-1 {
			query = fmt.Sprintf("%s s[$result%d], ", query, i)
		} else {
			query = fmt.Sprintf("%s s[$result%d]", query, i)
		}
		arguments["result"+strconv.Itoa(i)] = result
	}
	return query
}

func prepareResults(results []string, arguments map[string]interface{}) string {
	var query = "RETURN s"
	return query
}

func walk(v reflect.Value, separator string) (output []string) {
	// Indirect through pointers and interfaces
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			output = append(output, walk(v.Index(i), separator)...)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			for _, e := range walk(v.MapIndex(k), separator) {
				output = append(output, fmt.Sprintf("%s%s%s", k, separator, e))
			}
		}
	default:
		output = append(output, fmt.Sprintf("%s%v", separator, v))
	}
	return output
}

func ArgsToQueryNeo4jBrowser(args map[string]interface{}) (argsOutput string) {
	for k, v := range args {
		switch val := v.(type) {
		case bool:
			argsOutput += fmt.Sprintf(`:param %s => %t; `, k, val)
		default:
			argsOutput += fmt.Sprintf(`:param %s => "%s"; `, k, val)
		}
	}
	argsOutput = strings.Trim(argsOutput, " ")
	return
}
