$ErrorActionPreference = "Continue"
Set-Location D:\gocode\src\XiaoLong-Ridy
$enc = New-Object System.Text.UTF8Encoding($false)
$base = "http://127.0.0.1:8082/api/driver/v1"
# 1. 登录
[System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\sms.json", '{"phone":"19397622796"}', $enc)
curl.exe -s -X POST "$base/auth/send-sms-code" -H "Content-Type: application/json" --data "@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\sms.json" | Out-Null
Start-Sleep -Seconds 2
$code = go run .tmp_diag\getcode.go
[System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_l.json", '{"phone":"19397622796","code":"' + $code + '"}' , [System.Text.Encoding]::ASCII)
$resp = curl.exe -s -X POST "$base/auth/login-by-sms" -H "Content-Type: application/json" --data "@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_l.json"
$j = $resp | ConvertFrom-Json
if ($j.code -ne 0 -or -not $j.data.token) { Write-Output "登录失败: $resp"; exit 1 }
$tok = $j.data.token
Write-Output "登录OK"
# 2. 司机8 上线到订单33 起点
[System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_b.json", '{"deviceId":"verify-01","longitude":118.283231,"latitude":34.020904}', $enc)
$r1 = curl.exe -s -X POST "$base/drivers/online" -H "Content-Type: application/json" -H "Authorization: Bearer $tok" --data "@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_b.json"
Write-Output "上线: $r1"
# 3. 发布 order.created 事件（order 33 补派）
Write-Output "== 发布 order.created(order 33) =="
go run .tmp_diag\kpub2.go
Write-Output "等待 15s 让消费+派单完成..."
Start-Sleep -Seconds 15
# 4. 查 dispatch_record
Write-Output "== dispatch_record(order 33) =="
go run .tmp_diag\dr33.go
# 5. 查 Redis driver:available:8
Write-Output "== driver:available:8 =="
go run .tmp_diag\chkav.go
# 6. 司机端可接单接口
[System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_b.json", '{"longitude":118.283231,"latitude":34.020904}', $enc)
$r2 = curl.exe -s -X POST "$base/orders/available" -H "Content-Type: application/json" -H "Authorization: Bearer $tok" --data "@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_b.json"
Write-Output "orders/available: $r2"
