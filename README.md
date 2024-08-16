# About

[![Build status](https://img.shields.io/github/actions/workflow/status/rgl/example-docker-buildx-go/build.yml)](https://github.com/rgl/example-docker-buildx-go/actions/workflows/build.yml)
[![Docker pulls](https://img.shields.io/docker/pulls/ruilopes/example-docker-buildx-go)](https://hub.docker.com/repository/docker/ruilopes/example-docker-buildx-go)

This is an example on how to use docker buildx to build and publish a
multi-platform container image of a go application.

This uses qemu-user-static to run the non-native platform binaries in emulation
mode, i.e., docker buildx uses qemu to run `arm` binaries in a `amd64` host.

With this we can build container images that can be used in Raspberry Pi or
other ARM based architectures.

# Use (Docker)

You can use the multi-platform container image as:

```bash
docker run --rm ruilopes/example-docker-buildx-go:v1.11.0
```

# Use (Kubernetes)

You can use the multi-platform container image to launch an example web service `DaemonSet` as:

```bash
kubectl apply -f - <<'EOF'
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: example-app
spec:
  selector:
    matchLabels:
      app: example-app
  template:
    metadata:
      labels:
        app: example-app
    spec:
      enableServiceLinks: false
      tolerations:
        - effect: NoSchedule
          key: node.kubernetes.io/os
          operator: Equal
          value: windows
      containers:
        - name: example-app
          image: ruilopes/example-docker-buildx-go:v1.11.0
          args:
            - -listen=:8000
          ports:
            - name: web
              containerPort: 8000
EOF
```

See the complete example at https://github.com/rgl/talos-vagrant/blob/main/provision-example-daemonset.sh.

# Build (Ubuntu 22.04)

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
docker_version='25.0.3'
apt-get install -y apt-transport-https software-properties-common
wget -qO- https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/download.docker.com.gpg
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/download.docker.com.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" >/etc/apt/sources.list.d/docker.list
apt-get update
docker_package_version="$(apt-cache madison docker-ce | awk "/$docker_version/{print \$3}")"
apt-get install -y "docker-ce=$docker_package_version" "docker-ce-cli=$docker_package_version" containerd.io

# configure docker.
systemctl stop docker
install -m 750 -d /etc/docker
cat >/etc/docker/daemon.json <<'EOF'
{
    "experimental": false,
    "debug": false,
    "features": {
        "buildkit": true
    },
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
docker run -d --restart=unless-stopped --name registry -p 5000:5000 registry:2.8.3
docker exec registry registry --version
```

Build for multiple platforms:

```bash
export BUILDX_NO_DEFAULT_ATTESTATIONS=1
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
Date: Wed, 21 Feb 2024 07:52:19 GMT
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
Date: Wed, 21 Feb 2024 07:52:19 GMT
Docker-Distribution-Api-Version: registry/2.0
X-Content-Type-Options: nosniff

{
    "name": "example-docker-buildx-go",
    "tags": [
        "latest"
    ]
}
```

Show the image index manifest:

```bash
http get \
    http://localhost:5000/v2/example-docker-buildx-go/manifests/latest \
    Accept:application/vnd.docker.distribution.manifest.list.v2+json
```

**NB** You can also use `skopeo inspect --raw docker://localhost:5000/example-docker-buildx-go`.

**NB** You can also use `regctl manifest get --format raw-body localhost:5000/example-docker-buildx-go`.

Should return something alike:

```
HTTP/1.1 200 OK
Content-Length: 987
Content-Type: application/vnd.docker.distribution.manifest.list.v2+json
Date: Wed, 21 Feb 2024 07:52:19 GMT
Docker-Content-Digest: sha256:0ee26ecbf446a5b38155d2d512f822174120e33bb56bc09425bcd8952c0060b7
Docker-Distribution-Api-Version: registry/2.0
Etag: "sha256:0ee26ecbf446a5b38155d2d512f822174120e33bb56bc09425bcd8952c0060b7"
X-Content-Type-Options: nosniff

{
    "manifests": [
        {
            "digest": "sha256:ea5d80666d6a26690f2a3621a42eb01898db1351db5a6d5854b2f5300cba8b38",
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "platform": {
                "architecture": "amd64",
                "os": "linux"
            },
            "size": 702
        },
        {
            "digest": "sha256:b8fdc4bd2c9c20f3b7c6a22804fc017422969b7ae693dbc1134fefa27456697f",
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "platform": {
                "architecture": "arm64",
                "os": "linux"
            },
            "size": 702
        },
        {
            "digest": "sha256:ee164a134bec30bfcdf94a51c714f9b2b058e183c26cef82923e7bb1f93b286d",
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "platform": {
                "architecture": "arm",
                "os": "linux",
                "variant": "v7"
            },
            "size": 702
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
go1.23.0
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
go1.23.0
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
go1.23.0
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

`docker buildx` is not supported on Windows, as such, you have to build the
Windows images separately, then create a manifest list with all of the
different images. See how in the [`build workflow`](.github/workflows/build.yml).

# Reference

* https://github.com/docker/buildx
* https://www.docker.com/blog/multi-platform-docker-builds/
* https://docs.docker.com/engine/reference/commandline/buildx/
* https://docs.docker.com/registry/
* https://distribution.github.io/distribution/spec/manifest-v2-2/
