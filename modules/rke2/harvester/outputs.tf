output "rke2_server1__public_ip" {
  value = harvester_virtualmachine.rke2_server1.network_interface[0].ip_address
}

output "rke2_server1_private_ip" {
  value = harvester_virtualmachine.rke2_server1.network_interface[0].ip_address
}

output "rke2_server2__public_ip" {
  value = harvester_virtualmachine.rke2_server2.network_interface[0].ip_address
}

output "rke2_server3__public_ip" {
  value = harvester_virtualmachine.rke2_server3.network_interface[0].ip_address
}
