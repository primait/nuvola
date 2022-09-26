variable "name" {
  description = "the name of the stack, e.g. \"demo\""
}

variable "region" {
  description = "the AWS region in which resources are created, you must set the availability_zones variable as well if you define this value to something other than the default"
  default     = "eu-west-1"
}
