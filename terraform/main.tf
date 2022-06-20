
module "bq" {
  source     = "./modules/bq"
  project_id = var.project_id
  location   = "EU"
  labels     = var.labels
}

module "controller" {
  source = "./modules/cloud_run"
  args   = ["-controller"]
  scaling = {
    max = "1"
    min = "0"
  }
  project_id = var.project_id
  location   = "europe-west4"
  labels     = var.labels
  tag        = var.tag
}

module "worker" {
  source = "./modules/cloud_run"
  args   = ["-worker"]
  scaling = {
    max = "1"
    min = "0"
  }
  project_id = var.project_id
  location   = "europe-west4"
  labels     = var.labels
  tag        = var.tag
}

module "webhook" {
  source     = "./modules/cloud_run"
  args       = ["-webhook"]
  project_id = var.project_id
  location   = "europe-west4"
  labels     = var.labels
  tag        = var.tag
}


module "schedule" {
  source     = "./modules/schedule"
  project_id = var.project_id
  url        = module.controller.url
  location   = "europe-west4"
  labels     = var.labels
}
