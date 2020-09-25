#!/bin/bash
set -euxo pipefail

if [ "$BUILDPLATFORM" != "$TARGETPLATFORM" ]; then
    echo "Cross-compiling to $TARGETPLATFORM"
    # $TARGETPLATFORM is something like:
    #   linux/amd64
    #   linux/arm64
    #   linux/arm/v7
    target_platform=(${TARGETPLATFORM//\// })
    export GOOS=${target_platform[0]}
    export GOARCH=${target_platform[1]}
    if [ "${#target_platform[@]}" -gt 2 ]; then
        export GOARM=${target_platform[2]//v}
    fi
else
    echo "Compiling to $TARGETPLATFORM"
fi

CGO_ENABLED=0 go build -ldflags="-s"
