package neo4j_client

import (
	"fmt"
	"log"
	awsconfig "nuvola/config/aws"
	servicesDatabase "nuvola/dump/aws/database"
	servicesEC2 "nuvola/dump/aws/ec2"
	servicesIAM "nuvola/dump/aws/iam"
	servicesLambda "nuvola/dump/aws/lambda"
	servicesS3 "nuvola/dump/aws/s3"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (nc *Connector) AddUsers(users *[]servicesIAM.User) {
	for i := range *users {
		idUser := nc.createUser(&(*users)[i])
		for _, inlinePolicy := range (*users)[i].InlinePolicies {
			idPolicy := nc.createPolicyUser(idUser, "", inlinePolicy.PolicyName, "inline")
			nc.createPolicyRelationships(idPolicy, &inlinePolicy.Statement)
		}

		for _, attachedPolicy := range (*users)[i].AttachedPolicies {
			idPolicy := nc.createPolicyUser(idUser, *attachedPolicy.PolicyArn, *attachedPolicy.PolicyName, "attached")
			nc.createPolicyRelationships(idPolicy, &attachedPolicy.Versions[0].Document.Statement)
		}
	}
}

func (nc *Connector) AddGroups(groups *[]servicesIAM.Group) {
	for i := range *groups {
		idGroup := nc.createGroup(&(*groups)[i])
		for _, inlinePolicy := range (*groups)[i].InlinePolicies {
			idPolicy := nc.createPolicyGroup(idGroup, "", inlinePolicy.PolicyName, "inline")
			nc.createPolicyRelationships(idPolicy, &inlinePolicy.Statement)
		}

		for _, attachedPolicy := range (*groups)[i].AttachedPolicies {
			idPolicy := nc.createPolicyGroup(idGroup, *attachedPolicy.PolicyArn, *attachedPolicy.PolicyName, "attached")
			nc.createPolicyRelationships(idPolicy, &attachedPolicy.Versions[0].Document.Statement)
		}
	}
}

func (nc *Connector) AddRoles(roles *[]servicesIAM.Role) {
	for i := range *roles {
		idRole := nc.createRole(&(*roles)[i])
		for _, inlinePolicy := range (*roles)[i].InlinePolicies {
			idPolicy := nc.createPolicyRole(idRole, "", inlinePolicy.PolicyName, "inline")
			nc.createPolicyRelationships(idPolicy, &inlinePolicy.Statement)
		}

		for _, attachedPolicy := range (*roles)[i].AttachedPolicies {
			idPolicy := nc.createPolicyRole(idRole, *attachedPolicy.PolicyArn, *attachedPolicy.PolicyName, "attached")
			nc.createPolicyRelationships(idPolicy, &attachedPolicy.Versions[0].Document.Statement)
		}
	}
}

func (nc *Connector) createGroup(group *servicesIAM.Group) int64 {
	session := nc.NewWriteSession()
	defer session.Close()
	query := `MERGE (g:IAM:Group {GroupName: $GroupName, CreateDate: $CreateDate, Arn: $Arn, Path: $Path, GroupId: $GroupId}) RETURN id(g)`

	idGroup, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"GroupName":  group.GroupName,
			"CreateDate": fmt.Sprint(group.CreateDate),
			"Arn":        group.Arn,
			"Path":       group.Path,
			"GroupId":    group.GroupId,
		})

		if err != nil {
			return nil, err
		}

		result.Next()
		return result.Record().Values[0].(int64), result.Err()
	})

	if err != nil {
		log.Fatalln(err)
	}
	return idGroup.(int64)
}

func (nc *Connector) createPolicyGroup(idGroup int64, policyArn string, name string, policyType string) int64 {
	session := nc.NewWriteSession()
	defer session.Close()
	query := `%s
		WITH policy
		MATCH (g:IAM:Group) WHERE id(g) = $idGroup
		MERGE (g)-[:HAS_POLICY]->(policy)
		RETURN id(policy)`

	if policyType == "attached" {
		query = fmt.Sprintf(query, `CALL apoc.merge.node(["Policy", "Attached"], {Name: $Name, Type: $Type, Arn: $PolicyArn}) YIELD node AS policy`)
	} else if policyType == "inline" {
		query = fmt.Sprintf(query, `CALL apoc.create.node(["Policy", "Inline"], {Name: $Name, Type: $Type}) YIELD node AS policy`)
	}

	idPolicy, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"idGroup":   idGroup,
			"Name":      name,
			"Type":      policyType,
			"PolicyArn": policyArn,
		})

		if err != nil {
			return nil, err
		}

		result.Next()
		return result.Record().Values[0].(int64), result.Err()
	})

	if err != nil {
		log.Fatalln(err)
	}
	return idPolicy.(int64)
}

func (nc *Connector) createUser(user *servicesIAM.User) int64 {
	groupNames := make([]string, 0)
	query := `MERGE (u:IAM:User {
		UserName: $UserName, 
		Arn: $Arn,
		UserId: $UserId,
		PasswordEnabled: $PasswordEnabled,
		PasswordLastChanged: $PasswordLastChanged,
		MFAStatus: $MFAStatus})
		RETURN id(u)`

	session := nc.NewWriteSession()
	defer session.Close()
	idUser, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"UserName":            user.UserName,
			"UserId":              user.UserId,
			"Arn":                 user.Arn,
			"PasswordEnabled":     user.PasswordEnabled,
			"PasswordLastChanged": user.PasswordLastChanged,
			"MFAStatus":           user.MfaActive,
		})

		if err != nil {
			return nil, err
		}

		result.Next()
		return result.Record().Values[0].(int64), result.Err()
	})

	if err != nil {
		log.Fatalln(err)
	}

	for g := 0; g < len(user.Groups); g++ {
		groupNames = append(groupNames, aws.ToString(user.Groups[g].GroupName))

		session := nc.NewWriteSession()
		defer session.Close()
		queryGroup := `MATCH (g:Group {GroupName: $GroupName, CreateDate: $CreateDate, Arn: $Arn, GroupId: $GroupId, Path: $Path}) WITH g
					   MATCH (u:User {Arn: $uArn})
					   SET u.Groups = $groups
					   WITH u, g
					   MERGE (u)-[:MEMBER_OF]->(g)`

		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			var result, err = tx.Run(queryGroup, map[string]interface{}{
				"GroupName":  user.Groups[g].GroupName,
				"CreateDate": fmt.Sprint(user.Groups[g].CreateDate),
				"Arn":        user.Groups[g].Arn,
				"GroupId":    user.Groups[g].GroupId,
				"Path":       user.Groups[g].Path,
				"uArn":       user.Arn,
				"groups":     groupNames,
			})

			if err != nil {
				return nil, err
			}

			return result.Consume()
		})
		if err != nil {
			log.Fatalln(err)
		}
	}

	return idUser.(int64)
}

func (nc *Connector) createPolicyUser(idUser int64, policyArn string, name string, policyType string) int64 {
	session := nc.NewWriteSession()
	defer session.Close()
	query := `%s
		WITH policy
		MATCH (u:IAM:User) WHERE id(u) = $idUser
		MERGE (u)-[r:HAS_POLICY]->(policy)
		RETURN id(policy)`

	if policyType == "attached" {
		query = fmt.Sprintf(query, `CALL apoc.merge.node(["Policy", "Attached", "IAM"], {Name: $Name, Type: $Type, Arn: $PolicyArn}) YIELD node AS policy`)
	} else if policyType == "inline" {
		query = fmt.Sprintf(query, `CALL apoc.create.node(["Policy", "Inline", "IAM"], {Name: $Name, Type: $Type}) YIELD node AS policy`)
	}

	idPolicy, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"idUser":    idUser,
			"Name":      name,
			"Type":      policyType,
			"PolicyArn": policyArn,
		})

		if err != nil {
			return nil, err
		}

		result.Next()
		return result.Record().Values[0].(int64), result.Err()
	})

	if err != nil {
		log.Fatalln(err)
	}
	return idPolicy.(int64)
}

func (nc *Connector) createRole(role *servicesIAM.Role) int64 {
	session := nc.NewWriteSession()
	defer session.Close()
	query := ""
	if role.InstanceProfileId != "" {
		query = `MERGE (r:IAM:Role:InstanceProfile {RoleName: $RoleName, Arn: $Arn, Path: $Path, Description: $Description, RoleId: $RoleId, InstanceProfileArn: $InstanceProfileArn, IamInstanceProfileId: $IamInstanceProfileId, AssumableBy: $AssumableBy}) RETURN id(r)`
	} else {
		query = `MERGE (r:IAM:Role {RoleName: $RoleName, Arn: $Arn, Path: $Path, Description: $Description, RoleId: $RoleId, AssumableBy: $AssumableBy}) RETURN id(r)`
	}

	idRole, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"RoleName":             role.RoleName,
			"Arn":                  role.Arn,
			"Path":                 role.Path,
			"RoleId":               role.RoleId,
			"InstanceProfileArn":   role.InstanceProfileArn,
			"IamInstanceProfileId": role.InstanceProfileId,
			"AssumableBy":          role.AssumableBy,
			"Description":          role.Description,
		})

		if err != nil {
			return nil, err
		}

		result.Next()
		return result.Record().Values[0].(int64), result.Err()
	})

	if err != nil {
		log.Fatalln(err)
	}
	return idRole.(int64)
}

func (nc *Connector) createPolicyRole(idRole int64, policyArn string, name string, policyType string) int64 {
	session := nc.NewWriteSession()
	defer session.Close()
	query := `%s
		WITH policy
		MATCH (u:Role) WHERE id(u) = $idRole
		MERGE (u)-[r:HAS_POLICY]->(policy)
		RETURN id(policy)`

	if policyType == "attached" {
		query = fmt.Sprintf(query, `CALL apoc.merge.node(["Policy", "Attached", "IAM"], {Name: $Name, Type: $Type, Arn: $PolicyArn}) YIELD node AS policy`)
	} else if policyType == "inline" {
		query = fmt.Sprintf(query, `CALL apoc.create.node(["Policy", "Inline", "IAM"], {Name: $Name, Type: $Type}) YIELD node AS policy`)
	}

	idPolicy, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"idRole":    idRole,
			"Name":      name,
			"Type":      policyType,
			"PolicyArn": policyArn,
		})

		if err != nil {
			return nil, err
		}

		result.Next()
		return result.Record().Values[0].(int64), result.Err()
	})

	if err != nil {
		log.Fatalln(err)
	}
	return idPolicy.(int64)
}

func (nc *Connector) AddObjects(result map[string]interface{}, query string) {
	session := nc.NewWriteSession()
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, result)

		if err != nil {
			return nil, err
		}

		return result.Consume()
	})

	if err != nil {
		log.Fatalln(err)
	}
}

func (nc *Connector) addLinksToResources(service string, property string) {
	session := nc.NewWriteSession()
	defer session.Close()
	actionResourceRelations = uniqueActionsResources(&actionResourceRelations)
	// Filter only related service relationships
	var out []map[string]string
	for _, v := range actionResourceRelations {
		if strings.EqualFold(v["service"], service) && !strings.Contains(v["action"], "Create") {
			out = append(out, v)
		}
	}

	// https://neo4j.com/labs/apoc/4.4/overview/apoc.periodic/apoc.periodic.iterate/
	query := `CALL apoc.periodic.iterate("
		UNWIND $actionResourceMap AS armap
		MATCH (p:Policy)-[:ALLOWS]->(a:Action {Service: armap.service, Action: armap.action})
		WHERE id(p) = toInteger(armap.policy)
		MATCH (s:Service:` + cases.Title(language.Und).String(service) + `) WHERE
			(s.` + property + ` = armap.resource) OR
			(armap.resource = '') OR
			(armap.resource CONTAINS '*' AND s.` + property + ` =~ replace(armap.resource, '*', '.*?'))
		RETURN a, s",
		"MERGE (a)-[:ON]->(s)",
		{batchSize:10000, iterateList:true, params: {actionResourceMap: $actionResourceMap}})`
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"actionResourceMap": out,
		})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})

	if err != nil {
		log.Fatalln(err)
	}
}

func (nc *Connector) AddLinksToResourcesIAM() {
	session := nc.NewWriteSession()
	defer session.Close()
	// #nosec
	session.Run("CALL db.awaitIndexes(3000)", nil) //nolint:all

	// Filter only related service relationships
	var out []map[string]string
	actionResourceRelations = uniqueActionsResources(&actionResourceRelations)
	for _, v := range actionResourceRelations {
		if strings.EqualFold(v["service"], "iam") {
			v["resourceType"] = awsconfig.IAMActionResourceMap[v["action"]]
			out = append(out, v)
		}
	}

	query := `CALL apoc.periodic.iterate("
		UNWIND $actionResourceMap AS armap
		MATCH (p:Policy)-[:ALLOWS]->(act:Action {Service: 'iam', Action: armap.action})
		WHERE id(p) = toInteger(armap.policy)
		MATCH (principal:IAM) WHERE
			(
				(armap.resourceType = '*') OR
				(armap.resourceType <> '' AND armap.resourceType IN LABELS(principal))
			)
			AND
			(
				(principal.Arn = armap.arn) OR
				(armap.arn = '') OR
				(armap.arn CONTAINS '*' AND 
					principal.Arn =~ replace(armap.arn, '*', '.*?') OR 
					principal.InstanceProfileArn =~ replace(armap.arn, '*', '.*?')
				)
			)
		RETURN act, principal",
		"MERGE (act)-[:ON]->(principal)",
		{batchSize:5000, iterateList:true, params: {actionResourceMap: $actionResourceMap}})`
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run(query, map[string]interface{}{
			"actionResourceMap": out,
		})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})

	if err != nil {
		log.Fatalln(err, query)
	}
}

func (nc *Connector) AddBuckets(buckets *[]servicesS3.Bucket) {
	query := `UNWIND $objects AS bucket			
		CREATE (s:S3:Service)
		SET s = bucket`
	nc.AddObjects(flatObjects(*buckets), query)
	nc.addLinksToResources("s3", "Name")
}

func (nc *Connector) AddEC2(instances *[]servicesEC2.Instance) {
	query := `UNWIND $objects AS instance
		CREATE (e:Ec2:Service)
		SET e = instance`
	nc.AddObjects(flatObjects(*instances), query)

	session := nc.NewWriteSession()
	defer session.Close()
	_, err := session.Run(`call apoc.periodic.iterate("
		MATCH (role:Role) WHERE role.InstanceProfileArn <> '' RETURN role",
		"MATCH (n:Ec2:Service) WHERE
			n.IamInstanceProfile_Arn = role.InstanceProfileArn
		MERGE (n)-[:USES]->(role)", {batchSize:10000, parallel:true, iterateList:true})`, nil)
	if err != nil {
		log.Fatalln(err)
	}
	nc.addLinksToResources("ec2", "InstanceId")
}

func (nc *Connector) AddVPC(vpcs *servicesEC2.VPC) {
	queryVPC := `UNWIND $objects AS vpcs
		CREATE (vpc:Vpc:Service)
		SET vpc = vpcs
		SET vpc.Type = "Internal"
		WITH vpc, vpcs.VpcId AS vpcid
		MATCH (ec2:Ec2 {VpcId: vpcid})
		MERGE (ec2)-[:NETWORK]->(vpc)`

	queryPeering := `UNWIND $objects AS peerings
		WITH peerings, peerings.RequesterVpcInfo_VpcId AS vpcid_req, peerings.AccepterVpcInfo_VpcId AS vpcid_acc,
			 peerings.RequesterVpcInfo_OwnerId as req_id, peerings.AccepterVpcInfo_OwnerId as acc_id
		MERGE (req:Vpc {VpcId: vpcid_req, OwnerId: req_id})
		MERGE (acc:Vpc {VpcId: vpcid_acc, OwnerId: acc_id})
		SET (CASE WHEN req.Type IS NULL THEN req END).Type = 'External'
		SET (CASE WHEN acc.Type IS NULL THEN acc END).Type = 'External'
		WITH req, acc, peerings
		CALL apoc.merge.relationship(req, "PEERING", peerings, {}, acc, {}) YIELD rel
		RETURN rel`

	nc.AddObjects(flatObjects(vpcs.VPCs), queryVPC)
	nc.AddObjects(flatObjects(vpcs.Peerings), queryPeering)
}

func (nc *Connector) AddLambda(lambdas *[]servicesLambda.Lambda) {
	query := `UNWIND $objects AS lambdas
		CREATE (lbd:Lambda:Service)
		SET lbd = lambdas`

	nc.AddObjects(flatObjects(*lambdas), query)

	session := nc.NewWriteSession()
	defer session.Close()
	_, err := session.Run(`call apoc.periodic.iterate(
		"MATCH (role:Role) RETURN role",
		"MATCH (n:Lambda:Service) WHERE n.Role = role.Arn MERGE (n)-[:USES]->(role)",
		{batchSize:10000, parallel:true, iterateList:true})`, nil)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = session.Run(`call apoc.periodic.iterate(
		"MATCH (l:Lambda) WHERE l.VpcConfig_VpcId <> '' MATCH (v:Vpc) WHERE v.VpcId = l.VpcConfig_VpcId RETURN l, v",
		"MERGE (l)-[:NETWORK]->(v)",
		{batchSize:10000, parallel:true, iterateList:true})`, nil)
	if err != nil {
		log.Fatalln(err)
	}
	nc.addLinksToResources("lambda", "FunctionName")
}

func (nc *Connector) AddRDS(rdsdbs *servicesDatabase.RDS) {
	query := `UNWIND $objects AS rds				
		CREATE (s:Rds:Service)
		SET s = rds`

	nc.AddObjects(flatObjects(rdsdbs.Clusters), query)
	nc.AddObjects(flatObjects(rdsdbs.Instances), query)
	nc.addLinksToResources("rds", "DBClusterIdentifier")
	nc.addLinksToResources("rds", "DBInstanceIdentifier")

}

func (nc *Connector) AddDynamoDB(dynamodbs *[]servicesDatabase.DynamoDB) {
	query := `UNWIND $objects AS dynamodb				
		CREATE (s:Dynamodb:Service)
		SET s = dynamodb`

	nc.AddObjects(flatObjects(*dynamodbs), query)
	nc.addLinksToResources("dynamodb", "Name")
}

func (nc *Connector) AddRedshift(redshifts *[]servicesDatabase.RedshiftDB) {
	query := `UNWIND $objects AS redshift				
		CREATE (s:Redshift:Service)
		SET s = redshift`
	nc.AddObjects(flatObjects(*redshifts), query)
	nc.addLinksToResources("redshift", "DBName")

	session := nc.NewWriteSession()
	defer session.Close()
	_, err := session.Run(`call apoc.periodic.iterate(
		"MATCH (r:Redshift) WHERE r.VpcId <> '' MATCH (v:Vpc) WHERE v.VpcId = r.VpcId RETURN r, v",
		"MERGE (r)-[:NETWORK]->(v)",
		{batchSize:10000, parallel:true, iterateList:true})`, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
