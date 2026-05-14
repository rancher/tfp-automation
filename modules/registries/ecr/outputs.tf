output "ecr_registry_public_dns" {
  value = aws_instance.ecr_registry.public_dns
}