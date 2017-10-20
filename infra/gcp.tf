provider "google" {
  region      = "${var.region}"
  project     = "${var.project_name}"
  credentials = "${file("${var.credentials_file_path}")}"
}

resource "google_compute_disk" "apidisk" {
  name  = "apidisk"
  type  = "pd-ssd"
  zone  = "${var.region_zone}"
  image = "${var.disk_image}"
  size  = "${var.disk_size}"

  labels {
    environment = "dev"
  }
}

resource "google_compute_instance" "api" {
  name         = "tf-api-compute"
  machine_type = "${var.machine_type}"
  zone         = "${google_compute_disk.apidisk.zone}"
  tags         = ["http"]

  boot_disk {
    source      = "${google_compute_disk.apidisk.name}"
    auto_delete = true
  }

  network_interface {
    network       = "default"
    access_config = {}
  }

  metadata_startup_script = "${file("scripts/install.sh")}"

  service_account {
    scopes = ["https://www.googleapis.com/auth/compute.readonly"]
  }
}
