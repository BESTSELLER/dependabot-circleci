module "db" {
  source     = "./modules/db"
  project_id = var.project_id
  service    = var.service
  env        = var.env
  team       = var.team
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
  project_id  = var.project_id
  location    = "europe-west4"
  service     = var.service
  env         = var.env
  team        = var.team
  tag         = var.tag
  db_instance = module.db.db_instance
}

module "worker" {
  name   = "worker"
  source = "./modules/cloud_run"
  args   = ["-worker"]

  project_id            = var.project_id
  location              = "europe-west4"
  service               = var.service
  env                   = var.env
  team                  = var.team
  tag                   = var.tag
  db_instance           = module.db.db_instance
  container_concurrency = 1
}

module "webhook" {
  name                  = "webhook"
  source                = "./modules/cloud_run"
  worker_url            = module.worker.url
  args                  = ["-webhook"]
  project_id            = var.project_id
  location              = "europe-west4"
  service               = var.service
  env                   = var.env
  team                  = var.team
  tag                   = var.tag
  allow_unauthenticated = true
  db_instance           = module.db.db_instance
  custom_domain         = true
}


module "schedule" {
  source     = "./modules/schedule"
  project_id = var.project_id
  url        = module.controller.url
  location   = "europe-west4"
  service    = var.service
}

module "logs" {
  source             = "./modules/logs"
  project_id         = var.project_id
  monitor_project_id = var.monitor_project_id
}
