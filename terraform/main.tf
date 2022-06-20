
module "bq" {
  source     = "./modules/bq"
  project_id = var.project_id
  location   = "EU"
  labels     = var.labels
}

module "controller" {
  name       = "controller"
  source     = "./modules/cloud_run"
  args       = ["-controller"]
  worker_url = module.worker.url
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
  name   = "worker"
  source = "./modules/cloud_run"
  args   = ["-worker"]

  project_id = var.project_id
  location   = "europe-west4"
  labels     = var.labels
  tag        = var.tag
}

module "webhook" {
  name       = "webhook"
  source     = "./modules/cloud_run"
  worker_url = module.worker.url
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
