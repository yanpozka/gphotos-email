
variable "region" {
  default = "us-west1"
}

variable "region_zone" {
  default = "us-west1-a"
}

variable "project_name" {
  description = "The ID of the Google Cloud project"
}

variable "credentials_file_path" {
  description = "Path to the JSON file used to describe your account credentials"
  default     = "~/.gcloud/credentials.json"
}

variable "source_range" {
  description = "IPv4 range like 1.2.3.0/24"
}

variable "disk_image" {
  description = "Disk image type check here: https://cloud.google.com/compute/docs/images"
}

variable "disk_size" {
  description = "Disk size in Gb"
}

variable "machine_type" {
  description = "Machine type: https://cloud.google.com/compute/docs/machine-types"
}
