resource "aws_cloudformation_stack" "dummy_stack" {
  name = "stack-test1"

  iam_role_arn = var.iam_role

  template_body = <<STACK
    Resources:
        NullResource:
            Type: AWS::CloudFormation::WaitConditionHandle
    STACK
}
