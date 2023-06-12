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

resource "aws_iam_role" "lambda_admin_role" {
  name = "lambda-admin-role"
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

  inline_policy {
    name = "NotActionsAllow"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          NotAction = [
            "iam:*",
            "ec2:Create*",
            "ec2:Delete*",
            "ec2:TerminateInstances",
            "ec2:Cancel*",
            "ec2:Auth*",
          ]
          Resource = "*"
        }
      ]
    })
  }
}

resource "aws_iam_role_policy_attachment" "attach_admin" {
  role       = aws_iam_role.lambda_admin_role.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

output "lambda_runner_role" {
  value = aws_iam_role.lambda_runner_role
}

output "lambda_admin_role" {
  value = aws_iam_role.lambda_admin_role
}
