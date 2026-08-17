# 支付模块真实联调脚本：启动 paysvc 并跑通支付正向全链路
#
# 前置条件：
#   1. MySQL 已运行，且已执行 scripts/sql/migrate/05_trade_module.sql（建表 + 初始计价规则）
#      本地可用 docker 起 MySQL：
#        docker compose -f deploy/docker/infra.yml up -d mysql
#   2. 确认 rpc/paysvc/etc/paysvc.yaml 里 mysql.dsn 的账号密码与本地一致
#      （infra.yml 默认 root 密码为 root，paysvc.yaml 默认为 root123，需二选一对齐）
#   3. 已安装 Go
#
# 用法：
#   powershell -ExecutionPolicy Bypass -File scripts/e2e/run_pay_e2e.ps1
#   可选参数：-Target 127.0.0.1:50054

param(
    [string]$Target = "127.0.0.1:50054"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)   # 项目根目录

Write-Host "== 1/3 启动 paysvc ==" -ForegroundColor Cyan
$outLog = Join-Path $env:TEMP "paysvc.out.log"
$errLog = Join-Path $env:TEMP "paysvc.err.log"
$paysvc = Start-Process -FilePath "go" `
    -ArgumentList "run", (Join-Path $root "rpc\paysvc\paysvc.go"), "-f", (Join-Path $root "rpc\paysvc\etc\paysvc.yaml") `
    -PassThru -NoNewWindow -RedirectStandardOutput $outLog -RedirectStandardError $errLog
Write-Host "paysvc PID=$($paysvc.Id)，等待监听 $Target ..."

$ready = $false
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 1
    $conn = Test-NetConnection -ComputerName 127.0.0.1 -Port 50054 -WarningAction SilentlyContinue
    if ($conn.TcpTestSucceeded) { $ready = $true; break }
}
if (-not $ready) {
    Write-Host "paysvc 启动超时，错误日志如下：" -ForegroundColor Red
    Get-Content $errLog -ErrorAction SilentlyContinue
    exit 1
}
Write-Host "paysvc 已就绪" -ForegroundColor Green

Write-Host "`n== 2/3 运行联调客户端 ==" -ForegroundColor Cyan
try {
    go run (Join-Path $root "scripts\e2e\pay_e2e_client.go") -target $Target
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} finally {
    Write-Host "`n== 3/3 停止 paysvc ==" -ForegroundColor Cyan
    Stop-Process -Id $paysvc.Id -Force -ErrorAction SilentlyContinue
    Write-Host "完成"
}
