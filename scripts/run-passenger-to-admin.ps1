# 真实乘客端 → 管理后台 全链路本地启动脚本。
#
# 启动的服务与端口：
#   usersvc(50052) ordersvc(50051) dispatchsvc(50056) pricesvc(50053) paysvc(50054)
#   driversvc(50055) locationsvc(50057)
#   adminsvc(8084) api/passenger(8091) api/admin(8717) api/driver(8082)
#   管理前端 vite(5173) 司机前端 vite(5175) 乘客端 H5 vite(5174)
#
# 数据库/Redis/签名凭据不硬编码在脚本中，统一从 rpc/usersvc/etc/usersvc.yaml
# 提取（该文件已包含真实 DSN 与 Redis 密码），adminsvc 的凭据经环境变量注入。
#
# 用法（PowerShell）：
#   .\scripts\run-passenger-to-admin.ps1          # 构建并启动全链路
#   .\scripts\run-passenger-to-admin.ps1 -SkipBuild   # 跳过构建直接启动
#   .\scripts\run-passenger-to-admin.ps1 -Stop    # 停止本脚本启动的服务

param(
    [string]$RepoRoot = (Split-Path -Parent $PSScriptRoot),
    [switch]$Stop,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$logDir = Join-Path $RepoRoot ".gotmp\runtime-logs"
$binDir = Join-Path $RepoRoot ".gotmp\runtime-bin"
$pidFile = Join-Path $logDir "pids.txt"
$adminsvcCfg = Join-Path $logDir "adminsvc.runtime.yaml"

# Stop-StartedServices 只停止本脚本记录的子进程。
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
            Write-Host "已停止：$($process.ProcessName) (PID $processID)"
        }
    }
    Remove-Item -LiteralPath $pidFile -Force -ErrorAction SilentlyContinue
}

# Test-LocalPortAvailable 返回端口是否尚未被监听。
function Test-LocalPortAvailable {
    param([Parameter(Mandatory = $true)][int]$Port)
    return $null -eq (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -First 1)
}

# Wait-LocalPort 等待指定端口监听，限制等待时长。
function Wait-LocalPort {
    param(
        [Parameter(Mandatory = $true)][int]$Port,
        # 依赖的 MySQL/Redis 在远端，服务启动时的建表与连接探测可能耗时数十秒，
        # 超时窗口需明显大于本地场景，否则会误判为启动失败。
        [int]$TimeoutSeconds = 120
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

# Start-ManagedProcess 启动单个子进程并记录 PID。
function Start-ManagedProcess {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$FilePath,
        [string[]]$ArgumentList = @(),
        [Parameter(Mandatory = $true)][string]$WorkingDirectory
    )
    $stdout = Join-Path $logDir "$Name.out.log"
    $stderr = Join-Path $logDir "$Name.err.log"
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

# Fail-Startup 清理已启动进程后中止。
function Fail-Startup {
    param([Parameter(Mandatory = $true)][string]$Message)
    Stop-StartedServices
    throw $Message
}

if ($Stop) {
    Stop-StartedServices
    exit 0
}

New-Item -ItemType Directory -Force -Path $logDir | Out-Null
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
Set-Content -LiteralPath $pidFile -Value @() -Encoding ASCII

# 目标端口预检，避免误连入旧服务。
foreach ($port in @(50051, 50052, 50053, 50054, 50055, 50056, 50057, 8084, 8091, 8717, 8082, 5173, 5174, 5175)) {
    if (-not (Test-LocalPortAvailable -Port $port)) {
        throw "端口 $port 已被占用。为避免影响已有服务，本脚本未执行任何启动操作。"
    }
}

# 从 usersvc.yaml 提取共享的数据库 DSN、Redis 密码与令牌签名密钥。
$usersvcCfg = Get-Content -LiteralPath (Join-Path $RepoRoot "rpc\usersvc\etc\usersvc.yaml")
$dsnLine = $usersvcCfg | Where-Object { $_ -match '^\s*dsn:' } | Select-Object -First 1
$dsn = (($dsnLine -replace '^\s*dsn:\s*', '').Trim().Trim('"'))
$redisPassLine = $usersvcCfg | Where-Object { $_ -match '^\s*pass:' } | Select-Object -First 1
$redisPass = (($redisPassLine -replace '^\s*pass:\s*', '').Trim().Trim('"'))
$signingLine = $usersvcCfg | Where-Object { $_ -match '^\s*signingKey:' } | Select-Object -First 1
$signingKey = (($signingLine -replace '^\s*signingKey:\s*', '').Trim().Trim('"'))
if ([string]::IsNullOrWhiteSpace($dsn)) { throw "未能从 usersvc.yaml 提取 mysql dsn" }
if ([string]::IsNullOrWhiteSpace($redisPass)) { throw "未能从 usersvc.yaml 提取 redis pass" }

# driversvc 与 api/driver 会拒绝默认占位签名密钥（local-development-signing-key）。
# 当 usersvc.yaml 仍使用占位值时，改用持久化的本地开发密钥文件（位于 git 忽略的 .gotmp 下，
# 不写入仓库；跨重启保持稳定，避免每次启动都让已有登录态失效）。
$defaultSigningKey = "local-development-signing-key"
if ([string]::IsNullOrWhiteSpace($signingKey) -or $signingKey -eq $defaultSigningKey) {
    $localKeyFile = Join-Path $logDir "local-signing.key"
    if (-not (Test-Path -LiteralPath $localKeyFile)) {
        $generated = ([guid]::NewGuid().ToString("N") + [guid]::NewGuid().ToString("N"))
        Set-Content -LiteralPath $localKeyFile -Value $generated -Encoding ASCII -NoNewline
    }
    $signingKey = (Get-Content -LiteralPath $localKeyFile -Raw).Trim()
}
if ([string]::IsNullOrWhiteSpace($signingKey)) { throw "未能确定本地共享令牌签名密钥" }

# 生成 adminsvc 无 etcd 的临时配置，下游 RPC 全部指向本机真实端口并懒连接。
$adminsvcTemplate = @"
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
MenuRoles:
  1:
  - Name: 管理员
    Path: /admins
    Icon: Shield
    Perm: admin:manage
  - Name: 用户管理
    Path: /users
    Icon: User
    Perm: user:list
  - Name: 司机审核
    Path: /driver-certifications
    Icon: BadgeCheck
    Perm: driver:audit
  - Name: 司机列表
    Path: /drivers
    Icon: Avatar
    Perm: driver:list
  - Name: 司机提现
    Path: /driver-withdrawals
    Icon: Money
    Perm: driver:withdraw:list
  - Name: 订单监控
    Path: /orders
    Icon: ClipboardList
    Perm: order:list
  - Name: 异常订单
    Path: /orders/abnormal
    Icon: Warning
    Perm: order:abnormal
  - Name: 操作日志
    Path: /operation-logs
    Icon: ScrollText
    Perm: log:list
  - Name: 优惠券模板
    Path: /coupons
    Icon: Discount
    Perm: coupon:list
  - Name: 发券任务
    Path: /coupon-issue-tasks
    Icon: List
    Perm: coupon:issue_task:list
  - Name: 计价规则
    Path: /price-rules
    Icon: Setting
    Perm: price_rule:list
  - Name: 活动配置
    Path: /promotion-activities
    Icon: Discount
    Perm: promotion:list
  - Name: 投诉与申诉工单
    Path: /work-orders
    Icon: Document
    Perm: work-order:list
  - Name: 数据统计
    Path: /statistics
    Icon: DataAnalysis
    Perm: statistics:view
  - Name: 导出任务
    Path: /export-tasks
    Icon: Document
    Perm: export:list
  - Name: 黑名单
    Path: /blacklist
    Icon: Warning
    Perm: risk:blacklist:list
  - Name: 风控命中记录
    Path: /risk-hits
    Icon: Management
    Perm: risk:hit:list
  2:
  - Name: 用户管理
    Path: /users
    Icon: User
    Perm: user:list
  - Name: 司机审核
    Path: /driver-certifications
    Icon: BadgeCheck
    Perm: driver:audit
  - Name: 司机列表
    Path: /drivers
    Icon: Avatar
    Perm: driver:list
  - Name: 司机提现
    Path: /driver-withdrawals
    Icon: Money
    Perm: driver:withdraw:list
  - Name: 订单监控
    Path: /orders
    Icon: ClipboardList
    Perm: order:list
  - Name: 异常订单
    Path: /orders/abnormal
    Icon: Warning
    Perm: order:abnormal
  - Name: 操作日志
    Path: /operation-logs
    Icon: ScrollText
    Perm: log:list
  - Name: 优惠券模板
    Path: /coupons
    Icon: Discount
    Perm: coupon:list
  - Name: 发券任务
    Path: /coupon-issue-tasks
    Icon: List
    Perm: coupon:issue_task:list
  - Name: 计价规则
    Path: /price-rules
    Icon: Setting
    Perm: price_rule:list
  - Name: 活动配置
    Path: /promotion-activities
    Icon: Discount
    Perm: promotion:list
  - Name: 投诉与申诉工单
    Path: /work-orders
    Icon: Document
    Perm: work-order:list
  - Name: 数据统计
    Path: /statistics
    Icon: DataAnalysis
    Perm: statistics:view
  - Name: 导出任务
    Path: /export-tasks
    Icon: Document
    Perm: export:list
  - Name: 黑名单
    Path: /blacklist
    Icon: Warning
    Perm: risk:blacklist:list
  - Name: 风控命中记录
    Path: /risk-hits
    Icon: Management
    Perm: risk:hit:list
  3:
  - Name: 用户管理
    Path: /users
    Icon: User
    Perm: user:list
  - Name: 司机审核
    Path: /driver-certifications
    Icon: BadgeCheck
    Perm: driver:audit
  - Name: 司机列表
    Path: /drivers
    Icon: Avatar
    Perm: driver:list
  - Name: 订单监控
    Path: /orders
    Icon: ClipboardList
    Perm: order:list
  - Name: 异常订单
    Path: /orders/abnormal
    Icon: Warning
    Perm: order:abnormal
  - Name: 操作日志
    Path: /operation-logs
    Icon: ScrollText
    Perm: log:list
  - Name: 投诉与申诉工单
    Path: /work-orders
    Icon: Document
    Perm: work-order:list
OrdersRPC:
  target: 127.0.0.1:50051
  nonblock: true
  timeout: 30000
DispatchRPC:
  target: 127.0.0.1:50056
  nonblock: true
  timeout: 30000
UsersRPC:
  target: 127.0.0.1:50052
  nonblock: true
  timeout: 30000
DriversRPC:
  target: 127.0.0.1:50055
  nonblock: true
  timeout: 30000
PricesRPC:
  target: 127.0.0.1:50053
  nonblock: true
  timeout: 30000
LocationsRPC:
  target: 127.0.0.1:50057
  nonblock: true
  timeout: 30000
PushRPC:
  target: 127.0.0.1:9002
  nonblock: true
  timeout: 30000
"@
Set-Content -LiteralPath $adminsvcCfg -Value $adminsvcTemplate -Encoding UTF8

# 生成 locationsvc 本地运行时配置：去掉 Etcd 注册（本地无 etcd），凭据复用共享 DSN/Redis，
# 地图 ApiKey 与默认城市编码从 locationsvc 原始配置提取，避免在脚本中硬编码。
$locationsvcCfg = Join-Path $logDir "locationsvc.runtime.yaml"
$locationsvcSrc = Get-Content -LiteralPath (Join-Path $RepoRoot "rpc\locationsvc\etc\locationsvc.yaml")
$amapKeyLine = $locationsvcSrc | Where-Object { $_ -match '^\s*ApiKey:' } | Select-Object -First 1
$amapKey = (($amapKeyLine -replace '^\s*ApiKey:\s*', '').Trim().Trim('"'))
if ([string]::IsNullOrWhiteSpace($amapKey)) { throw "未能从 locationsvc.yaml 提取 amap ApiKey" }
$cityCodeLine = $locationsvcSrc | Where-Object { $_ -match '^\s*defaultCityCode:' } | Select-Object -First 1
$cityCode = (($cityCodeLine -replace '^\s*defaultCityCode:\s*', '').Trim().Trim('"'))
if ([string]::IsNullOrWhiteSpace($cityCode)) { throw "未能从 locationsvc.yaml 提取 defaultCityCode" }
$locationsvcTemplate = @"
Name: locationsvc.rpc
defaultCityCode: "$cityCode"
ListenOn: 0.0.0.0:50057
Timeout: 30000
Mysql:
  Dsn: "$dsn"
  MaxOpenConn: 100
  MaxIdleConn: 10
  MaxLifeTime: 3600
myredis:
  host: 115.191.16.159:6379
  pass: "$redisPass"
  db: 0
  poolSize: 20
  dialTimeout: 0
  readTimeout: 0
  writeTimeout: 0
MapService:
  ApiKey: "$amapKey"
  Provider: "amap"
  BaseUrl: "https://restapi.amap.com/v3"
Log:
  ServiceName: locationsvc
  Mode: console
  Level: info
"@
Set-Content -LiteralPath $locationsvcCfg -Value $locationsvcTemplate -Encoding UTF8

# 构建全部服务二进制（-SkipBuild 可跳过）。
if (-not $SkipBuild) {
    Push-Location $RepoRoot
    try {
        $targets = @(
            @{Name = "usersvc";      Pkg = "./rpc/usersvc/main.go"},
            @{Name = "ordersvc";     Pkg = "./rpc/ordersvc/ordersvc.go"},
            @{Name = "dispatchsvc";  Pkg = "./rpc/dispatchsvc/dispatchsvc.go"},
            @{Name = "pricesvc";     Pkg = "./rpc/pricesvc/pricesvc.go"},
            @{Name = "paysvc";       Pkg = "./rpc/paysvc/paysvc.go"},
            @{Name = "driversvc";    Pkg = "./rpc/driversvc/driversvc.go"},
            @{Name = "locationsvc";  Pkg = "./rpc/locationsvc/locationsvc.go"},
            @{Name = "adminsvc";     Pkg = "./rpc/adminsvc/admin.go"},
            @{Name = "passenger-api";Pkg = "./api/passenger"},
            @{Name = "admin-api";    Pkg = "./api/admin"},
            @{Name = "driver-api";   Pkg = "./api/driver"}
        )
        foreach ($t in $targets) {
            Write-Host "构建 $($t.Name) ..."
            go build -o (Join-Path $binDir "$($t.Name).exe") $t.Pkg
            if ($LASTEXITCODE -ne 0) { throw "构建 $($t.Name) 失败" }
        }
    } finally {
        Pop-Location
    }
}

foreach ($path in @(
    (Join-Path $binDir "usersvc.exe"),
    (Join-Path $binDir "ordersvc.exe"),
    (Join-Path $binDir "dispatchsvc.exe"),
    (Join-Path $binDir "pricesvc.exe"),
    (Join-Path $binDir "paysvc.exe"),
    (Join-Path $binDir "driversvc.exe"),
    (Join-Path $binDir "locationsvc.exe"),
    (Join-Path $binDir "adminsvc.exe"),
    (Join-Path $binDir "passenger-api.exe"),
    (Join-Path $binDir "admin-api.exe"),
    (Join-Path $binDir "driver-api.exe")
)) {
    if (-not (Test-Path -LiteralPath $path)) {
        throw "缺少构建产物：$path"
    }
}

# adminsvc 真实凭据通过环境变量注入。
$env:ADMINSVC_MYSQL_DSN = $dsn
$env:ADMINSVC_REDIS_PASSWORD = $redisPass
# 乘客端与 usersvc 共享令牌签名密钥，保证 JWT 可互验。
$env:JWT_SIGNING_KEY = $signingKey
$env:PASSENGER_TOKEN_SIGNING_KEY = $signingKey
# 司机端网关与 driversvc 共享令牌签名密钥，保证司机 JWT 可互验。
$env:DRIVER_SIGNING_KEY = $signingKey
$env:DRIVERSVC_SIGNING_KEY = $signingKey

# Go 运行时限制空闲堆保留。
$env:GOMEMLIMIT = "256MiB"
$env:GOGC = "75"

# 1. usersvc（无下游 RPC 依赖）
Start-ManagedProcess -Name "usersvc" -FilePath (Join-Path $binDir "usersvc.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "rpc\usersvc\etc\usersvc.yaml")) `
    -WorkingDirectory $RepoRoot
if (-not (Wait-LocalPort -Port 50052)) { Fail-Startup -Message "usersvc 未在 50052 监听，请查看 $logDir\usersvc.err.log" }

# 2. dispatchsvc（对 ordersvc 为懒连接，可先行启动）
Start-ManagedProcess -Name "dispatchsvc" -FilePath (Join-Path $binDir "dispatchsvc.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "rpc\dispatchsvc\etc\dispatchsvc.yaml")) `
    -WorkingDirectory $RepoRoot
if (-not (Wait-LocalPort -Port 50056)) { Fail-Startup -Message "dispatchsvc 未在 50056 监听，请查看 $logDir\dispatchsvc.err.log" }

# 3. pricesvc（无下游 RPC 依赖）
Start-ManagedProcess -Name "pricesvc" -FilePath (Join-Path $binDir "pricesvc.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "rpc\pricesvc\etc\pricesvc.yaml")) `
    -WorkingDirectory $RepoRoot
if (-not (Wait-LocalPort -Port 50053)) { Fail-Startup -Message "pricesvc 未在 50053 监听，请查看 $logDir\pricesvc.err.log" }

# 4. ordersvc（启动时阻塞拨号 dispatchsvc 50056 / pricesvc 50053，必须在其后启动）
Start-ManagedProcess -Name "ordersvc" -FilePath (Join-Path $binDir "ordersvc.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "rpc\ordersvc\etc\ordersvc.yaml")) `
    -WorkingDirectory $RepoRoot
if (-not (Wait-LocalPort -Port 50051)) { Fail-Startup -Message "ordersvc 未在 50051 监听，请查看 $logDir\ordersvc.err.log" }

# 5. paysvc（启动时阻塞拨号 ordersvc 50051，须在 ordersvc 之后启动）
Start-ManagedProcess -Name "paysvc" -FilePath (Join-Path $binDir "paysvc.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "rpc\paysvc\etc\paysvc.yaml")) `
    -WorkingDirectory $RepoRoot
if (-not (Wait-LocalPort -Port 50054)) { Fail-Startup -Message "paysvc 未在 50054 监听，请查看 $logDir\paysvc.err.log" }

# 6. driversvc（司机域，adminsvc 冻结/解冻/提现审核依赖）
Start-ManagedProcess -Name "driversvc" -FilePath (Join-Path $binDir "driversvc.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "rpc\driversvc\etc\driversvc.yaml")) `
    -WorkingDirectory $RepoRoot
if (-not (Wait-LocalPort -Port 50055)) { Fail-Startup -Message "driversvc 未在 50055 监听，请查看 $logDir\driversvc.err.log" }

# 7. locationsvc（订单轨迹/司机位置，无下游 RPC 依赖；本地无 etcd，用运行时配置）
Start-ManagedProcess -Name "locationsvc" -FilePath (Join-Path $binDir "locationsvc.exe") `
    -ArgumentList @("-f", $locationsvcCfg) `
    -WorkingDirectory (Join-Path $RepoRoot "rpc\locationsvc")
if (-not (Wait-LocalPort -Port 50057)) { Fail-Startup -Message "locationsvc 未在 50057 监听，请查看 $logDir\locationsvc.err.log" }

# 8. adminsvc
Start-ManagedProcess -Name "adminsvc" -FilePath (Join-Path $binDir "adminsvc.exe") `
    -ArgumentList @("-f", $adminsvcCfg) `
    -WorkingDirectory (Join-Path $RepoRoot "rpc\adminsvc")
if (-not (Wait-LocalPort -Port 8084)) { Fail-Startup -Message "adminsvc 未在 8084 监听，请查看 $logDir\adminsvc.err.log" }

# 9. api/passenger（8091）
Start-ManagedProcess -Name "passenger-api" -FilePath (Join-Path $binDir "passenger-api.exe") `
    -ArgumentList @() `
    -WorkingDirectory (Join-Path $RepoRoot "api\passenger")
if (-not (Wait-LocalPort -Port 8091)) { Fail-Startup -Message "api/passenger 未在 8091 监听，请查看 $logDir\passenger-api.err.log" }

# 10. api/admin（8717）
Start-ManagedProcess -Name "admin-api" -FilePath (Join-Path $binDir "admin-api.exe") `
    -ArgumentList @() `
    -WorkingDirectory (Join-Path $RepoRoot "api\admin")
if (-not (Wait-LocalPort -Port 8717)) { Fail-Startup -Message "api/admin 未在 8717 监听，请查看 $logDir\admin-api.err.log" }

# 11. api/driver（司机端网关 8082，下游全部指向本机 RPC）
$env:DRIVER_HTTP_ADDR = ":8082"
$env:DRIVER_GRPC_ADDR = "127.0.0.1:50055"
$env:ORDER_GRPC_ADDR = "127.0.0.1:50051"
$env:DISPATCH_GRPC_ADDR = "127.0.0.1:50056"
$env:LOCATION_GRPC_ADDR = "127.0.0.1:50057"
Start-ManagedProcess -Name "driver-api" -FilePath (Join-Path $binDir "driver-api.exe") `
    -ArgumentList @("-f", (Join-Path $RepoRoot "api\driver\etc\driver.yaml")) `
    -WorkingDirectory (Join-Path $RepoRoot "api\driver")
if (-not (Wait-LocalPort -Port 8082)) { Fail-Startup -Message "api/driver 未在 8082 监听，请查看 $logDir\driver-api.err.log" }

# 12. 管理前端 vite（5173）
if (Test-Path (Join-Path $RepoRoot "web\admin\node_modules\vite\bin\vite.js")) {
    $env:NODE_OPTIONS = "--max-old-space-size=384"
    Start-ManagedProcess -Name "admin-web" -FilePath "node.exe" `
        -ArgumentList @((Join-Path $RepoRoot "web\admin\node_modules\vite\bin\vite.js"), "--host", "127.0.0.1") `
        -WorkingDirectory (Join-Path $RepoRoot "web\admin")
    if (-not (Wait-LocalPort -Port 5173)) { Write-Host "警告：管理前端未在 5173 监听，请查看 $logDir\admin-web.err.log" }
} else {
    Write-Host "警告：web\admin 未安装 node_modules，跳过前端启动。"
}

# 13. 司机前端 vite（5175）
if (Test-Path (Join-Path $RepoRoot "web\driver\node_modules\vite\bin\vite.js")) {
    Start-ManagedProcess -Name "driver-web" -FilePath "node.exe" `
        -ArgumentList @((Join-Path $RepoRoot "web\driver\node_modules\vite\bin\vite.js"), "--host", "127.0.0.1") `
        -WorkingDirectory (Join-Path $RepoRoot "web\driver")
    if (-not (Wait-LocalPort -Port 5175)) { Write-Host "警告：司机前端未在 5175 监听，请查看 $logDir\driver-web.err.log" }
} else {
    Write-Host "警告：web\driver 未安装 node_modules，跳过司机前端启动。"
}

# 14. 乘客端 H5 vite（5174）
if (Test-Path (Join-Path $RepoRoot "web\user\node_modules\vite\bin\vite.js")) {
    Start-ManagedProcess -Name "user-web" -FilePath "node.exe" `
        -ArgumentList @((Join-Path $RepoRoot "web\user\node_modules\vite\bin\vite.js"), "--host", "127.0.0.1") `
        -WorkingDirectory (Join-Path $RepoRoot "web\user")
    if (-not (Wait-LocalPort -Port 5174)) { Write-Host "警告：乘客端前端未在 5174 监听，请查看 $logDir\user-web.err.log" }
} else {
    Write-Host "警告：web\user 未安装 node_modules，跳过乘客端前端启动。"
}

Write-Host ""
Write-Host "乘客端 → 后台管理 → 司机端 全链路已启动："
Write-Host "  乘客端 API      http://127.0.0.1:8091"
Write-Host "  乘客端 H5       http://127.0.0.1:5174"
Write-Host "  管理后台        http://127.0.0.1:5173  (API 经 8717)"
Write-Host "  司机端 API      http://127.0.0.1:8082"
Write-Host "  司机端 H5       http://127.0.0.1:5175"
Write-Host "  RPC: usersvc=50052 ordersvc=50051 dispatchsvc=50056 pricesvc=50053 paysvc=50054 driversvc=50055 locationsvc=50057 adminsvc=8084"
Write-Host "  日志目录：$logDir"
Write-Host "  停止服务：.\scripts\run-passenger-to-admin.ps1 -Stop"
