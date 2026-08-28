# XiaoLong-Ridy 一键启动全部服务（Windows / PowerShell）
# 前置：
#   1. 已安装 Go 1.21+ 且 go 在 PATH
#   2. 先启动中间件： docker compose -f deploy/docker/infra.yml up -d  (MySQL/Redis/Kafka/Etcd)
#   3. 可选：设置高德 Key 环境变量后 locationsvc 才会真正调用高德
#        $env:AMAP_API_KEY="你的高德Key"
$root = (Get-Item $PSScriptRoot).Parent.FullName
Set-Location $root
$logs = Join-Path $root "logs"
New-Item -ItemType Directory -Force -Path $logs | Out-Null

$services = @(
    @{dir = "rpc/usersvc";              yaml = "usersvc" },
    @{dir = "rpc/pricesvc";             yaml = "pricesvc" },
    @{dir = "rpc/paysvc";               yaml = "paysvc" },
    @{dir = "rpc/locationsvc";          yaml = "locationsvc" },
    @{dir = "rpc/pushesvc";             yaml = "pushesvc" },
    @{dir = "rpc/ordersvc";             yaml = "ordersvc" },
    @{dir = "rpc/dispatchsvc";          yaml = "dispatchsvc" },
    @{dir = "mq-consumer/location-consumer"; yaml = "location-consumer" },
    @{dir = "job";                      yaml = "job" }
)

foreach ($s in $services) {
    $logOut = Join-Path $logs "$($s.yaml).log"
    $logErr = Join-Path $logs "$($s.yaml).err"
    Write-Host "=> 启动 $($s.yaml) (日志: $($s.yaml).log)"
    Start-Process -FilePath "go" `
        -ArgumentList @("run", "./$($s.dir)", "-f", "./$($s.dir)/etc/$($s.yaml).yaml") `
        -RedirectStandardOutput $logOut `
        -RedirectStandardError $logErr `
        -WindowStyle Hidden
}

Write-Host "已全部提交后台启动，日志统一在 logs/ 目录。停止请关闭对应 go 进程。"
