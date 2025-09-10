output "rke2_server1_public_ip" {
  value = aws_instance.rke2_server1.public_ip
}

output "rke2_server1_private_ip" {
  value = aws_instance.rke2_server1.private_ip
}

output "rke2_server2_public_ip" {
  value = aws_instance.rke2_server2.public_ip
}

output "rke2_server3_public_ip" {
  value = aws_instance.rke2_server3.public_ip
}