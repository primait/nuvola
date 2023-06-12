variable "name" {
  type        = string
  description = "the name of the stack, e.g. \"test\""
}

variable "lambda_1_role" {
  type        = string
  description = "role for the vulnerable Lambda"
}

variable "lambda_2_role" {
  type        = string
  description = "role for the dummy Lambda"
}
