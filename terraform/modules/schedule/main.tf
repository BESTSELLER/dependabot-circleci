resource "google_cloud_scheduler_job" "job" {
  name    = var.labels["service"]
  project = var.project_id
  region  = "europe-west4"

  schedule  = "0 05 * * *"
  time_zone = "Europe/Copenhagen"

  http_target {
    http_method = "POST"
    uri         = "${var.url}/start_controller"

    oidc_token {
      service_account_email = data.google_compute_default_service_account.default.email
      audience              = var.url
    }
  }
}
