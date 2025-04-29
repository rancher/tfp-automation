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

output "rke2_server1_public_dns" {
  value = aws_instance.rke2_server1.public_dns
}

output "rke2_server1_private_ip" {
  value = aws_instance.rke2_server1.private_ip
}

output "rke2_server2_public_dns" {
  value = aws_instance.rke2_server2.public_dns
}

output "rke2_server3_public_dns" {
  value = aws_instance.rke2_server3.public_dns
}