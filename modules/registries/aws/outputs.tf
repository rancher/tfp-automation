output "auth_registry_public_dns" {
  value = aws_instance.auth.public_dns
}

output "non_auth_registry_public_dns" {
  value = aws_instance.non_auth.public_dns
}

output "auth_global_registry_public_dns" {
  value = aws_instance.auth-global.public_dns
}

output "nauth_global_registry_public_dns" {
  value = aws_instance.nauth-global.public_dns
}

output "auth_registry_route_53_fqdn" {
  value = aws_route53_record.auth.fqdn
}

output "auth_global_registry_route_53_fqdn" {
  value = aws_route53_record.auth_global.fqdn
}

output "nauth_global_registry_route_53_fqdn" {
  value = aws_route53_record.nauth_global.fqdn
}

output "ecr_registry_public_dns" {
  value = aws_instance.ecr.public_dns
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