package neo4j_client

import (
	"fmt"
	"log"
	"nuvola/io/yamler"
	"reflect"
	"strconv"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (nc *Connector) PrepareQuery(config *yamler.Conf) (query string, arguments map[string]interface{}) {
	// Convention: white spaces always to end of the concatenation or insertion point
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
		log.Fatalln("Malformed rule!", config.Name)
	}
	return query, arguments
}

func preparePathQuery(rule *yamler.Conf, arguments map[string]interface{}) (query string) {
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

func prepareQueryPrivEsc(rule *yamler.Conf, arguments map[string]interface{}) (query string) {
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
	query = query + "WITH s\n"
	return query
}

func prepareProperties(props []map[string]interface{}, arguments map[string]interface{}) string {
	var query string = "MATCH (s)\n"
	var separator string = "_"
	var count int = 0

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
			count += 1
		}
	}

	query = strings.TrimRight(query, " AND") + "\n"
	return query
}

//nolint:unused
func prepareExactResults(results []string, arguments map[string]interface{}) string {
	var query string = "RETURN "

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
	var query string = "RETURN s"
	return query
}

func (nc *Connector) Query(query string, arguments map[string]interface{}) []map[string]interface{} {
	session := nc.NewReadSession()
	defer session.Close()

	results, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, arguments)
		if err != nil {
			log.Fatalln(err, query, arguments)
		}

		results := make([]map[string]interface{}, 0)
		for result.Next() {
			record, ok := result.Record().Get("result")
			if ok {
				nodeAttributes := record.(dbtype.Node).Props
				results = append(results, nodeAttributes)
			} else {
				// iterates through all results
				keys, ok := result.Keys()
				if ok == nil {
					for _, key := range keys {
						nodesMap, _ := result.Record().Get(key)
						nodeAttributes := nodesMap.(dbtype.Node).Props
						results = append(results, nodeAttributes)
					}
				}
			}

		}
		return results, result.Err()
	})
	if err != nil {
		log.Panicln(err)
	}

	return results.([]map[string]interface{})
}
