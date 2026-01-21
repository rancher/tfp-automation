output "google_load_balancer_ip_address" {
  value       = google_compute_address.google_compute_address.address
}

output "server1_public_ip" {
  value = google_compute_instance.server1.network_interface[0].access_config[0].nat_ip
}

output "server1_private_ip" {
  value = google_compute_instance.server1.network_interface[0].network_ip
}

output "server2_public_ip" {
  value = google_compute_instance.server2.network_interface[0].access_config[0].nat_ip
}

output "server3_public_ip" {
  value = google_compute_instance.server3.network_interface[0].access_config[0].nat_ip
}
