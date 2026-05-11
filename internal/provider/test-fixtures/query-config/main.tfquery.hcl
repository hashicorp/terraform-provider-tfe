# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

list "concept_pet" "pets" {
    provider = concept
    include_resource = true
}
list "concept_pet" "animals_with_legs" {
    provider = concept
    limit = 1000
    include_resource = true
    config {
        count = var.animals
    }
}
variable "animals" {
    type = number
    default = 10
}
