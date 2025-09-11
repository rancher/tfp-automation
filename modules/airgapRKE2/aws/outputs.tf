output "registry_public_ip" {
  value = aws_instance.registry.public_ip
}

output "bastion_public_ip" {
    value = aws_instance.bastion.public_ip
}

output "server1_private_ip" {
  value = aws_instance.server1.private_ip
}

output "server2_private_ip" {
  value = aws_instance.server2.private_ip
}

output "server3_private_ip" {
  value = aws_instance.server3.private_ip
}