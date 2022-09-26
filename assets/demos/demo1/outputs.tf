output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = module.ec2.instance_public_ip
}

output "private_key" {
  description = "SSH private key in PEM format"
  value       = module.ec2.private_key
  sensitive   = true
}
