# Development bootstrap for driver, passenger, admin, and related backend services.
# Usage:
#   powershell -ExecutionPolicy Bypass -File scripts/start_all_dev.ps1

$ErrorActionPreference = "Continue"
$root = "D:\gocode\src\XiaoLong-Ridy"

# Keep cache on a dedicated path to avoid temp disk issues.
$env:GOCACHE = "D:\gocode\gocache"
$env:GOFLAGS = "-mod=mod"

$logDir = Join-Path $root "scripts\logs"
if (-not (Test-Path $logDir)) { New-Item -ItemType Directory -Path $logDir | Out-Null }

function Start-Svc {
    param(
        [string]$Name,
        [string]$Dir,
        [string]$Cmd,
        [string]$LogFile
    )
    $logPath = Join-Path $logDir $LogFile
    Write-Host "[START] $Name => $Dir : $Cmd"
    $ps = New-Object System.Diagnostics.ProcessStartInfo
    $ps.FileName = "powershell.exe"
    $ps.Arguments = "-NoProfile -ExecutionPolicy Bypass -Command `"Set-Location '$Dir'; $Cmd 2>&1 | Tee-Object -FilePath '$logPath'`""
    $ps.WorkingDirectory = $Dir
    $ps.RedirectStandardOutput = $false
    $ps.UseShellExecute = $true
    $p = [System.Diagnostics.Process]::Start($ps)
    Write-Host "[PID $($p.Id)] $Name started, log -> $logPath"
}

# RPC services first so the gateways have their dependencies.
Start-Svc "driversvc"    "$root\rpc\driversvc"    "go run .\driversvc.go -f .\etc\driversvc.yaml"    "driversvc.log"
Start-Svc "ordersvc"     "$root\rpc\ordersvc"     "go run .\ordersvc.go -f .\etc\ordersvc.yaml"      "ordersvc.log"
Start-Svc "dispatchsvc"  "$root\rpc\dispatchsvc"  "go run .\dispatchsvc.go -f .\etc\dispatchsvc.yaml" "dispatchsvc.log"
Start-Svc "adminsvc"     "$root\rpc\adminsvc"     "go run .\admin.go -f .\etc\admin.yaml"            "adminsvc.log"

# API gateways.
$env:DRIVER_GRPC_ADDR = "127.0.0.1:50055"
$env:ORDER_GRPC_ADDR = "127.0.0.1:50051"
$env:DISPATCH_GRPC_ADDR = "127.0.0.1:50056"
Start-Svc "api-driver"   "$root\api\driver"      "go run ."  "api-driver.log"
Start-Svc "api-passenger" "$root\api\passenger"   "go run ."  "api-passenger.log"
Start-Svc "api-admin"    "$root\api\admin"       "go run ."   "api-admin.log"

Write-Host "=========================================="
Write-Host "All backend services were started in the background."
Write-Host "Log directory: $logDir"
Write-Host "driver api default :18082 | passenger api :8091 | admin api :8717"
Write-Host "Start the frontend in separate terminals:"
Write-Host "  web/admin   -> http://127.0.0.1:5173"
Write-Host "  web/driver  -> http://127.0.0.1:5175"
Write-Host "=========================================="
