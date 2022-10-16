package neo4j_connector

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	servicesDatabase "nuvola/connector/services/aws/database"
	servicesEC2 "nuvola/connector/services/aws/ec2"
	servicesIAM "nuvola/connector/services/aws/iam"
	servicesLambda "nuvola/connector/services/aws/lambda"
	servicesS3 "nuvola/connector/services/aws/s3"
	nuvolaerror "nuvola/tools/error"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notdodo/arner"
	"github.com/notdodo/goflat"
)

type EnumAWSTypes interface {
	servicesS3.Bucket | servicesEC2.Instance | ec2types.Vpc | ec2types.VpcPeeringConnection | servicesLambda.Lambda | rdstypes.DBCluster | rdstypes.DBInstance | servicesDatabase.DynamoDB | servicesDatabase.RedshiftDB
}

var actionResourceRelations []map[string]string

func parseResources(resources any, service, action, policy string) {
	switch v := resources.(type) {
	case []interface{}:
		for _, resource := range v {
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

func (nc *Neo4jClient) createPolicyRelationships(idPolicy int64, statements *[]servicesIAM.Statement) {
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
					parseResources(statement.Resource, service, action, strconv.Itoa(int(idPolicy)))
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
				parseResources(statement.Resource, service, action, strconv.Itoa(int(idPolicy)))
			default:
				nuvolaerror.HandleError(nil, "Neo4j", "createPolicyRelationships - case not implemented")
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
			nuvolaerror.HandleError(err, "Neo4j - createPolicyRelationships", "Error on executing query")
		}
	}
}

func flatObjects[N EnumAWSTypes](o []N) (result map[string]interface{}) {
	result = make(map[string]interface{}, 0)
	items := make([]map[string]interface{}, 0)

	for _, obj := range o {
		flat, err := goflat.FlatStruct(obj, "", "_")
		if err != nil {
			log.Fatalln(err)
		}
		items = append(items, flat)
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
