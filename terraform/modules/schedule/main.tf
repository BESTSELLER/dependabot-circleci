resource "google_project_service" "cloudscheduler" {
  project            = var.project_id
  disable_on_destroy = false

  service = "cloudscheduler.googleapis.com"
}

resource "google_cloud_scheduler_job" "job" {
  name       = var.labels["service"]
  project    = var.project_id
  region     = "europe-west4"
  depends_on = [google_project_services.cloudscheduler]

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
