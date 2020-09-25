#!/bin/bash
set -euxo pipefail

# configure.
systemctl stop docker
cat >/etc/docker/daemon.json <<'EOF'
{
    "experimental": true,
    "debug": false,
    "log-driver": "journald",
    "labels": [
        "os=linux"
    ],
    "hosts": [
        "unix://"
    ]
}
EOF

# start docker without any command line flags as its entirely configured from daemon.json.
install -d /etc/systemd/system/docker.service.d
cat >/etc/systemd/system/docker.service.d/override.conf <<'EOF'
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
EOF

# start.
systemctl daemon-reload
systemctl start docker

# show information.
docker info

# install dependencies.
apt-get install -y qemu-user-static httpie

# start a local registry.
docker run -d --restart=unless-stopped --name registry -p 5000:5000 registry:2.7.1
docker exec registry registry --version
