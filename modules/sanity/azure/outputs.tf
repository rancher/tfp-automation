output "server1_public_ip" {
  value = azurerm_public_ip.azurerm_public_ip-server1.ip_address
}

output "server1_private_ip" {
  value = azurerm_network_interface.azurerm_network_interface-server1.private_ip_address
}

output "server2_public_ip" {
  value = azurerm_public_ip.azurerm_public_ip-server2.ip_address
}

output "server3_public_ip" {
  value = azurerm_public_ip.azurerm_public_ip-server3.ip_address
}