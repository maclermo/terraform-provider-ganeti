terraform {
  required_providers {
    ganeti = {
      source  = "github.com/maclermo/ganeti"
      version = "1.0.0"
    }
  }
}

provider "ganeti" {
  ssl_verify = false
}

resource "ganeti_instance" "super_instance" {
  name          = "terraform-test"
  vcpus         = 2
  memory        = "8G"
  hypervisor    = "kvm"
  group_name    = "default"
  disk_template = "rbd"
  os_type       = "debootstrap+buster"

  network {
    link = "br1716"
  }

  disk {
    size = "20G"
  }
}

output "instance_status" {
  description = "value"
  value       = ganeti_instance.super_instance.status
}
