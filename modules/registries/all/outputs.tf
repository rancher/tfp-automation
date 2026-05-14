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