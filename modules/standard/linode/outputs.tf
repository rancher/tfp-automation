output "linode_node_balancer_hostname" {
  value = linode_nodebalancer.linode_nodebalancer.hostname
}

output "server1_public_ip" {
  value = linode_instance.server1.ip_address
}

output "server1_private_ip" {
  value = linode_instance.server1.private_ip_address
}

output "server2_public_ip" {
  value = linode_instance.server2.ip_address
}

output "server3_public_ip" {
  value = linode_instance.server3.ip_address
}