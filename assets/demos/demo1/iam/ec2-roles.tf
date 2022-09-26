resource "aws_iam_role" "ec2_init_role" {
  name = "cloudformation-deployer"
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

  inline_policy {
    name = "EC2-CloudFormationDeployerPolicy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "cloudformation:CreateStack",
            "iam:ListRolePolicies",
            "iam:GetRolePolicy",
            "cloudformation:DescribeStacks",
          ]
          Resource = "*"
        },
        {
          Effect = "Allow"
          Action = [
            "iam:PassRole"
          ]
          Resource = "arn:aws:iam::${local.account_id}:role/${var.name}/service-deployer"
        }
      ]
    })
  }
}

resource "aws_iam_role" "ec2_runner" {
  name = "ec2-runner"
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

  inline_policy {
    name = "EC2-RunnerPolicy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "ec2:RunInstances"
          ]
          Resource = "*"
        }
      ]
    })
  }
}

# resource "aws_iam_policy_attachment" "ssm" {
#   name       = "SSM"
#   roles      = [aws_iam_role.ec2_init_role.id]
#   policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
# }

resource "aws_iam_instance_profile" "ec2_instance_role" {
  name = "ec2-instance-role"
  role = aws_iam_role.ec2_init_role.name
}

output "instance_profile" {
  value = aws_iam_instance_profile.ec2_instance_role.name
}

resource "aws_iam_instance_profile" "ec2_runner_instance_role" {
  name = "ec2-runner-instance-role"
  role = aws_iam_role.ec2_runner.name
}
