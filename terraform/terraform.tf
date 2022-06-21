terraform {
  required_version = ">= 0.13"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.25.0"
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
  credentials = file("/tmp/cloudrun-admin.json")
}