# this build is being run in the native $BUILDPLATFORM platform.
# here you would do a cross-compilation.
FROM --platform=$BUILDPLATFORM golang:1.22.0-bookworm AS build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
WORKDIR /build
COPY build.sh go.* *.go ./
RUN ./build.sh

# this build is being run in $TARGETPLATFORM platform as defined in the
# buildx --platform argument.
FROM debian:bookworm-slim
COPY --from=build /build/example-docker-buildx-go /app/
# NB 65534:65534 is the uid:gid of the nobody:nogroup user:group.
# NB we use a numeric uid:gid to easy the use in kubernetes securityContext.
#    k8s will only be able to infer the runAsUser and runAsGroup values when
#    the USER intruction has a numeric uid:gid. otherwise it will fail with:
#       kubelet Error: container has runAsNonRoot and image has non-numeric
#       user (nobody), cannot verify user is non-root
USER 65534:65534
ENTRYPOINT ["/app/example-docker-buildx-go"]
