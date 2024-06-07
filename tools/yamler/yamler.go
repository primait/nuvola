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
	logger      logging.LogManager
}

type Find struct {
	Who    []string
	To     []string
	With   []string            `yaml:"with,omitempty"`
	Target []map[string]string `yaml:"target,omitempty"`
}

func GetConf(file string) (c *Conf) {
	logger := logging.GetLogManager()
	c = &Conf{Enabled: true, logger: logger}
	yamlFile, err := os.ReadFile(files.NormalizePath(file))
	if err != nil {
		logger.Error("Error on reading rule file", "err", err)
	}
	c.Enabled = true // Default value is: Enabled
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logger.Error("Error unmarshalling yamlFile", "err", err)
	}

	return c
}

func PrepareQuery(config *Conf) (query string, arguments map[string]interface{}) {
	arguments = make(map[string]interface{}, 0)
	if len(config.Services) > 0 {
		// Direct access to properties
		query = prepareService(config.Services, arguments) +
			prepareProperties(config.Properties, arguments) +
			prepareResults(config.Return, arguments)
	} else if len(config.Find.To) > 0 || len(config.Find.Who) > 0 || len(config.Find.With) > 0 {
		if len(config.Find.Target) > 0 {
			query = prepareQueryPrivEsc(config, arguments)
		} else {
			query = preparePathQuery(config, arguments)
		}
	} else {
		config.logger.Error("Malformed rule")
	}
	return query, arguments
}

func preparePathQuery(rule *Conf, arguments map[string]interface{}) string {
	template := "MATCH m%d = (who)-[:HAS_POLICY]->(:Policy)-[:ALLOWS]->(:Action {Service: $service%d, Action: $action%d}) \n"
	var matchQueries, whereFilters, returnValues strings.Builder

	for i, perm := range rule.Find.With {
		withSplit := strings.Split(perm, ":")
		service, action := withSplit[0], withSplit[1]
		arguments[fmt.Sprintf("action%d", i)] = action
		arguments[fmt.Sprintf("service%d", i)] = service

		matchQueries.WriteString(fmt.Sprintf(template, i, i, i))
		returnValues.WriteString(fmt.Sprintf("NODES(m%d) + ", i))
	}
	query := matchQueries.String()
	returnValuesStr := strings.TrimSuffix(returnValues.String(), " + ")

	if len(rule.Find.Who) > 0 {
		whereFilters.WriteString("WHERE (")
		for i, who := range rule.Find.Who {
			arguments[fmt.Sprintf("who%d", i)] = cases.Title(language.Und).String(who)
			whereFilters.WriteString(fmt.Sprintf(`$who%d IN LABELS(who) OR `, i))
		}
		whereFiltersStr := strings.TrimSuffix(whereFilters.String(), " OR ")
		query += whereFiltersStr + ") "
	}
	query += fmt.Sprintf("\nWITH %s AS nds UNWIND nds as nd RETURN DISTINCT nd", returnValuesStr)
	return query
}

func prepareQueryPrivEsc(rule *Conf, arguments map[string]interface{}) string {
	template := "MATCH m%d = (who)-[:HAS_POLICY]->(:Policy)-[:ALLOWS]->(:Action {Service: $service%d, Action: $action%d}) \n"
	var matchQueries, whereFilters, returnValues, shortestPath strings.Builder

	for i, perm := range rule.Find.With {
		withSplit := strings.Split(perm, ":")
		service, action := withSplit[0], withSplit[1]
		arguments[fmt.Sprintf("action%d", i)] = action
		arguments[fmt.Sprintf("service%d", i)] = service

		matchQueries.WriteString(fmt.Sprintf(template, i, i, i))
		returnValues.WriteString(fmt.Sprintf("NODES(m%d) + ", i))
	}
	returnValuesStr := strings.TrimSuffix(returnValues.String(), " + ")
	query := matchQueries.String()

	if len(rule.Find.Who) > 0 {
		whereFilters.WriteString("WHERE (")
		for i, who := range rule.Find.Who {
			arguments[fmt.Sprintf("who%d", i)] = cases.Title(language.Und).String(who)
			whereFilters.WriteString(fmt.Sprintf(`$who%d IN LABELS(who) OR `, i))
		}
		whereFiltersStr := strings.TrimSuffix(whereFilters.String(), " OR ")
		query += whereFiltersStr + ") "
	}

	if len(rule.Find.Target) > 0 {
		var targetWhereLabelFilters, targetWherePropertyFilters strings.Builder
		for i, target := range rule.Find.Target {
			for what, id := range target {
				what = cases.Title(language.Und).String(what)
				arguments[fmt.Sprintf("targetType%d", i)] = what
				targetWhereLabelFilters.WriteString(fmt.Sprintf(`$targetType%d IN LABELS(target) OR `, i))

				var blockProperty string
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
				arguments[fmt.Sprintf("target%d", i)] = id
				targetWherePropertyFilters.WriteString(fmt.Sprintf("%s OR ", blockProperty))
			}
		}
		targetWhereLabelFiltersStr := strings.TrimSuffix(targetWhereLabelFilters.String(), " OR ")
		targetWherePropertyFiltersStr := strings.TrimSuffix(targetWherePropertyFilters.String(), " OR ")
		shortestPath.WriteString(fmt.Sprintf("\nMATCH p0 = allShortestPaths((who)-[*1..10]->(target))\nWHERE (%s) AND (%s)", targetWhereLabelFiltersStr, targetWherePropertyFiltersStr))
		returnValues.WriteString(" + NODES(p0)")
	}
	query += shortestPath.String()
	query += fmt.Sprintf("\nWITH %s AS nds UNWIND nds as nd RETURN DISTINCT nd", returnValuesStr)
	fmt.Println(query)
	return query
}

func prepareService(services []string, arguments map[string]interface{}) string {
	var query strings.Builder
	query.WriteString("MATCH (s:Service)\nWHERE")
	for i, service := range services {
		service = cases.Title(language.Und).String(service)
		query.WriteString(fmt.Sprintf(" $name%d IN LABELS(s)", i))
		if i < len(services)-1 {
			query.WriteString(" OR")
		}
		arguments[fmt.Sprintf("name%d", i)] = service
	}
	query.WriteString("\nWITH s\n")
	return query.String()
}

func prepareProperties(props []map[string]interface{}, arguments map[string]interface{}) string {
	if len(props) == 0 {
		return "MATCH (s)\n"
	}

	var query strings.Builder
	separator := "_"
	query.WriteString("MATCH (s)\nWHERE")
	var count int
	for _, prop := range props {
		subprops := walk(reflect.ValueOf(prop), separator)
		for _, p := range subprops {
			key, value := strings.Split(p, "__")[0], strings.Split(p, "__")[1]
			arguments[fmt.Sprintf("key%d", count)] = key
			if b, err := strconv.ParseBool(value); err == nil {
				arguments[fmt.Sprintf("value%d", count)] = b
			} else {
				arguments[fmt.Sprintf("value%d", count)] = value
			}
			query.WriteString(fmt.Sprintf(` any(prop in keys(s) where toLower(prop) STARTS WITH toLower($key%d) AND s[prop] = $value%d) AND`, count, count))
			count++
		}
	}
	queryStr := strings.TrimSuffix(query.String(), " AND") + "\n"
	return queryStr
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

func walk(v reflect.Value, separator string) []string {
	var output []string
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
