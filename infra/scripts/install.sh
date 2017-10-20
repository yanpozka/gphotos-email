#!/bin/bash -xe

sudo apt update
sudo apt upgrade -y

yes | curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt install -y docker-ce

sudo systemctl status docker > ~/status_docker
sudo docker version > ~/docker_version
