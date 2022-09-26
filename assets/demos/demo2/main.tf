module "iam" {
  source = "./iam"
  name   = var.name
}
module "s3" {
  source = "./s3"
  name   = var.name
}
