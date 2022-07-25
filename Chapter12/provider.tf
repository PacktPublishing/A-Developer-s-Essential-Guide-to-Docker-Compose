terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=2.48.0"
    }
  }
    backend "azurerm" {
        resource_group_name  = "guide-to-docker-compose-tf"
        storage_account_name = "guidetodockercomposetf"
        container_name       = "tfstate"
        key                  = "terraform.tfstate"
    }

}

provider "azurerm" {
  features {}
}
