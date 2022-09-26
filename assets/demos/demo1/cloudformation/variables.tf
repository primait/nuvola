variable "name" {
  type        = string
  description = "the name of the stack, e.g. \"demo\""
}

variable "iam_role" {
  type        = string
  description = "ARN of the role to Pass to the stack"
}
