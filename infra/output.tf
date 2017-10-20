output "application_public_ip" {
  value = "${google_compute_instance.api.ip_address}"
}