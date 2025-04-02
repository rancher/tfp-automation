output "rke2_server1_public_ip" {
  value = linode_instance.rke2_server1.ip_address
}

output "rke2_server1_private_ip" {
  value = linode_instance.rke2_server1.private_ip_address
}

output "rke2_server2_public_ip" {
  value = linode_instance.rke2_server2.ip_address
}

output "rke2_server3_public_ip" {
  value = linode_instance.rke2_server3.ip_address
}