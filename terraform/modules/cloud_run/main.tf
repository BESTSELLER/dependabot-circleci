resource "google_cloud_run_v2_service" "main" {
  provider     = google-beta
  name         = var.name
  location     = var.location
  project      = var.project_id
  launch_stage = "BETA"
  labels = {
    env     = var.env
    service = var.service
    team    = var.team
    version = replace(var.tag, ".", "_")
  }
  template {
    service_account                  = "${var.service}-v3@${var.project_id}.iam.gserviceaccount.com"
    timeout                          = "1800s"
    max_instance_request_concurrency = var.container_concurrency
    scaling {
      min_instance_count = var.scaling["min"]
      max_instance_count = var.scaling["max"]
    }
    labels = {
      env     = var.env
      service = var.service
      team    = var.team
      version = replace(var.tag, ".", "_")
    }
    containers {
      name       = var.name
      image      = "europe-docker.pkg.dev/artifacts-pub-prod-b57f/public-docker/${var.service}:${var.tag}"
      args       = var.args
      depends_on = ["secret-dumper"]
      env {
        name  = "DEPENDABOT_WORKERURL"
        value = var.worker_url
      }
      env {
        name  = "DEPENDABOT_CONFIG"
        value = "/secrets/app-secrets"
      }
      env {
        name  = "DEPENDABOT_DBCONFIG"
        value = "/secrets/db-secrets"
      }
      ports {
        name           = "http1"
        container_port = 3000
      }
      volume_mounts {
        name       = "secrets"
        mount_path = "/secrets"
      }
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
      startup_probe {
        http_get {
          path = "/healthz"
          port = 3000
        }
        initial_delay_seconds = 2
      }
    }
    containers {
      name  = "secret-dumper"
      image = "europe-docker.pkg.dev/artifacts-pub-prod-b57f/public-docker/harpocrates:2.4.0"
      args = [
        jsonencode({
          "format" : "json",
          "output" : "/secrets",
          "secrets" : [
            {
              "ES/data/${var.service}/prod" : {
                "filename" : "app-secrets"
              }
            },
            {
              "ES/data/${var.service}/db" : {
                "filename" : "db-secrets"
              }
            }
          ]
        })
      ]
      env {
        name  = "VAULT_ADDR"
        value = "https://vault.bestsellerit.com"
      }
      env {
        name  = "AUTH_NAME"
        value = "dependabot-circleci-v3"
      }
      env {
        name  = "ROLE_NAME"
        value = "dependabot-circleci-v3"
      }
      env {
        name  = "GCP_WORKLOAD_ID"
        value = "true"
      }
      env {
        name  = "CONTINUOUS"
        value = "true"
      }
      env {
        name  = "INTERVAL"
        value = "60s"
      }
      env {
        name  = "LOG_LEVEL"
        value = "warn"
      }
      volume_mounts {
        name       = "secrets"
        mount_path = "/secrets"
      }
      startup_probe {
        http_get {
          path = "/status"
          port = 8000
        }
        initial_delay_seconds = 2
      }
    }
    volumes {
      name = "secrets"
      empty_dir {
        medium = "MEMORY"
        size_limit = "1Mi"
      }
    }
    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [var.db_instance]
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

resource "google_cloud_run_v2_service_iam_member" "allow_unauthenticated" {
  count    = var.allow_unauthenticated ? 1 : 0
  location = google_cloud_run_v2_service.main.location
  project  = google_cloud_run_v2_service.main.project
  name     = google_cloud_run_v2_service.main.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
