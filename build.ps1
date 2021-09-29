Set-StrictMode -Version Latest
$ProgressPreference = 'SilentlyContinue'
$ErrorActionPreference = 'Stop'

$TARGETPLATFORM = "$(go env GOOS)/$(go env GOARCH)"
$env:CGO_ENABLED = '0'

go build -ldflags="-s -X main.TARGETPLATFORM=$TARGETPLATFORM"
