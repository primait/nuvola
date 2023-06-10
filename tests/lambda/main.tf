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
  tags = {
    Name = "${var.name}-lambda-2"
  }
}
