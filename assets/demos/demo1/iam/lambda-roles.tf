data "aws_caller_identity" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_iam_role" "lambda_runner_role" {
  path = "/${var.name}/"
  name = "lambda-runner"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = "LambdaRunnerPolicy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "logs:CreateLogStream",
            "logs:PutLogEvents",
            "iam:ListAttachedRolePolicies",
            "iam:ListRolePolicies",
            "iam:GetRolePolicy",
          ]
          Resource = "*"
        },
        {
          Effect = "Allow"
          Action = [
            "lambda:InvokeFunction",
            "lambda:ListFunctions",
          ]
          Resource = "*",
        }
      ]
    })
  }
}

resource "aws_iam_role" "temp_backend_api_role_runner" {
  name = "temp-backend-api-role-runner"
  path = "/${var.name}/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attach_admin" {
  role       = aws_iam_role.temp_backend_api_role_runner.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

output "lambda_runner_role" {
  value = aws_iam_role.lambda_runner_role.arn
}

output "temp_backend_api_role" {
  value = aws_iam_role.temp_backend_api_role_runner.arn
}
