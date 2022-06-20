module "bq" {
  source     = "./modules/bq"
  project_id = var.project_id
  location   = "EU"
  labels     = var.labels
}

module "cloud_run" {
  source     = "./modules/cloud_run"
  project_id = var.project_id
  location   = "europe-west4"
  labels     = var.labels
  tag        = var.tag
}

module "schedule" {
  source     = "./modules/schedule"
  project_id = var.project_id
  url        = module.cloud_run.url
  location   = "europe-west4"
  labels     = var.labels
}
