resource "azurerm_resource_group" "guide_to_docker_compose_resource_group" {
  name     = "guidetodockercompose"
  location = "eastus"
}

resource "azurerm_container_registry" "guide_to_docker_compose_registry" {
  name                = "developerguidetocomposeacr"
  resource_group_name = azurerm_resource_group.guide_to_docker_compose_resource_group.name
  location            = azurerm_resource_group.guide_to_docker_compose_resource_group.location
  sku                 = "Basic"
  admin_enabled       = false
}

