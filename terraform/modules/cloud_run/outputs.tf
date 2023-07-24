output "url" {
  value = google_cloud_run_v2_service.main.status.0.url
}

