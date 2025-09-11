output "server1_public_ip" {
  value = aws_instance.server1.public_ip
}

output "server1_private_ip" {
  value = aws_instance.server1.private_ip
}

output "server2_public_ip" {
  value = aws_instance.server2.public_ip
}

output "server3_public_ip" {
  value = aws_instance.server3.public_ip
}