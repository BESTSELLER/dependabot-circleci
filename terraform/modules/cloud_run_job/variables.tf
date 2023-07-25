variable "project_id" {
  type = string
}

variable "name" {
  type = string
}

variable "team" {
  type = string
}

variable "env" {
  type = string
}

variable "service" {
  type = string
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
