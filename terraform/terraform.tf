terraform {
  required_version = ">= 1.1.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.4.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "5.7.0"
    }
  }

  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "BESTSELLER"

    workspaces {
      name = "dependabot-circleci"
    }
  }
}

provider "google" {
}
provider "google-beta" {
}