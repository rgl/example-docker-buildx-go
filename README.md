# About

[![Build status](https://img.shields.io/github/workflow/status/rgl/example-docker-buildx-go/Build)](https://github.com/rgl/example-docker-buildx-go/actions?query=workflow%3ABuild)
[![Docker pulls](https://img.shields.io/docker/pulls/ruilopes/example-docker-buildx-go)](https://hub.docker.com/repository/docker/ruilopes/example-docker-buildx-go)

This is an example on how to use docker buildx to build and publish a
multi-platform container image of a go application.

This uses qemu-user-static to run the non-native platform binaries in emulation
mode, i.e., docker buildx uses qemu to run `arm` binaries in a `amd64` host.

With this we can build container images that can be used in Raspberry Pi or
other ARM based architectures.

# Usage (Ubuntu 20.04)

Install skopeo and docker:

```bash
# run all the commands as root.
sudo -i

# install skopeo.
# see https://github.com/containers/skopeo/blob/master/install.md
LSB_ID="$(lsb_release --id -s)"
LSB_RELEASE="$(lsb_release --release -s)"
echo "deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/x${LSB_ID}_${LSB_RELEASE}/ /" \
    >/etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list
wget -qO- "https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/x${LSB_ID}_${LSB_RELEASE}/Release.key" \
    | apt-key add -
apt-get update
apt-get install -y skopeo
install -d -m 755 /etc/containers/registries.conf.d
cat >/etc/containers/registries.conf.d/localhost-5000.conf <<'EOF'
[[registry]]
location = 'localhost:5000'
insecure = true
EOF

# install docker.
# see https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/#install-using-the-repository
docker_version='20.10.8'
apt-get install -y apt-transport-https software-properties-common
wget -qO- https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update
docker_apt_version="$(apt-cache madison docker-ce | awk "/$docker_version~/{print \$3}")"
apt-get install -y "docker-ce=$docker_apt_version" "docker-ce-cli=$docker_apt_version" containerd.io

# configure docker.
systemctl stop docker
cat >/etc/docker/daemon.json <<'EOF'
{
    "experimental": false,
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
systemctl daemon-reload
systemctl start docker

# exit the root shell.
exit
```

**NB** The docker experimental mode is needed to be able to run non-native
platform containers (in emulated mode).

Install dependencies:

```bash
sudo apt-get install qemu-user-static httpie
```

Create a local buildx builder:

```bash
docker buildx create \
    --name local \
    --driver docker-container \
    --driver-opt network=host \
    --use
docker buildx ls
```

Start an ephemeral local registry to be the target of our buildx build:

```bash
docker run -d --restart=unless-stopped --name registry -p 5000:5000 registry:2.7.1
docker exec registry registry --version
```

Build for multiple platforms:

```bash
docker buildx build \
    --tag localhost:5000/example-docker-buildx-go \
    --output type=registry \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    --progress plain \
    .
```

**NB** multiple platforms images [cannot be exported to local docker](https://github.com/docker/buildx#docker)
that's why we are using a local registry (and `--driver-opt network=host` when
we create the `local` builder).

List the available repositories:

```bash
http get http://localhost:5000/v2/_catalog
```

Should return something alike:

```
HTTP/1.1 200 OK
Content-Length: 66
Content-Type: application/json; charset=utf-8
Date: Fri, 25 Sep 2020 07:33:06 GMT
Docker-Distribution-Api-Version: registry/2.0
X-Content-Type-Options: nosniff

{
    "repositories": [
        "example-docker-buildx-go"
    ]
}
```

List the tags:

```bash
http get http://localhost:5000/v2/example-docker-buildx-go/tags/list
```

Should return something alike:

```
HTTP/1.1 200 OK
Content-Length: 54
Content-Type: application/json; charset=utf-8
Date: Fri, 25 Sep 2020 07:34:06 GMT
Docker-Distribution-Api-Version: registry/2.0
X-Content-Type-Options: nosniff

{
    "name": "example-docker-buildx-go",
    "tags": [
        "latest"
    ]
}
```

List the fat manifest:

```bash
http get \
    http://localhost:5000/v2/example-docker-buildx-go/manifests/latest \
    Accept:application/vnd.docker.distribution.manifest.list.v2+json
```

**NB** You can also use `skopeo inspect --raw docker://localhost:5000/example-docker-buildx-go`.

Should return something alike:

```
HTTP/1.1 200 OK
Content-Length: 1076
Content-Type: application/vnd.docker.distribution.manifest.list.v2+json
Date: Fri, 25 Sep 2020 07:34:10 GMT
Docker-Content-Digest: sha256:5b81907eb34e2fd9a197fb55609362a02727c2b1f5b70b1a1033444e1e425983
Docker-Distribution-Api-Version: registry/2.0
Etag: "sha256:5b81907eb34e2fd9a197fb55609362a02727c2b1f5b70b1a1033444e1e425983"
X-Content-Type-Options: nosniff

{
    "manifests": [
        {
            "digest": "sha256:5885438d35170aaa8f500ef90173467c9251a147f77a1e80f24be6c30808ce38",
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "platform": {
                "architecture": "amd64",
                "os": "linux"
            },
            "size": 739
        },
        {
            "digest": "sha256:faeb5602dd3a90623edf91411dddabec315ef1e144c715850bef0ef5113883e7",
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "platform": {
                "architecture": "arm64",
                "os": "linux"
            },
            "size": 739
        },
        {
            "digest": "sha256:b4806787851eb651c6017fdd31ff547d425894a3c42a31fddadd078b0ee0547e",
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "platform": {
                "architecture": "arm",
                "os": "linux",
                "variant": "v7"
            },
            "size": 739
        }
    ],
    "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
    "schemaVersion": 2
}
```

Run the example application container as a `linux/amd64` native platform:

```bash
docker rmi -f localhost:5000/example-docker-buildx-go
docker run --rm -t localhost:5000/example-docker-buildx-go
```

You should something alike:

```
go1.17.1
TARGETPLATFORM=linux/amd64
GOOS=linux
GOARCH=amd64
```

Run the example application container as a `linux/arm64` emulated platform:

```bash
docker rmi -f localhost:5000/example-docker-buildx-go
docker run --platform linux/arm64 --rm -t localhost:5000/example-docker-buildx-go
```

You should something alike:

```
go1.17.1
TARGETPLATFORM=linux/arm64
GOOS=linux
GOARCH=arm64
```

Run the example application container as a `linux/arm/v7` emulated platform:

```bash
docker rmi -f localhost:5000/example-docker-buildx-go
docker run --platform linux/arm/v7 --rm -t localhost:5000/example-docker-buildx-go
```

You should something alike:

```
go1.17.1
TARGETPLATFORM=linux/arm/v7
GOOS=linux
GOARCH=arm
```

Publish the multi-platform images to Docker Hub:

```bash
DOCKER_HUB_USER='YOUR-DOCKER-HUB-ACCOUNT-USER'
DOCKER_HUB_ACCESS_TOKEN='CREATE-THIS-FROM-YOUR-DOCKER-HUB-ACCOUNT-SECURITY-PAGE'
skopeo copy --all \
    --dest-creds "$DOCKER_HUB_USER:$DOCKER_HUB_ACCESS_TOKEN" \
    docker://localhost:5000/example-docker-buildx-go \
    docker://docker.io/$DOCKER_HUB_USER/example-docker-buildx-go:latest
```

Use the image:

```bash
docker run --rm -t $DOCKER_HUB_USER/example-docker-buildx-go
```

# Reference

* https://github.com/docker/buildx
* https://www.docker.com/blog/multi-platform-docker-builds/
* https://docs.docker.com/engine/reference/commandline/buildx/
* https://docs.docker.com/registry/spec/api/
* https://docs.docker.com/registry/spec/manifest-v2-2/
