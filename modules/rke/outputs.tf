output "rke_server1_public_ip" {
  value = aws_instance.rke_server1.public_ip
}

output "rke_server2_public_ip" {
  value = aws_instance.rke_server2.public_ip
}

output "rke_server3_public_ip" {
  value = aws_instance.rke_server3.public_ip
}