resource "google_cloud_run_v2_job" "main" {
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
    labels = {
      env     = var.env
      service = var.service
      team    = var.team
      version = replace(var.tag, ".", "_")
    }
    parallelism = 1
    template {
      service_account = "${var.service}-v3@${var.project_id}.iam.gserviceaccount.com"
      timeout         = "3600s"
      max_retries     = 0
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
        volume_mounts {
          name       = "secrets"
          mount_path = "/secrets"
        }
        volume_mounts {
          name       = "cloudsql"
          mount_path = "/cloudsql"
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
        empty_dir {}
      }
      volumes {
        name = "cloudsql"
        cloud_sql_instance {
          instances = [var.db_instance]
        }
      }
    }
  }
}

resource "google_project_service" "cloudscheduler" {
  project            = var.project_id
  disable_on_destroy = false

  service = "cloudscheduler.googleapis.com"
}

resource "google_cloud_scheduler_job" "job" {
  name    = var.name
  project = var.project_id
  region  = "us-central1"

  schedule  = "0 05 * * *"
  time_zone = "Europe/Copenhagen"

  http_target {
    http_method = "POST"
    uri         = "https://${google_cloud_run_v2_job.main.location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${data.google_project.project.number}/jobs/${google_cloud_run_v2_job.main.name}:run"

    oauth_token {
      service_account_email = "${var.service}-v3@${var.project_id}.iam.gserviceaccount.com"
    }
  }
}
