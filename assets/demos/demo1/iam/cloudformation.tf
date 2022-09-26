resource "aws_iam_role" "service_deployer" {
  path = "/${var.name}/"
  name = "service-deployer"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = "Ultima-CustomDeployerPolicy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "lambda:*",
          ]
          Resource = "*",
        },
        {
          Effect = "Allow"
          Action = [
            "ec2:*",
          ]
          Resource = "*",
        },
        {
          Effect = "Allow"
          Action = [
            "iam:PassRole",
          ]
          Resource = "arn:aws:iam::${local.account_id}:role/${var.name}/*-runner", # bruteforce on the name: lambda
        }
      ]
    })
  }
}

output "service_deployer_arn" {
  value = aws_iam_role.service_deployer.arn
}
