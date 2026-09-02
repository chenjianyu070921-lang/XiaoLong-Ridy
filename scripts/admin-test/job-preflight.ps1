param(
    [string]$ConfigPath = "job/etc/job.yaml"
)

# job 依赖预检：只检查配置字段和目标地址，不输出凭据、不写入业务系统。
$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$config = Join-Path $root $ConfigPath
if (-not (Test-Path -LiteralPath $config)) {
    Write-Error "找不到 job 配置文件: $ConfigPath"
    exit 1
}

$content = Get-Content -Raw -LiteralPath $config
$required = @("mysql:", "redis:", "kafka:", "orderrpc:", "dispatchrpc:", "driverrpc:", "pushrpc:")
$missing = @($required | Where-Object { $content -notmatch [regex]::Escape($_) })
if ($missing.Count -gt 0) {
    Write-Error ("job 配置缺少必要段: " + ($missing -join ", "))
    exit 1
}

# 仅验证配置段存在；不解析和打印 DSN、密码或 Kafka 认证信息。
Write-Output "job 配置预检通过: $ConfigPath"
Write-Output "下一步可在授权环境执行 job 启动预检，或运行 go test ./job/..."
exit 0
