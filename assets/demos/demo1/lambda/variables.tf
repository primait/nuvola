variable "name" {
  type        = string
  description = "the name of the stack, e.g. \"demo\""
}

variable "vulnerable_lambda_role" {
  type        = string
  description = "role for the vulnerable Lambda"
}

variable "dummy_lambda_role" {
  type        = string
  description = "role for the dimmy Lambda"
}
