output "auth_registry_public_dns" {
  value = aws_instance.auth_registry.public_dns
}

output "non_auth_registry_public_dns" {
  value = aws_instance.non_auth_registry.public_dns
}

output "global_registry_public_dns" {
  value = aws_instance.global_registry.public_dns
}

output "ecr_registry_public_dns" {
  value = aws_instance.ecr_registry.public_dns
}

output "server1_public_dns" {
  value = aws_instance.server1.public_dns
}

output "server1_private_ip" {
  value = aws_instance.server1.private_ip
}

output "server2_public_dns" {
  value = aws_instance.server2.public_dns
}

output "server3_public_dns" {
  value = aws_instance.server3.public_dns
}