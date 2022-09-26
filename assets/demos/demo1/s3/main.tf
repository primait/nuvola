resource "aws_s3_bucket" "bucket" {
  bucket = "${var.name}-flag"

  tags = {
    Name = "${var.name}-flag-bucket"
  }
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.bucket.id
  key    = "flag.txt"
  source = "${path.module}/flag.txt"
  etag   = filemd5("${path.module}/flag.txt")
}
