resource "aws_ecr_repository" "developer_guide_to_compose_ecr" {
  name                 = "developer-guide-to-compose-ecr"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_cloudwatch_log_group" "task_api" {
  name              = "/ecs/task-api" 
}

resource "aws_cloudwatch_log_stream" "task_log_stream" {
  name           = "task-log-stream"
  log_group_name = aws_cloudwatch_log_group.task_api.name
}


resource "aws_vpc" "compose_vpc" {
  cidr_block = "172.17.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support = true
  
}

resource "aws_subnet" "private_subnet" {
  count = length(var.availability_zones)
  cidr_block        = cidrsubnet(aws_vpc.compose_vpc.cidr_block, 8, count.index)
  availability_zone = var.availability_zones[count.index]
  vpc_id            = aws_vpc.compose_vpc.id
}

resource "aws_subnet" "public_subnet" {
  count = length(var.availability_zones)
  cidr_block        = cidrsubnet(aws_vpc.compose_vpc.cidr_block, 8, length(var.availability_zones) + count.index)
  availability_zone = var.availability_zones[count.index]
  vpc_id            = aws_vpc.compose_vpc.id
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.compose_vpc.id
}

resource "aws_route" "internet_route" {
  route_table_id         = aws_vpc.compose_vpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
}

resource "aws_eip" "nat_ips" {
  count      = length(var.availability_zones)
  vpc        = true
  depends_on = [aws_internet_gateway.internet_gateway]
}

resource "aws_nat_gateway" "nat_gateway" {
  count         = length(var.availability_zones)
  subnet_id     = element(aws_subnet.public_subnet.*.id, count.index)
  allocation_id = element(aws_eip.nat_ips.*.id, count.index)
}

resource "aws_route_table" "private_route_table" {
  count  = length(var.availability_zones)
  vpc_id = aws_vpc.compose_vpc.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = element(aws_nat_gateway.nat_gateway.*.id, count.index)
  }
}

resource "aws_route_table_association" "private_association" {
  count          = length(var.availability_zones)
  subnet_id      = element(aws_subnet.private_subnet.*.id, count.index)
  route_table_id = element(aws_route_table.private_route_table.*.id, count.index)
}

resource "aws_security_group" "lb_security_group" {
  name        = "load-balancer-security-group"
  vpc_id      = aws_vpc.compose_vpc.id

  ingress {
    protocol    = "tcp"
    from_port   = "80"
    to_port     = "80"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "compose_security_group" {
  name        = "compose-security-group"
  vpc_id      = aws_vpc.compose_vpc.id

  ingress {
    protocol        = "tcp"
    from_port       = "80"
    to_port         = "80"
    security_groups = [aws_security_group.lb_security_group.id]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group_rule" "allow_services_connectivity" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = aws_security_group.compose_security_group.id
  security_group_id        = aws_security_group.compose_security_group.id
}

resource "aws_alb" "compose_alb" {
  name            = "guide-to-compose-load-balancer"
  subnets         = aws_subnet.public_subnet.*.id
  security_groups = [aws_security_group.lb_security_group.id]
}


resource "aws_ecs_cluster" "compose_ecs" {
  name = "guide-to-compose-ecs"
}