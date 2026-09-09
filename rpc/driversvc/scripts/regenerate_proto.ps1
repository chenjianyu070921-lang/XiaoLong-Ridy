$ErrorActionPreference = 'Stop'

$serviceRoot = Resolve-Path (Join-Path $PSScriptRoot '..')
Set-Location $serviceRoot

# Regenerate the protobuf and go-zero driver-service outputs from the service-local proto path.
goctl rpc protoc proto/driversvc.proto --go_out=proto --go-grpc_out=proto --zrpc_out=. --style=go_zero

function Use-ProtoAlias {
    param([Parameter(Mandatory = $true)][string]$RelativePath)

    $path = Join-Path $serviceRoot $RelativePath
    if (-not (Test-Path -LiteralPath $path)) {
        return
    }

    $content = Get-Content -Raw -LiteralPath $path
    $content = [regex]::Replace(
        $content,
        '(?m)^(\s*)(?:__proto\s+)*("XiaoLong-Ridy/rpc/driversvc/proto")',
        '$1__proto $2'
    )
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText((Resolve-Path $path), $content.TrimEnd() + [Environment]::NewLine, $utf8NoBom)
}

Use-ProtoAlias 'driverservice/driver_service.go'
Use-ProtoAlias 'internal/server/driver_service_server.go'
Use-ProtoAlias 'internal/logic/login_by_sms_logic.go'
Use-ProtoAlias 'internal/logic/register_driver_logic.go'

$obsolete = @(
    'driversvc/driversvc.go',
    'driversvcclient/driversvc.go',
    'internal/server/driversvc_server.go',
    'internal/logic/login_by_s_m_s_logic.go'
)
foreach ($relative in $obsolete) {
    $path = Join-Path $serviceRoot $relative
    if (Test-Path -LiteralPath $path) {
        Remove-Item -LiteralPath $path -Force
    }
}