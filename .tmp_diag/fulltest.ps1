$ErrorActionPreference = "Continue"
$base = "http://127.0.0.1:8082/api/driver/v1"
$enc = New-Object System.Text.UTF8Encoding($false)
function Json($o) { [System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_b.json", ($o | ConvertTo-Json -Compress), $enc) }
function Call($name, $method, $path, $body) {
  if ($body) { Json $body }
  $args = @("-s","-X",$method,"$base$path","-H","Content-Type: application/json","-H","Authorization: Bearer $global:tok")
  if ($body) { $args += "--data","@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_b.json" }
  $r = curl.exe @args
  try { $j = $r | ConvertFrom-Json; $code = $j.code; $msg = $j.message } catch { $code = "PARSE_ERR"; $msg = $r }
  $mark = if ($code -eq 0) { "OK " } else { "FAIL" }
  Write-Output ("{0}  {1,-22} code={2,-4} msg={3}" -f $mark, $name, $code, $msg)
}
# 登录
[System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\sms.json", '{"phone":"19397622796"}', $enc)
curl.exe -s -X POST "$base/auth/send-sms-code" -H "Content-Type: application/json" --data "@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\sms.json" | Out-Null
Start-Sleep -Seconds 2
$code = go run .tmp_diag\getcode.go
[System.IO.File]::WriteAllText("D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_l.json", '{"phone":"19397622796","code":"'+$code+'"}' , [System.Text.Encoding]::ASCII)
$resp = curl.exe -s -X POST "$base/auth/login-by-sms" -H "Content-Type: application/json" --data "@D:\gocode\src\XiaoLong-Ridy\.tmp_diag\_l.json"
$j = $resp | ConvertFrom-Json
if ($j.code -ne 0 -or -not $j.data.token) { Write-Output "登录失败: $resp"; exit 1 }
$global:tok = $j.data.token
Write-Output ("登录 OK token长度={0}" -f $global:tok.Length)
Call "drivers/get" GET "/drivers/get" $null
Call "drivers/ai-score" GET "/drivers/ai-score" $null
Call "drivers/certification" GET "/drivers/certification" $null
Call "vehicles/get" GET "/vehicles/get?id=7" $null
Call "drivers/online" POST "/drivers/online" @{deviceId="diag-dev-01";longitude=116.397;latitude=39.909}
Call "drivers/nearby" POST "/drivers/nearby" @{longitude=116.397;latitude=39.909;radius=5000}
Call "drivers/heartbeat" POST "/drivers/heartbeat" @{deviceId="diag-dev-01"}
Call "drivers/location/report" POST "/drivers/location/report" @{longitude=116.397;latitude=39.909}
Call "orders/available" POST "/orders/available" @{longitude=116.397;latitude=39.909}
Call "orders/list" POST "/orders/list" @{page=1;pageSize=10}
Call "orders/dispatches" POST "/orders/dispatches" @{page=1;pageSize=10}
Call "withdraws/list" POST "/withdraws/list" @{page=1;pageSize=10}
Call "income/summary" GET "/income/summary" $null
Call "income/today" GET "/income/today" $null
Call "income/week" GET "/income/week" $null
Call "income/bills" POST "/income/bills" @{page=1;pageSize=10}
Call "reviews/list" POST "/reviews/list" @{page=1;pageSize=10}
Call "drivers/offline" POST "/drivers/offline" @{deviceId="diag-dev-01"}
Write-Output "== 检测完成 =="
