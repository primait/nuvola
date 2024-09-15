terraform {
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      configuration_aliases = [aws, aws.us]
    }
  }
}

data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "${path.module}/../${path.module}/handler.js"
  output_path = "${path.module}/../${path.module}/lambda.zip"
}


resource "aws_vpc" "lambda_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_internet_gateway" "lambda_igw" {
  vpc_id = aws_vpc.lambda_vpc.id
}

resource "aws_subnet" "lambda_subnet_a" {
  vpc_id            = aws_vpc.lambda_vpc.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "eu-west-1a"
}

resource "aws_route_table" "lambda_route_table" {
  vpc_id = aws_vpc.lambda_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.lambda_igw.id
  }
}

resource "aws_route_table_association" "rds_cluster_subnet_a_association" {
  subnet_id      = aws_subnet.lambda_subnet_a.id
  route_table_id = aws_route_table.lambda_route_table.id
}

resource "aws_security_group" "lambda_sg" {
  name   = "lambda-sg"
  vpc_id = aws_vpc.lambda_vpc.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lambda_function" "lambda_1" {
  provider         = aws.us
  filename         = "${path.module}/../${path.module}/lambda.zip"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  function_name    = "lambda-1"
  description      = "Lambda One"
  role             = var.lambda_1_role
  publish          = true
  timeout          = "5"
  handler          = "index.handler"
  runtime          = "nodejs18.x"
  memory_size      = 512

  tags = {
    Name = "${var.name}-lambda-1"
  }

  environment {
    variables = {
      SECRET = "SuperS3cr3t!"
    }
  }
}

resource "aws_lambda_function" "lambda_2" {
  filename         = "${path.module}/../${path.module}/lambda.zip"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  function_name    = "lambda-2"
  description      = "Lambda Two"
  role             = var.lambda_2_role
  publish          = true
  timeout          = "5"
  handler          = "index.handler"
  runtime          = "nodejs18.x"

  vpc_config {
    subnet_ids         = [aws_subnet.lambda_subnet_a.id]
    security_group_ids = [aws_security_group.lambda_sg.id]
  }
  tags = {
    Name = "${var.name}-lambda-2"
  }
}
