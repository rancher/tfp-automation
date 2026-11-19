output "server1_public_ip" {
  value = vsphere_virtual_machine.server1.default_ip_address
}

output "server1_private_ip" {
  value = vsphere_virtual_machine.server1.default_ip_address
}

output "server2_public_ip" {
  value = vsphere_virtual_machine.server2.default_ip_address
}

output "server3_public_ip" {
  value = vsphere_virtual_machine.server3.default_ip_address
}
