variable "project_id" {
  type = string
}

variable "name" {
  type = string
}

variable "labels" {
  type = map(string)
}

variable "location" {
  type = string
}

variable "tag" {
  type = string
}

variable "args" {
  type = list(string)
}

variable "scaling" {
  default = {
    max = "10"
    min = "0"
  }
}

variable "worker_url" {
  type    = string
  default = ""
}

variable "allow_unauthenticated" {
  type    = bool
  default = false
}

variable "db_instance" {
  type    = string
  default = ""
}