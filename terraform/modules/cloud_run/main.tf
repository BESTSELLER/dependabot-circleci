resource "google_cloud_run_service" "main" {
  name     = var.labels["service"]
  location = var.location
  project  = var.project_id

  template {
    metadata {
      labels = var.labels
    }
    spec {
      containers {
        image = "europe-docker.pkg.dev/artifacts-pub-prod-b57f/es-docker/${var.labels["service"]}:${var.tag}"
        env {
          name  = "VAULT_ADDR"
          value = "https://vault.bestsellerit.com"
        }
        env {
          name  = "VAULT_ROLE"
          value = var.labels["service"]
        }
        env {
          name  = "VAULT_SECRET"
          value = "ES/data/${var.labels["service"]}/prod"
        }
        ports {
          name           = "http1"
          container_port = 3000
        }
      }
      service_account_name = "${var.labels["service"]}-v3@${var.project_id}.iam.gserviceaccount.com"
      timeout_seconds      = 1800
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}
