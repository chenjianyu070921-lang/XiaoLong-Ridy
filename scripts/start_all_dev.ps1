# 开发环境一键启动脚本（司机端 + 订单 + 后台 + 乘客 全部业务）
# 用法：powershell -ExecutionPolicy Bypass -File scripts/start_all_dev.ps1
$ErrorActionPreference = "Continue"
# 根据脚本所在目录推导仓库根目录，避免本机工作区路径变化导致服务无法启动。
$root = Split-Path -Parent $PSScriptRoot
# GOCACHE 之前遇到过损坏，固定到独立目录避免占用临时盘
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

# ---- 下游 RPC 服务（先起，网关依赖它们）----
Start-Svc "driversvc"  "$root\rpc\driversvc"  "go run .\driversvc.go -f .\etc\driversvc.yaml"  "driversvc.log"
Start-Svc "ordersvc"   "$root\rpc\ordersvc"   "go run .\ordersvc.go -f .\etc\ordersvc.yaml"    "ordersvc.log"
Start-Svc "dispatchsvc" "$root\rpc\dispatchsvc" "go run .\dispatchsvc.go -f .\etc\dispatchsvc.yaml" "dispatchsvc.log"
# 订单事件消费者必须在网关接收下单前启动，负责消费 order.created 并触发派单。
Start-Svc "order-event-consumer" "$root\mq-consumer\order-event-consumer" "go run .\main.go -f .\etc\order-event-consumer.yaml" "order-event-consumer.log"
Start-Svc "adminsvc"   "$root\rpc\adminsvc"   "go run .\admin.go -f .\etc\admin.yaml"          "adminsvc.log"

# ---- API 网关 ----
$env:DRIVER_GRPC_ADDR = "127.0.0.1:50055"
$env:ORDER_GRPC_ADDR = "127.0.0.1:50051"
$env:DISPATCH_GRPC_ADDR = "127.0.0.1:50056"
Start-Svc "api-driver"   "$root\api\driver"   "go run ."  "api-driver.log"
Start-Svc "api-passenger" "$root\api\passenger" "go run ." "api-passenger.log"
Start-Svc "api-admin"    "$root\api\admin"    "go run ."  "api-admin.log"

Write-Host "=========================================="
Write-Host "所有后端服务已在后台启动，等待编译+监听..."
Write-Host "日志目录: $logDir"
Write-Host "driver api 默认 :8082 | passenger api :8091 | admin api :8717"
Write-Host "前端请另开终端启动："
Write-Host "  后台web: cd web/admin; pnpm install; pnpm dev  -> http://127.0.0.1:5173"
Write-Host "  司机端web: cd web/driver; pnpm install; pnpm dev  -> http://127.0.0.1:5175"
Write-Host "=========================================="
