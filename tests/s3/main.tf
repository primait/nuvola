resource "aws_s3_bucket" "test-bucket" {
  bucket = "nuvola-s3-test"
  tags = {
    Name        = "nuvola-s3-test"
    Environment = "dev"
    Stack       = var.name
  }
}

resource "aws_s3_bucket_policy" "allow_all" {
  bucket = aws_s3_bucket.test-bucket.id
  policy = jsonencode({
    Version : "2012-10-17",
    Statement : [
      {
        Effect : "Allow",
        Principal : "*",
        Action : ["s3:*"],
        Resource : [
          "${aws_s3_bucket.test-bucket.arn}",
          "${aws_s3_bucket.test-bucket.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_s3_bucket" "lambda-output" {
  bucket = "lambda-output"
}

resource "aws_kms_key" "s3-key" {
  description             = "This key is used to encrypt bucket objects"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test-bucket-encryption" {
  bucket = aws_s3_bucket.test-bucket.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3-key.arn
      sse_algorithm     = "aws:kms"
    }
  }
}
