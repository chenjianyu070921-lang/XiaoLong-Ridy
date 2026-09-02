# Starts the admin HTTP/RPC test stack for the admin module:
#   ordersvc (50051) + driversvc (50055) -> adminsvc (8084) -> api/admin (8717)
#
# 默认启动真实 ordersvc，确保取消、退款、改派能够进行真实跨服务联调。
# 只有显式传入 -UseDummyOrders 才启动占位 gRPC；该模式只适用于路由/只读测试，
# 不得作为订单写操作闭环通过的依据。
#
# Usage (PowerShell):
#   .\scripts\admin-test\start_admin_stack.ps1
#   .\scripts\admin-test\start_admin_stack.ps1 -Stop
#
# Requirements:
#   - Go toolchain in PATH
#   - Remote MySQL (xiaolong_ridy) reachable
#   - Local Redis running on 127.0.0.1:6379

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot)),
    [switch]$SkipBuild,
    [switch]$UseDummyOrders,
    [switch]$Stop
)

$ErrorActionPreference = "Stop"
$logDir = Join-Path $RepoRoot ".gotmp\admin-test-logs"
$binDir = Join-Path $RepoRoot ".gotmp\bin"
$pidFile = Join-Path $logDir "pids.txt"
$modeFile = Join-Path $logDir "ordersvc-mode.txt"
New-Item -ItemType Directory -Force -Path $logDir | Out-Null
New-Item -ItemType Directory -Force -Path $binDir | Out-Null

function Stop-StartedServices {
    if (-not (Test-Path -LiteralPath $pidFile)) {
        Write-Host "No pid file found. Nothing to stop."
        return
    }
    Get-Content -LiteralPath $pidFile | ForEach-Object {
        $procId = [int]$_
        try {
            $p = Get-Process -Id $procId -ErrorAction Stop
            Stop-Process -Id $procId -Force
            Write-Host "Stopped $($p.ProcessName) pid=$procId"
        } catch {
            Write-Host "Process $procId already exited."
        }
    }
    Remove-Item -LiteralPath $pidFile -Force -ErrorAction SilentlyContinue
}

if ($Stop) {
    Stop-StartedServices
    exit 0
}

# Use the project-local Go module/cache dirs so we do not touch the user's global cache.
$env:GOMODCACHE = Join-Path $RepoRoot ".gotmp\pkg\mod"
$env:GOCACHE = Join-Path $RepoRoot ".gotmp\gocache"

# 后台服务的真实数据库凭据只允许由调用环境注入，测试脚本不得保存或生成共享库密码。
if ([string]::IsNullOrWhiteSpace($env:ADMINSVC_MYSQL_DSN)) {
    throw "请先设置 ADMINSVC_MYSQL_DSN"
}
if (-not $UseDummyOrders -and [string]::IsNullOrWhiteSpace($env:ORDERSVC_MYSQL_DSN)) {
    throw "真实订单联调需要先设置 ORDERSVC_MYSQL_DSN；仅做路由/只读测试时可显式传入 -UseDummyOrders"
}
if ([string]::IsNullOrWhiteSpace($env:ADMIN_API_MYSQL_DSN)) {
    $env:ADMIN_API_MYSQL_DSN = $env:ADMINSVC_MYSQL_DSN
}

# adminsvc needs a test config without etcd registration (the default admin.yaml
# requires a local etcd on 2379 which is not part of this test stack).
$adminsvcTestCfg = @"
Name: admin.rpc
ListenOn: 0.0.0.0:8084
MySQL:
  DSN: ""
Cache:
  Host: 115.191.16.159:6379
  Password: ""
  DB: 0
Session:
  SessionTTLHours: 24
  TokenPrefix: "admin:sess:"
OrdersRPC:
  Target: 127.0.0.1:50051
DriversRPC:
  Target: 127.0.0.1:50055
"@
$adminsvcCfgPath = Join-Path $RepoRoot ".gotmp\adminsvc-admin-test.yaml"
Set-Content -LiteralPath $adminsvcCfgPath -Value $adminsvcTestCfg -Encoding ASCII
Write-Host "adminsvc test config: $adminsvcCfgPath"

$ordersvcTestCfg = @"
Name: ordersvc.rpc
ListenOn: 0.0.0.0:50051
mysql:
  dsn: $env:ORDERSVC_MYSQL_DSN
kafka:
  brokers:
    - 115.191.16.159:9092
  topic: order.created
  group: order-event-consumer
dispatchrpc:
  Target: 127.0.0.1:18083
pricerpc:
  Target: 127.0.0.1:50053
payrpc:
  Target: 127.0.0.1:50054
paychannel: 1
"@
$ordersvcCfgPath = Join-Path $RepoRoot ".gotmp\ordersvc-admin-test.yaml"
Set-Content -LiteralPath $ordersvcCfgPath -Value $ordersvcTestCfg -Encoding ASCII
Write-Host "ordersvc test config: $ordersvcCfgPath"

if (-not $SkipBuild) {
    Push-Location $RepoRoot
    try {
        Write-Host "Building ordersvc..."
        go build -o .gotmp\bin\ordersvc.exe ./rpc/ordersvc/ordersvc.go
        Write-Host "Building driversvc..."
        go build -o .gotmp\bin\driversvc.exe ./rpc/driversvc/driversvc.go
        Write-Host "Building adminsvc..."
        go build -o .gotmp\bin\adminsvc.exe ./rpc/adminsvc/admin.go
        Write-Host "Building api/admin..."
        Push-Location (Join-Path $RepoRoot "api\admin")
        go build -o ..\..\.gotmp\bin\admin-api.exe .
        Pop-Location
        if ($UseDummyOrders) {
            Write-Host "Building dummygrpc..."
            go build -o .gotmp\bin\dummygrpc.exe .\.gotmp\dummygrpc\main.go
        }
    } finally {
        Pop-Location
    }
}

function Start-ServiceProcess {
    param([string]$Name, [string]$Exe, [string[]]$ArgList, [string]$WorkingDir)
    $outLog = Join-Path $logDir "$Name.out.log"
    $errLog = Join-Path $logDir "$Name.err.log"
    # Windows PowerShell 5.1 rejects an empty ArgumentList (@()) in Start-Process,
    # so only pass -ArgumentList when there are actual arguments.
    if ($ArgList -and $ArgList.Count -gt 0) {
        $p = Start-Process -FilePath $Exe -ArgumentList $ArgList -WorkingDirectory $WorkingDir `
            -RedirectStandardOutput $outLog -RedirectStandardError $errLog `
            -WindowStyle Hidden -PassThru
    } else {
        $p = Start-Process -FilePath $Exe -WorkingDirectory $WorkingDir `
            -RedirectStandardOutput $outLog -RedirectStandardError $errLog `
            -WindowStyle Hidden -PassThru
    }
    Add-Content -LiteralPath $pidFile -Value $p.Id
    Write-Host "$Name started (pid=$($p.Id)) log=$outLog"
    return $p
}

function Wait-TcpPort {
    param([int]$Port, [int]$TimeoutSec = 60)
    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        $client = New-Object System.Net.Sockets.TcpClient
        try {
            $async = $client.BeginConnect("127.0.0.1", $Port, $null, $null)
            if ($async.AsyncWaitHandle.WaitOne(1000)) {
                $client.EndConnect($async)
                $client.Close()
                return $true
            }
            $client.Close()
        } catch {
            $client.Close()
        }
        Start-Sleep -Milliseconds 500
    }
    return $false
}

if ($UseDummyOrders) {
    Set-Content -LiteralPath $modeFile -Value "dummy" -Encoding ASCII
    Start-ServiceProcess -Name "dummygrpc" -Exe (Join-Path $binDir "dummygrpc.exe") `
        -ArgList @() -WorkingDir $RepoRoot | Out-Null
    Start-Sleep -Seconds 1
    Write-Host "WARNING: using dummy ordersvc; order write operations are not end-to-end validated."
} else {
    Set-Content -LiteralPath $modeFile -Value "real" -Encoding ASCII
    Start-ServiceProcess -Name "ordersvc" -Exe (Join-Path $binDir "ordersvc.exe") `
        -ArgList @("-f", $ordersvcCfgPath) -WorkingDir $RepoRoot | Out-Null
    Write-Host "Waiting for ordersvc:50051 ..."
    if (-not (Wait-TcpPort -Port 50051)) {
        Write-Host "ERROR: ordersvc did not start. See $logDir\ordersvc.err.log"
        throw "真实 ordersvc 启动失败，订单跨服务联调阻塞"
    }
}

$driversvc = Start-ServiceProcess -Name "driversvc" -Exe (Join-Path $binDir "driversvc.exe") `
    -ArgList @("-f", (Join-Path $RepoRoot "rpc\driversvc\etc\driversvc.yaml")) -WorkingDir $RepoRoot

Write-Host "Waiting for driversvc:50055 ..."
if (-not (Wait-TcpPort -Port 50055)) { Write-Host "ERROR: driversvc did not start. See $logDir\driversvc.err.log" }

$adminsvc = Start-ServiceProcess -Name "adminsvc" -Exe (Join-Path $binDir "adminsvc.exe") `
    -ArgList @("-f", ".gotmp\adminsvc-admin-test.yaml") -WorkingDir $RepoRoot

Write-Host "Waiting for adminsvc:8084 ..."
if (-not (Wait-TcpPort -Port 8084)) { Write-Host "ERROR: adminsvc did not start. See $logDir\adminsvc.err.log" }

$adminApi = Start-ServiceProcess -Name "admin-api" -Exe (Join-Path $binDir "admin-api.exe") `
    -ArgList @() -WorkingDir (Join-Path $RepoRoot "api\admin")

Write-Host "Waiting for api/admin:8717 ..."
if (-not (Wait-TcpPort -Port 8717)) { Write-Host "ERROR: api/admin did not start. See $logDir\admin-api.err.log" }

Write-Host ""
Write-Host "Admin test stack is up:"
Write-Host "  api/admin   http://127.0.0.1:8717"
Write-Host "  adminsvc    grpc 127.0.0.1:8084"
Write-Host "  driversvc   grpc 127.0.0.1:5055"
if ($UseDummyOrders) {
    Write-Host "  dummygrpc   grpc 127.0.0.1:15051 (ordersvc placeholder; no write-flow validation)"
} else {
    Write-Host "  ordersvc    grpc 127.0.0.1:50051 (real cross-service target)"
}
Write-Host ""
Write-Host "Next: .\scripts\admin-test\admin_api_test.ps1"
Write-Host "Stop stack with: .\scripts\admin-test\start_admin_stack.ps1 -Stop"
