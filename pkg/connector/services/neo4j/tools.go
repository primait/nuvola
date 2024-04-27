package neo4j_connector

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	servicesDatabase "github.com/primait/nuvola/pkg/connector/services/aws/database"
	servicesEC2 "github.com/primait/nuvola/pkg/connector/services/aws/ec2"
	servicesIAM "github.com/primait/nuvola/pkg/connector/services/aws/iam"
	servicesLambda "github.com/primait/nuvola/pkg/connector/services/aws/lambda"
	servicesS3 "github.com/primait/nuvola/pkg/connector/services/aws/s3"
	"github.com/primait/nuvola/pkg/io/logging"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notdodo/arner"
	"github.com/notdodo/goflat/v2"
	"github.com/ohler55/ojg/oj"
)

type EnumAWSTypes interface {
	servicesS3.Bucket | servicesEC2.Instance | ec2types.Vpc | ec2types.VpcPeeringConnection | servicesLambda.Lambda | rdstypes.DBCluster | rdstypes.DBInstance | servicesDatabase.DynamoDB | servicesDatabase.RedshiftDB
}

var actionResourceRelations []map[string]string

func parseResources(resources any, service, action, policy, principal string) {
	replacer := strings.NewReplacer("${aws:username}", principal, "${", "", "}", "")
	switch v := resources.(type) {
	case []interface{}:
		for _, resource := range v {
			resource = replacer.Replace(resource.(string))
			arned, _ := arner.ParseARN(resource.(string))
			itemAR := make(map[string]string)
			itemAR["arn"] = resource.(string)
			itemAR["policy"] = policy
			itemAR["service"] = strings.ToLower(service)
			itemAR["action"] = action
			itemAR["resource"] = arned.Resource
			actionResourceRelations = append(actionResourceRelations, itemAR)
		}
	case string:
		v = replacer.Replace(v)
		arned, _ := arner.ParseARN(v)
		itemAR := make(map[string]string)
		itemAR["arn"] = v
		itemAR["policy"] = policy
		itemAR["service"] = strings.ToLower(service)
		itemAR["action"] = action
		itemAR["resource"] = arned.Resource
		actionResourceRelations = append(actionResourceRelations, itemAR)
	}
}

func (nc *Neo4jClient) createPolicyRelationships(idPolicy int64, statements *[]servicesIAM.Statement, principal string) {
	// Prepare the map for the UNWIND syntax
	actions := make(map[string]interface{})
	actions["actions"] = make([]map[string]string, 0)

	for _, statement := range *statements {
		if statement.Effect == "Allow" {
			items := make([]map[string]string, 0)

			switch v := statement.Action.(type) {
			case []interface{}:
				// list of Actions
				for _, serviceAction := range v {
					item := make(map[string]string)
					actionStr := strings.Split(serviceAction.(string), ":")
					service := actionStr[0]
					action := actionStr[1]
					item["service"] = strings.ToLower(service)
					item["action"] = action
					item["policy"] = fmt.Sprint(idPolicy)
					items = append(items, item)
					parseResources(statement.Resource, service, action, strconv.Itoa(int(idPolicy)), principal)
				}
			case string:
				// single Action
				item := make(map[string]string)
				actionStr := strings.Split(v, ":")
				service := actionStr[0]
				action := actionStr[1]
				item["service"] = strings.ToLower(service)
				item["action"] = action
				item["policy"] = fmt.Sprint(idPolicy)
				items = append(items, item)
				parseResources(statement.Resource, service, action, strconv.Itoa(int(idPolicy)), principal)
			default:
				logging.HandleError(nil, "Neo4j", fmt.Sprintf("createPolicyRelationships - case not implemented for %v of type: %v", statement.Action, v))
			}

			// Append all actions of this statement to the UNWIND map
			actions["actions"] = append(actions["actions"].([]map[string]string), items...)
		}
	}

	if actions["actions"] != nil {
		session := nc.NewSession()
		defer session.Close(context.TODO())
		_, err := session.ExecuteWrite(context.TODO(), func(tx neo4j.ManagedTransaction) (any, error) {
			linkPolicy := `UNWIND $actions AS actions
				MATCH (p:Policy) WHERE id(p) = toInteger(actions.policy)
				MERGE (p)-[:ALLOWS]->(:Action {Action: actions.action, Service: actions.service})`
			var result, err = tx.Run(context.TODO(), linkPolicy, actions)
			if err != nil {
				return nil, err
			}

			return result.Consume(context.TODO())
		})

		if err != nil {
			logging.HandleError(err, "Neo4j - createPolicyRelationships", "Error on executing query")
		}
	}
}

func flatObjects[N EnumAWSTypes](o []N) (result map[string]interface{}) {
	result = make(map[string]interface{}, 0)
	items := make([]map[string]interface{}, 0, len(o))

	for _, obj := range o {
		jsonString, _ := oj.Marshal(obj)
		flat, _ := goflat.FlatJSON(string(jsonString), goflat.FlattenerConfig{
			Prefix:    "",
			Separator: "_",
			OmitNil:   true,
			OmitEmpty: true,
		})
		flatObject := make(map[string]interface{})
		oj.Unmarshal([]byte(flat), &flatObject)
		items = append(items, flatObject)
	}

	result["objects"] = items
	return
}

func uniqueActionsResources(slice *[]map[string]string) (list []map[string]string) {
	keys := make(map[string]bool)

	for _, v := range *slice {
		key := v["policy"] + v["service"] + v["action"] + v["resource"]
		if _, value := keys[key]; !value {
			keys[key] = true
			list = append(list, v)
		}
	}
	return list
}
