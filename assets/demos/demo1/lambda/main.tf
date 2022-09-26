data "http" "myip" {
  url = "http://ipv4.icanhazip.com"
}

resource "null_resource" "vulnerable_lambda_build" {
  triggers = {
    main_sha1 = "${sha1(file("${path.module}/../${path.module}-code/main.go"))}"
    version   = "0.1"
  }
  provisioner "local-exec" {
    command = "export GO111MODULE=on"
  }

  provisioner "local-exec" {
    working_dir = "${path.module}/../${path.module}-code/"
    command     = "go get -u; go mod tidy"
  }

  provisioner "local-exec" {
    working_dir = "${path.module}/../${path.module}-code/"
    command     = "GOTRACEBACK=system CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags '-s -w' -o server main.go"
  }
}

data "archive_file" "vulnerable_lambda_zip" {
  depends_on  = [null_resource.vulnerable_lambda_build]
  type        = "zip"
  source_file = "${path.module}/../${path.module}-code/server"
  output_path = "${path.module}/../${path.module}-code/lambdaserver.zip"
}

resource "aws_lambda_function" "vulnerable_lambda" {
  filename         = "${path.module}/../${path.module}-code/lambdaserver.zip"
  function_name    = "backend-lambda-api"
  description      = "Vulnerable SSRF Lambda function"
  role             = var.vulnerable_lambda_role
  handler          = "server"
  publish          = true
  source_code_hash = data.archive_file.vulnerable_lambda_zip.output_base64sha256
  runtime          = "go1.x"
  timeout          = "5"
  tags = {
    Name = "${var.name}-lambda"
  }

  environment {
    variables = {
      IP_ALLOWED = "${chomp(data.http.myip.response_body)}"
    }
  }
}

resource "aws_lambda_function_url" "publish_latest" {
  function_name      = aws_lambda_function.vulnerable_lambda.function_name
  authorization_type = "NONE"
}

data "archive_file" "dummy_lambda_zip" {
  type        = "zip"
  source_file = "${path.module}/../${path.module}-code/dummy.py"
  output_path = "${path.module}/../${path.module}-code/lambdadummy.zip"
}

resource "aws_lambda_function" "dummy_lambda" {
  filename         = "${path.module}/../${path.module}-code/lambdadummy.zip"
  function_name    = "backend-api-temp"
  description      = "Dummy Lambda Function"
  role             = var.dummy_lambda_role
  handler          = "dummy.handler"
  publish          = true
  source_code_hash = data.archive_file.dummy_lambda_zip.output_base64sha256
  runtime          = "python3.9"
  timeout          = "5"
  tags = {
    Name = "${var.name}-lambda-dummy"
    URL  = aws_lambda_function_url.publish_latest.function_url
  }
}


output "function_url" {
  value = aws_lambda_function_url.publish_latest.function_url
}
