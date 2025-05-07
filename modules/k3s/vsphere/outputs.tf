output "k3s_server1_public_ip" {
  value = vsphere_virtual_machine.k3s_server1.default_ip_address
}

output "k3s_server1_private_ip" {
  value = vsphere_virtual_machine.k3s_server1.default_ip_address
}

output "k3s_server2_public_ip" {
  value = vsphere_virtual_machine.k3s_server2.default_ip_address
}

output "k3s_server3_public_ip" {
  value = vsphere_virtual_machine.k3s_server3.default_ip_address
}
