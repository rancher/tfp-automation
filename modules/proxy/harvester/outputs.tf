output "rke2_server1_public_ip" {
  value = harvester_virtualmachine.rke2_server1.network_interface[0].ip_address
}

output "rke2_server1_private_ip" {
  value = harvester_virtualmachine.rke2_server1.network_interface[0].ip_address
}

output "rke2_server2_public_ip" {
  value = harvester_virtualmachine.rke2_server2.network_interface[0].ip_address
}

output "rke2_server3_public_dns" {
  value = harvester_virtualmachine.rke2_server3.network_interface[0].ip_address
}
