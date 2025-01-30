output "rke2_bastion_public_dns" {
    value = aws_instance.rke2_bastion.public_dns
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