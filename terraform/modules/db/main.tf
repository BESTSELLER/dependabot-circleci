resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "random_password" "password" {
  length = 20
}

resource "google_sql_database_instance" "main" {
  name             = "dependabot_circleci-${random_id.db_name_suffix.hex}"
  database_version = "POSTGRES_14"
  region           = "europe-west4"

  settings {
    tier              = "db-custom-1-3840"
    availability_type = "REGIONAL"
    backup_configuration {
      enabled = true
    }
    insights_config {
      query_insights_enabled = true
    }
    user_labels = var.labels
  }
}

resource "google_sql_database" "database" {
  name     = "repos"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "users" {
  name     = "dependabot_circleci"
  instance = google_sql_database_instance.main.name
  password = random_password.password.result
}

resource "vault_generic_secret" "sql_password" {
  path = "ES/dependabot-circleci/db"

  data_json = <<EOT
{
  "password": "${random_password.password.result}"
}
EOT
}
