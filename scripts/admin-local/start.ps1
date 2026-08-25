# 管理后台最小本地服务集启动脚本。
#
# 本脚本仅启动管理 RPC、管理 HTTP API 与管理前端，不启动订单、司机、价格等下游服务，
# 以降低本地开发时的常驻内存。数据库凭据只能由 ADMINSVC_MYSQL_DSN 环境变量注入。

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot)),
    [switch]$Stop
)

$ErrorActionPreference = "Stop"
$logDir = Join-Path $RepoRoot ".gotmp\admin-local-logs"
$pidFile = Join-Path $logDir "pids.txt"
$adminSvcExecutable = Join-Path $RepoRoot ".gotmp\admin-local-bin\adminsvc.exe"
$adminAPIExecutable = Join-Path $RepoRoot "api\admin\admin-api-local-20260822.exe"
$viteExecutable = Join-Path $RepoRoot "web\admin\node_modules\vite\bin\vite.js"

# Stop-StartedServices 只停止本脚本记录的子进程，不扫描或影响用户手工启动的服务。
function Stop-StartedServices {
    if (-not (Test-Path -LiteralPath $pidFile)) {
        Write-Host "未找到本地启动记录，无需停止。"
        return
    }

    Get-Content -LiteralPath $pidFile | ForEach-Object {
        $processID = [int]$_
        $process = Get-Process -Id $processID -ErrorAction SilentlyContinue
        if ($null -ne $process) {
            Stop-Process -Id $processID
            Write-Host "已停止本脚本启动的进程：$($process.ProcessName) (PID $processID)"
        }
    }
    Remove-Item -LiteralPath $pidFile -Force -ErrorAction SilentlyContinue
}

# Test-LocalPortAvailable 返回端口是否尚未被其他进程监听，避免误连入旧服务。
function Test-LocalPortAvailable {
    param([Parameter(Mandatory = $true)][int]$Port)

    return $null -eq (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -First 1)
}

# Wait-LocalPort 等待指定端口监听，并限制等待时长，防止启动失败时脚本无限阻塞。
function Wait-LocalPort {
    param(
        [Parameter(Mandatory = $true)][int]$Port,
        [int]$TimeoutSeconds = 20
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        if (-not (Test-LocalPortAvailable -Port $Port)) {
            return $true
        }
        Start-Sleep -Milliseconds 250
    }
    return $false
}

# Start-ManagedProcess 启动单个子进程并记录 PID；所有日志写入已忽略的 .gotmp 目录。
function Start-ManagedProcess {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$FilePath,
        [string[]]$ArgumentList = @(),
        [Parameter(Mandatory = $true)][string]$WorkingDirectory
    )

    $stdout = Join-Path $logDir "$Name.out.log"
    $stderr = Join-Path $logDir "$Name.err.log"
    # Windows PowerShell 5.1 不接受空 ArgumentList，未传参数的 API 进程需单独启动。
    if ($ArgumentList -and $ArgumentList.Count -gt 0) {
        $process = Start-Process -FilePath $FilePath -ArgumentList $ArgumentList -WorkingDirectory $WorkingDirectory `
            -RedirectStandardOutput $stdout -RedirectStandardError $stderr -WindowStyle Hidden -PassThru
    } else {
        $process = Start-Process -FilePath $FilePath -WorkingDirectory $WorkingDirectory `
            -RedirectStandardOutput $stdout -RedirectStandardError $stderr -WindowStyle Hidden -PassThru
    }
    Add-Content -LiteralPath $pidFile -Value $process.Id
    Write-Host "$Name 已启动，PID=$($process.Id)，日志：$stdout"
}

# Fail-Startup 清理本次已记录的子进程后中止启动，避免部分启动成功时遗留常驻进程。
function Fail-Startup {
    param([Parameter(Mandatory = $true)][string]$Message)

    Stop-StartedServices
    throw $Message
}

if ($Stop) {
    Stop-StartedServices
    exit 0
}

if ([string]::IsNullOrWhiteSpace($env:ADMINSVC_MYSQL_DSN)) {
    throw "缺少 ADMINSVC_MYSQL_DSN，拒绝在脚本中保存数据库凭据。"
}

foreach ($path in @($adminSvcExecutable, $adminAPIExecutable, $viteExecutable)) {
    if (-not (Test-Path -LiteralPath $path)) {
        throw "缺少本地启动依赖：$path。请先按项目构建流程生成可执行文件或安装前端依赖。"
    }
}

foreach ($port in @(8084, 8717, 5173)) {
    if (-not (Test-LocalPortAvailable -Port $port)) {
        throw "端口 $port 已被占用。为避免影响已有服务，本脚本未执行任何启动操作。"
    }
}

New-Item -ItemType Directory -Force -Path $logDir | Out-Null
Set-Content -LiteralPath $pidFile -Value @() -Encoding ASCII

# 临时配置取消 etcd 注册，并保持下游 RPC 非阻塞，避免最小服务集因未启动下游而阻塞。
$adminSvcConfigPath = Join-Path $logDir "adminsvc.local.yaml"
@"
Name: admin.rpc
ListenOn: 0.0.0.0:8084
MySQL:
  DSN: ""
Cache:
  Host: 115.191.16.159:6379
<<<<<<< HEAD
  Password: "4ay1nkal3u8ed77y"
=======
  Password: ""
>>>>>>> origin/develop
  DB: 0
Session:
  SessionTTLHours: 24
  TokenPrefix: "admin:sess:"
OrdersRPC:
  Target: 127.0.0.1:50051
  NonBlock: true
DriversRPC:
  Target: 127.0.0.1:8080
  NonBlock: true
PricesRPC:
  Target: 127.0.0.1:50053
  NonBlock: true
DisableDownstreamRPC: true
"@ | Set-Content -LiteralPath $adminSvcConfigPath -Encoding ASCII

# Go 运行时限制空闲堆保留；Node 堆上限为 384 MB，保留 Vite 编译所需余量。
$env:GOMEMLIMIT = "256MiB"
$env:GOGC = "75"
Start-ManagedProcess -Name "adminsvc" -FilePath $adminSvcExecutable -ArgumentList @("-f", $adminSvcConfigPath) `
    -WorkingDirectory (Join-Path $RepoRoot "rpc\adminsvc")
if (-not (Wait-LocalPort -Port 8084)) {
    Fail-Startup -Message "adminsvc 未在 8084 监听，请查看 $(Join-Path $logDir 'adminsvc.err.log')。"
}

Start-ManagedProcess -Name "admin-api" -FilePath $adminAPIExecutable -WorkingDirectory (Join-Path $RepoRoot "api\admin")
if (-not (Wait-LocalPort -Port 8717)) {
    Fail-Startup -Message "管理 API 未在 8717 监听，请查看 $(Join-Path $logDir 'admin-api.err.log')。"
}

# 直接调用 node 与 Vite 入口，规避 PowerShell 对 npm.ps1 的执行策略限制。
$env:NODE_OPTIONS = "--max-old-space-size=384"
Start-ManagedProcess -Name "admin-web" -FilePath "node.exe" -ArgumentList @($viteExecutable, "--host", "127.0.0.1") `
    -WorkingDirectory (Join-Path $RepoRoot "web\admin")
if (-not (Wait-LocalPort -Port 5173)) {
    Fail-Startup -Message "管理前端未在 5173 监听，请查看 $(Join-Path $logDir 'admin-web.err.log')。"
}

Write-Host "管理后台最小服务集已启动： http://127.0.0.1:5173"

