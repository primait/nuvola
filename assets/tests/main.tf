terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

locals {
  default_tags = {
    Environment = "test"
    Owner       = "primait"
    Project     = "nuvola"
  }
}

provider "aws" {
  access_key = "test"
  secret_key = "test"
  region     = "eu-west-1"
  default_tags {
    tags = local.default_tags
  }
}

provider "aws" {
  access_key = "test"
  secret_key = "test"
  region     = "us-east-1"
  alias      = "us_east_1"
  default_tags {
    tags = local.default_tags
  }
}

module "iam" {
  source = "./iam"
  name   = var.name
}

module "s3" {
  source = "./s3"
  name   = var.name
}

module "lambda" {
  source        = "./lambda"
  name          = var.name
  lambda_1_role = module.iam.lambda_runner_role.arn
  lambda_2_role = module.iam.lambda_admin_role.arn
  providers = {
    aws    = aws
    aws.us = aws.us_east_1
  }
}
