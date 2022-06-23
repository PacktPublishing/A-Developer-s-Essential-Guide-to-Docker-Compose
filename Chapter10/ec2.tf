resource "aws_security_group" "remote_docker_host_security_group" {
  name        = "remote-docker-host-security-group"
  description = "Allow SSH to workstation IP"
  vpc_id      = var.myvpc

  ingress {
    description      = "SSH from workstation"
    from_port        = 22   
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = ["${var.myip}/32"]
  }
  egress = [
    {
      cidr_blocks      = [ "0.0.0.0/0", ]
      description      = ""
      from_port        = 0
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      protocol         = "-1"
      security_groups  = []
      self             = false
      to_port          = 0
    }
  ]

  tags = {
    app = "remote-docker-host"
  }
}

resource "aws_key_pair" "docker_remote_host_key" {
  key_name   = "docker-remote-host-key"
  public_key = file("${path.module}/ssh.key.pub")

 tags = {
    app = "remote-docker-host"
  }
}

resource "aws_instance" "remote_docker_host" {
    ami = "ami-078a289ddf4b09ae0"
    instance_type = "t2.micro"

    key_name = aws_key_pair.docker_remote_host_key.key_name


    vpc_security_group_ids = [ 
        aws_security_group.remote_docker_host_security_group.id
     ]

  user_data = <<-EOF
    #!/bin/bash
    yum install docker -y
    usermod -aG docker ec2-user
    systemctl start docker
    su ec2-user
    mkdir -p /home/ec2-user/.docker/cli-plugins
    curl -SL https://github.com/docker/compose/releases/download/v2.2.3/docker-compose-linux-x86_64 -o /home/ec2-user/.docker/cli-plugins/docker-compose
    chmod +x /home/ec2-user/.docker/cli-plugins/docker-compose
  EOF


  tags = {
    app = "remote-docker-host"
  }
}

output "instance_ip" {
  description = "Remote host ip"
  value       = aws_instance.remote_docker_host.public_ip
}