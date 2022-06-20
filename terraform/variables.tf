variable "project_id" {
  default = "dependabot-pub-prod-586e"
}

variable "labels" {
  default = {
    env     = "dev"
    service = "dependabot-circleci"
    team    = "engineering-services"
  }
}

variable "tag" {
  type    = string
  default = "0.0.1"
}
