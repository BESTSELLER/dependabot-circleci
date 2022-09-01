resource "google_cloud_run_service" "main" {
  name     = var.name
  location = var.location
  project  = var.project_id
  metadata {
    labels = {
      env     = var.env
      service = var.service
      team    = var.team
      version = replace(var.tag, ".", "_")
    }
  }
  template {
    metadata {
      labels = {
        env     = var.env
        service = var.service
        team    = var.team
        version = replace(var.tag, ".", "_")
      }
      annotations = {
        "autoscaling.knative.dev/maxScale"      = var.scaling["max"]
        "autoscaling.knative.dev/minScale"      = var.scaling["min"]
        "run.googleapis.com/cloudsql-instances" = var.db_instance

      }
    }
    spec {
      containers {
        image = "europe-docker.pkg.dev/artifacts-pub-prod-b57f/public-docker/${var.service}:${var.tag}"
        args  = var.args
        env {
          name  = "DEPENDABOT_WORKERURL"
          value = var.worker_url
        }
        env {
          name  = "VAULT_ADDR"
          value = "https://vault.bestsellerit.com"
        }
        env {
          name  = "VAULT_ROLE"
          value = "dependabot-circleci-v3"
        }
        env {
          name  = "APP_SECRET"
          value = "ES/data/${var.service}/v2"
        }
        env {
          name  = "DB_SECRET"
          value = "ES/data/${var.service}/db"
        }
        ports {
          name           = "http1"
          container_port = 3000
        }
      }
      service_account_name = "${var.service}-v3@${var.project_id}.iam.gserviceaccount.com"
      timeout_seconds      = 1800
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

resource "google_cloud_run_service_iam_member" "allow_unauthenticated" {
  count    = var.allow_unauthenticated ? 1 : 0
  location = google_cloud_run_service.main.location
  project  = google_cloud_run_service.main.project
  service  = google_cloud_run_service.main.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
