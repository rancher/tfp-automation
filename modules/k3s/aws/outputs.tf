output "k3s_server1_public_ip" {
  value = aws_instance.k3s_server1.public_ip
}

output "k3s_server1_private_ip" {
  value = aws_instance.k3s_server1.private_ip
}

output "k3s_server2_public_ip" {
  value = aws_instance.k3s_server2.public_ip
}

output "k3s_server3_public_ip" {
  value = aws_instance.k3s_server3.public_ip
}