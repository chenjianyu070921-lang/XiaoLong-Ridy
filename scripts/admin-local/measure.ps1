# 管理后台最小服务集内存采集脚本。
#
# 本脚本仅读取 5173、8717、8084 端口对应进程的内存指标，不会修改服务、配置、数据库或业务数据。

$ErrorActionPreference = "Stop"

# Get-ProcessMemoryByPort 按监听端口定位进程，并返回工作集和私有内存的 MB 数值。
function Get-ProcessMemoryByPort {
    param([Parameter(Mandatory = $true)][int]$Port)

    $connection = Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($null -eq $connection) {
        return [PSCustomObject]@{
            Port = $Port
            Status = "未监听"
            PID = $null
            ProcessName = $null
            WorkingSetMB = $null
            PrivateMemoryMB = $null
        }
    }

    $process = Get-Process -Id $connection.OwningProcess -ErrorAction Stop
    return [PSCustomObject]@{
        Port = $Port
        Status = "运行中"
        PID = $process.Id
        ProcessName = $process.ProcessName
        WorkingSetMB = [math]::Round($process.WorkingSet64 / 1MB, 1)
        PrivateMemoryMB = [math]::Round($process.PrivateMemorySize64 / 1MB, 1)
    }
}

@(5173, 8717, 8084) | ForEach-Object { Get-ProcessMemoryByPort -Port $_ } | Format-Table -AutoSize

