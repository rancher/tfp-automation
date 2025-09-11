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

output "server1_public_ip" {
  value = aws_instance.server1.ipv6_addresses[0]
}

output "server2_public_ip" {
  value = aws_instance.server2.ipv6_addresses[0]
}

output "server3_public_ip" {
  value = aws_instance.server3.ipv6_addresses[0]
}