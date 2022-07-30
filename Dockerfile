# this build is being run in the native $BUILDPLATFORM platform.
# here you would do a cross-compilation.
FROM --platform=$BUILDPLATFORM golang:1.18.4-bullseye AS build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
WORKDIR /build
COPY build.sh go.* *.go ./
RUN ./build.sh

# this build is being run in $TARGETPLATFORM platform as defined in the
# buildx --platform argument.
FROM debian:bullseye-slim
COPY --from=build /build/example-docker-buildx-go /app/
USER nobody:nogroup
ENTRYPOINT ["/app/example-docker-buildx-go"]
