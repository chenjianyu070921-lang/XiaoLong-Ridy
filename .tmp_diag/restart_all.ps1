$ErrorActionPreference = "Continue"
Set-Location D:\gocode\src\XiaoLong-Ridy
$enc = New-Object System.Text.UTF8Encoding($false)
$svcs = @(
  @{n='driversvc';e='.run\bin\driversvc.exe';f='rpc\driversvc\etc\driversvc.yaml';port=50055},
  @{n='ordersvc';e='.run\bin\ordersvc.exe';f='rpc\ordersvc\etc\ordersvc.yaml';port=50051},
  @{n='dispatchsvc';e='.run\bin\dispatchsvc.exe';f='rpc\dispatchsvc\etc\dispatchsvc.yaml';port=50056},
  @{n='locationsvc';e='.run\bin\locationsvc.exe';f='rpc\locationsvc\etc\locationsvc.yaml';port=9001},
  @{n='api-driver';e='.run\bin\api-driver.exe';f='api\driver\etc\driver.yaml';port=8082},
  @{n='order-event-consumer';e='.run\bin\order-event-consumer.exe';f='mq-consumer\order-event-consumer\etc\order-event-consumer.yaml';port=0}
)
Write-Output "== 杀旧进程 =="
foreach ($name in @('driversvc','ordersvc','dispatchsvc','locationsvc','api-driver','order-event-consumer')) { Get-Process -Name $name -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue }
Start-Sleep -Seconds 2
Write-Output "== 启动服务 =="
foreach ($s in $svcs) {
  $p = Start-Process -FilePath (Join-Path (Get-Location) $s.e) -ArgumentList '-f',(Join-Path (Get-Location) $s.f) -RedirectStandardOutput (Join-Path (Get-Location) "logs\$($s.n).out.log") -RedirectStandardError (Join-Path (Get-Location) "logs\$($s.n).err.log") -WindowStyle Hidden -PassThru
  Write-Output "started $($s.n) pid=$($p.Id)"
}
Start-Sleep -Seconds 12
Write-Output "== 端口检查 =="
foreach ($s in $svcs) { if ($s.port -gt 0) { $c = Get-NetTCPConnection -LocalPort $s.port -State Listen -ErrorAction SilentlyContinue; if ($c) { Write-Output "$($s.n) $($s.port) UP" } else { Write-Output "$($s.n) $($s.port) DOWN" } } }
$consumerAlive = Get-Process -Name order-event-consumer -ErrorAction SilentlyContinue
Write-Output "consumer 存活: $(if ($consumerAlive) { 'yes' } else { 'NO' })"
Write-Output "== consumer stderr =="
if (Test-Path logs\order-event-consumer.err.log) { $e = Get-Content logs\order-event-consumer.err.log -Tail 5; if ($e) { $e } }
