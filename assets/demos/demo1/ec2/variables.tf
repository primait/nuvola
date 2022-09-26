variable "name" {
  type        = string
  description = "the name of the stack, e.g. \"demo\""
}

variable "region" {
  type        = string
  description = "the AWS region in which resources are created, you must set the availability_zones variable as well if you define this value to something other than the default"
  default     = "eu-west-1"
}

variable "subnet_id" {
  type        = string
  description = "the network inteface to attach to the EC2"
}

variable "sg_id" {
  type        = string
  description = "the security group to add to the EC2"
}

variable "instance_profile" {
  type        = string
  description = "instance profile ARN of the role"
}
