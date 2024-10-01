terraform {
  required_version = ">= 1.1.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.1.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "6.5.0"
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
