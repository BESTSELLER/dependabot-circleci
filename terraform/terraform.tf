terraform {
  required_version = ">= 1.1.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.0.1"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "5.39.1"
    }
  }

  backend "gcs" {
    bucket = "bs-tfstate-es"
    prefix = "dependabot-circleci"
  }
}

provider "google" {
}
provider "google-beta" {
}
