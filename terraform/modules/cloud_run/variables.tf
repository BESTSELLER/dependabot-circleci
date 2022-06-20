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
