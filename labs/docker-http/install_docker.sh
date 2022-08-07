#!/bin/bash

sudo apt-get update -y && sudo apt-get install --yes apt-transport-https ca-certificates curl gnupg lsb-release

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo \
  "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update -y
sudo apt-get install docker-ce docker-ce-cli containerd.io -y

sudo echo '{"hosts": ["tcp://0.0.0.0:2375"]}' > /etc/docker/daemon.json

sudo sed -i 's/\(ExecStart=\/usr\/bin\/dockerd\).*/\1/g' /lib/systemd/system/docker.service

sudo systemctl daemon-reload

sudo systemctl restart docker.service

docker system info