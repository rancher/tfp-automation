output "server1_public_ip" {
  value = google_compute_instance.server1.network_interface[0].access_config[0].nat_ip
}