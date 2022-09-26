data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["amzn2-ami-hvm*"]
  }
}

resource "tls_private_key" "sshkey" {
  algorithm = "RSA"
  rsa_bits  = "4096"
}

# The private key is stored inside the state file!
resource "aws_key_pair" "generated_key" {
  key_name   = "${var.name}-key"
  public_key = tls_private_key.sshkey.public_key_openssh
}

locals {
  cloud_config_config = <<-END
    #cloud-config
    ${jsonencode({
  write_files = [
    {
      path        = "/tmp/main.go"
      permissions = "0644"
      encoding    = "b64"
      content     = filebase64("${path.module}/../${path.module}-code/main.go")
    },
  ]
})}
  END
}

data "cloudinit_config" "server_demo1" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/cloud-config"
    filename     = "cloud-config.yaml"
    content      = local.cloud_config_config
  }

  part {
    content_type = "text/x-shellscript"
    filename     = "startup.sh"
    content      = <<-EOF
      #!/bin/bash
      sudo yum install -y golang
      sudo crontab -l > /tmp/new_cron
      echo "* * * * * go run /tmp/main.go" >> /tmp/new_cron
      sudo crontab /tmp/new_cron
    EOF
  }
}

resource "aws_instance" "demo1" {
  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "t2.micro"
  subnet_id                   = var.subnet_id
  associate_public_ip_address = true
  iam_instance_profile        = var.instance_profile
  key_name                    = aws_key_pair.generated_key.key_name

  # Enable IMDSv1
  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "optional"
  }

  user_data = data.cloudinit_config.server_demo1.rendered

  tags = {
    Name = "${var.name}-ec2"
  }
}

output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.demo1.public_ip
}

output "private_key" {
  value     = tls_private_key.sshkey.private_key_pem
  sensitive = true
}
