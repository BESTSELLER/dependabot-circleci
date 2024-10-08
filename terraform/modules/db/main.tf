resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "random_password" "password" {
  length = 20
}

resource "google_sql_database_instance" "main" {
  project          = var.project_id
  name             = "dependabot-circleci-${random_id.db_name_suffix.hex}"
  database_version = "POSTGRES_14"
  region           = "europe-west4"

  settings {
    tier              = "db-custom-1-3840"
    availability_type = "REGIONAL"
    backup_configuration {
      enabled = true
    }
    ip_configuration {
      ssl_mode = "TRUSTED_CLIENT_CERTIFICATE_REQUIRED"
    }
    insights_config {
      query_insights_enabled = true
    }
    user_labels = {
      env     = var.env
      service = var.service
      team    = var.team
    }
  }
}

resource "google_sql_database" "database" {
  project  = var.project_id
  name     = "repos"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "users" {
  project  = var.project_id
  name     = "dependabot-circleci"
  instance = google_sql_database_instance.main.name
  password = random_password.password.result
}

resource "vault_generic_secret" "db" {
  path = "ES/dependabot-circleci/db"

  data_json = <<EOT
{
  "connection_name": "${google_sql_database_instance.main.connection_name}",
  "db_name": "repos",
  "instance": "dependabot-circleci-${random_id.db_name_suffix.hex}",
  "password": "${random_password.password.result}",
  "username": "dependabot-circleci"
}
EOT
}
