module "vpc" {
  source = "./vpc"
  name   = var.name
  region = var.region
}

module "iam" {
  source = "./iam"
  name   = var.name
}

module "cfn" {
  source   = "./cloudformation"
  name     = var.name
  iam_role = module.iam.service_deployer_arn
}

module "ec2" {
  source           = "./ec2"
  name             = var.name
  region           = var.region
  subnet_id        = module.vpc.subnet_id
  sg_id            = module.vpc.sg_id
  instance_profile = module.iam.instance_profile
}

module "lambda" {
  source                 = "./lambda"
  name                   = var.name
  vulnerable_lambda_role = module.iam.lambda_runner_role
  dummy_lambda_role      = module.iam.temp_backend_api_role
}

module "s3" {
  source = "./s3"
  name   = var.name
}
