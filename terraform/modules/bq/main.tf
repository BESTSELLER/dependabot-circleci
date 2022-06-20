resource "google_bigquery_dataset" "default" {
  project     = var.project_id
  dataset_id  = "dependabot_circleci"
  description = "Dependabot circle ci database"
  location    = var.location

  labels = var.labels
}

resource "google_bigquery_table" "bq_table" {
  project             = var.project_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  table_id            = "repos"
  deletion_protection = false

  labels = var.labels

  schema = <<EOF
[
  {
    "name": "repo",
    "type": "STRING",
    "mode": "REQUIRED"
  },
  {
    "name": "owner",
    "type": "STRING",
    "mode": "REQUIRED"

  },
  {
    "name": "schedule",
    "type": "STRING",
    "mode": "REQUIRED"
  }
]
EOF

}
