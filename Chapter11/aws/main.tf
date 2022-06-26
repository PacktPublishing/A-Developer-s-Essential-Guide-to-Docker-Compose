terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }
  backend "s3" {
    bucket = "developer-guide-to-compose-state"
    region = "eu-west-1"
    key = "terraform.tfstate"
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region  = "eu-west-1"
}

resource "aws_ecr_repository" "foo" {
  name                 = "developer-guide-to-compose-ecr"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

