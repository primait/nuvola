data "aws_caller_identity" "current" {}

resource "aws_iam_role" "datascientist" {
  path = "/${var.name}/"
  name = "Ultima-DataScientist"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          AWS = data.aws_caller_identity.current.arn
        }
      },
    ]
  })

  inline_policy {
    name = "DenyPrivs"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Deny"
          Action = [
            "cloudformation:CreateStack",
            "cloudformation:UpdateStack",
            "ec2:RunInstances",
            "lambda:Create*",
            "lambda:Update*",
            "s3:Delete*"
          ]
          Resource = "*"
        },
        {
          Effect = "Allow"
          Action = [
            "iam:Get*",
            "iam:List*",
          ]
          Resource = "*"
        }
      ]
    })
  }
}

resource "aws_iam_role_policy_attachment" "attach_data" {
  for_each = toset([
    "arn:aws:iam::aws:policy/job-function/DataScientist",
    "arn:aws:iam::aws:policy/AmazonElasticMapReduceFullAccess"
  ])

  role       = aws_iam_role.datascientist.name
  policy_arn = each.value
}

resource "aws_iam_role" "ec2_admin" {
  path = "/${var.name}/"
  name = "EC2Admin"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attach_admin" {
  role       = aws_iam_role.ec2_admin.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

resource "aws_iam_instance_profile" "privesc_instance_role" {
  name = aws_iam_role.ec2_admin.name
  role = aws_iam_role.ec2_admin.name
}
