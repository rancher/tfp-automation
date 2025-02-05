output "k3s_server1_public_dns" {
  value = aws_instance.k3s_server1.public_dns
}

output "k3s_server1_private_ip" {
  value = aws_instance.k3s_server1.private_ip
}

output "k3s_server2_public_dns" {
  value = aws_instance.k3s_server2.public_dns
}

output "k3s_server3_public_dns" {
  value = aws_instance.k3s_server3.public_dns
}