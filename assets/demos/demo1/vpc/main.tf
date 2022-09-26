data "http" "myip" {
  url = "http://ipv4.icanhazip.com"
}

resource "aws_vpc" "vpc" {
  cidr_block           = "171.254.0.0/28"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    Name = "${var.name}-vpc"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.vpc.id


  tags = {
    Name = "${var.name}-igtw"
  }
}

resource "aws_subnet" "public-subnet" {
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = "171.254.0.0/28"
  map_public_ip_on_launch = true
  tags = {
    Name = "${var.name}-public-subnet-1"
  }
}

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = {
    Name = "${var.name}-route-table-1"
  }
}
#route table connect to public subnet
resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.public-subnet.id
  route_table_id = aws_route_table.r.id
}


resource "aws_default_security_group" "default" {
  # name = "${var.name}-default"
  # description = "Allow inbound traffic from myip"
  vpc_id = aws_vpc.vpc.id

  ingress {
    description = "All from myip"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["${chomp(data.http.myip.response_body)}/32"]
  }

  egress {
    description = "All to Internet"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name}-allow-myip"
  }
}

output "subnet_id" {
  value = aws_subnet.public-subnet.id
}

output "sg_id" {
  value = aws_default_security_group.default.id
}
