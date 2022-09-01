resource "google_logging_project_sink" "cloud-run" {
  name                   = "cloud-run"
  project                = var.project_id
  description            = "Export logs from all Cloud Run Services to datadog"
  unique_writer_identity = true
  destination            = "pubsub.googleapis.com/projects/${var.monitor_project_id}/topics/export-logs-to-datadog"
  filter                 = "resource.type = cloud_run_revision OR resource.type = cloud_run_job severity>=DEFAULT"
}

resource "google_project_iam_member" "pubsub_publish" {
  project = var.monitor_project_id
  role    = "roles/pubsub.publisher"
  member  = google_logging_project_sink.cloud-run.writer_identity
}
