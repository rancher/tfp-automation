output "k3s_server1_public_ip" {
  value = linode_instance.k3s_server1.ip_address
}

output "k3s_server1_private_ip" {
  value = linode_instance.k3s_server1.private_ip_address
}

output "k3s_server2_public_ip" {
  value = linode_instance.k3s_server2.ip_address
}

output "k3s_server3_public_ip" {
  value = linode_instance.k3s_server3.ip_address
}