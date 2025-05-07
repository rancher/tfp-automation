output "rke2_server1_public_ip" {
  value = vsphere_virtual_machine.rke2_server1.default_ip_address
}

output "rke2_server1_private_ip" {
  value = vsphere_virtual_machine.rke2_server1.default_ip_address
}

output "rke2_server2_public_ip" {
  value = vsphere_virtual_machine.rke2_server2.default_ip_address
}

output "rke2_server3_public_ip" {
  value = vsphere_virtual_machine.rke2_server3.default_ip_address
}
