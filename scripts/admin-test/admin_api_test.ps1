# Automated HTTP test runner for the admin module (api/admin).
# Covers auth, operation logs, users, driver certifications, orders, coupons,
# coupon issue tasks, promotion activities, statistics, export tasks, blacklist
# and risk hit records, following the admin API doc under docs/api/.
#
# Usage:
#   .\scripts\admin-test\admin_api_test.ps1
#   .\scripts\admin-test\admin_api_test.ps1 -WriteOps   # also run positive write tests
#
# Positive write tests create test records named AUTOTEST_<timestamp> and then
# disable / release them so the database stays in a sane state.

param(
    [string]$BaseUrl = "http://127.0.0.1:8717",
    [string]$Username = "admin",
    [string]$Password = "123456",
    [switch]$WriteOps,
    [string]$ReportPath = "",
    [int]$TimeoutSec = 20
)

$ErrorActionPreference = "Stop"
$script:Timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$script:Results = New-Object System.Collections.Generic.List[object]
$script:Token = $null
$script:CouponId = 0
$script:ActivityId = 0
$script:BlacklistId = 0
$script:OrdersvcMode = "unknown"
$modeFile = Join-Path (Split-Path -Parent (Split-Path -Parent $PSScriptRoot)) ".gotmp\admin-test-logs\ordersvc-mode.txt"
if (Test-Path -LiteralPath $modeFile) {
    $script:OrdersvcMode = (Get-Content -LiteralPath $modeFile -Raw).Trim()
}

if ($ReportPath -eq "") {
    $logDir = Join-Path (Split-Path -Parent (Split-Path -Parent $PSScriptRoot)) ".gotmp\admin-test-logs"
    New-Item -ItemType Directory -Force -Path $logDir | Out-Null
    $ReportPath = Join-Path $logDir "admin-api-report-$script:Timestamp.json"
}

function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Path,
        [hashtable]$Body = $null,
        [string]$Token = $null
    )
    $headers = @{ "Content-Type" = "application/json" }
    if ($Token) { $headers["Authorization"] = "Bearer $Token" }
    $uri = $BaseUrl.TrimEnd('/') + $Path
    $params = @{
        Uri = $uri
        Method = $Method
        Headers = $headers
        UseBasicParsing = $true
        TimeoutSec = $TimeoutSec
    }
    if ($null -ne $Body) {
        $params["Body"] = ($Body | ConvertTo-Json -Compress -Depth 8)
    }
    try {
        $resp = Invoke-WebRequest @params
        $status = [int]$resp.StatusCode
        $content = $resp.Content
    } catch {
        $status = 0
        if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
            $status = [int]$_.Exception.Response.StatusCode
        }
        $content = ""
        if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
            $content = $_.ErrorDetails.Message
        }
    }
    $json = $null
    if ($content) {
        try { $json = $content | ConvertFrom-Json } catch { $json = $null }
    }
    return @{ Status = $status; Content = $content; Json = $json }
}

function Test-StatusMatch {
    param($Actual, $Expected)
    if ($Expected -is [array]) { return ($Expected -contains $Actual) }
    if ($Expected -eq "4xx") { return ($Actual -ge 400 -and $Actual -lt 500) }
    if ($Expected -eq "5xx") { return ($Actual -ge 500 -and $Actual -lt 600) }
    return ($Actual -eq $Expected)
}

function Run-Case {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Path,
        [hashtable]$Body = $null,
        [bool]$Auth = $true,
        $ExpectStatus = 200,
        $ExpectCode = $null,
        [string]$Note = ""
    )
    $token = $null
    if ($Auth) { $token = $script:Token }
    $resp = Invoke-ApiRequest -Method $Method -Path $Path -Body $Body -Token $token

    $actualCode = $null
    if ($resp.Json -and $null -ne $resp.Json.code) { $actualCode = [int]$resp.Json.code }

    $statusOk = Test-StatusMatch -Actual $resp.Status -Expected $ExpectStatus
    $codeOk = $true
    if ($null -ne $ExpectCode -and $null -ne $actualCode) {
        $codeOk = ($actualCode -eq $ExpectCode)
    } elseif ($null -ne $ExpectCode -and $null -eq $actualCode) {
        $codeOk = $false
    }
    $pass = ($statusOk -and $codeOk)

    $displayBody = $resp.Content
    if ($displayBody.Length -gt 300) { $displayBody = $displayBody.Substring(0, 300) + "..." }

    $case = [PSCustomObject]@{
        Name = $Name
        Method = $Method
        Path = $Path
        Status = $resp.Status
        ExpectStatus = ($ExpectStatus -join "/")
        Code = $actualCode
        ExpectCode = $ExpectCode
        Pass = $pass
        Note = $Note
        Body = $displayBody
    }
    $script:Results.Add($case)

    $mark = "PASS"
    if (-not $pass) { $mark = "FAIL" }
    Write-Host ("[{0}] {1}  {2} {3} -> HTTP {4} code {5}" -f $mark, $Name, $Method, $Path, $resp.Status, $actualCode)
    if (-not $pass) {
        Write-Host ("      expected HTTP {0} code {1}; body: {2}" -f ($ExpectStatus -join "/"), $ExpectCode, $displayBody)
    }
    return $resp
}

function Find-ListedId {
    param([string]$Path, [string]$Key, [string]$Value)
    $resp = Invoke-ApiRequest -Method "GET" -Path $Path -Token $script:Token
    if ($resp.Status -ne 200 -or -not $resp.Json -or -not $resp.Json.data) { return 0 }
    foreach ($item in $resp.Json.data.list) {
        if ($item.$Key -eq $Value) { return [int64]$item.id }
    }
    return 0
}

Write-Host "Preflight: checking $BaseUrl/healthz ..."
$pre = Invoke-ApiRequest -Method "GET" -Path "/healthz"
if ($pre.Status -ne 200) {
    Write-Host "ERROR: api/admin is not reachable. Start the stack first:"
    Write-Host "  .\scripts\admin-test\start_admin_stack.ps1"
    exit 2
}
Write-Host "api/admin reachable."

# ---------- auth ----------
Run-Case -Name "healthz" -Method GET -Path "/healthz" -Auth $false -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "root info" -Method GET -Path "/" -Auth $false -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "register invalid body" -Method POST -Path "/admin/v1/auth/register" -Body @{} -Auth $false -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "login empty body" -Method POST -Path "/admin/v1/auth/login" -Body @{} -Auth $false -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "login wrong password" -Method POST -Path "/admin/v1/auth/login" -Body @{ username = $Username; password = "wrong-password" } -Auth $false -ExpectStatus 401 -ExpectCode 40004
Run-Case -Name "users without token" -Method GET -Path "/admin/v1/users" -Auth $false -ExpectStatus 401 -ExpectCode 40004
Run-Case -Name "register wrong method" -Method GET -Path "/admin/v1/auth/register" -Auth $false -ExpectStatus 405 -ExpectCode 40001

$login = Invoke-ApiRequest -Method "POST" -Path "/admin/v1/auth/login" -Body @{ username = $Username; password = $Password }
if ($login.Status -eq 200 -and $login.Json -and $login.Json.code -eq 0 -and $login.Json.data.token) {
    $script:Token = [string]$login.Json.data.token
    Write-Host ("LOGIN OK for {0}, token length={1}" -f $Username, $script:Token.Length)
} else {
    Write-Host "LOGIN FAILED - trying to register the first admin account..."
    $reg = Invoke-ApiRequest -Method "POST" -Path "/admin/v1/auth/register" -Body @{
        username = $Username
        password = $Password
        real_name = "System Admin"
        role = 1
    }
    if ($reg.Status -eq 200 -and $reg.Json -and $reg.Json.code -eq 0 -and $reg.Json.data.token) {
        $script:Token = [string]$reg.Json.data.token
        Write-Host ("FIRST ADMIN REGISTERED, token length={0}" -f $script:Token.Length)
    } else {
        Write-Host "REGISTER FAILED - cannot continue. Check admin credentials or seed data."
        Write-Host ("  register response: HTTP {0} body {1}" -f $reg.Status, $reg.Content)
        exit 1
    }
}

Run-Case -Name "me" -Method GET -Path "/admin/v1/auth/me" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "menus" -Method GET -Path "/admin/v1/menus" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "operation logs" -Method GET -Path "/admin/v1/operation-logs" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "operation logs filters" -Method GET -Path "/admin/v1/operation-logs?page=1&page_size=10&module=auth" -ExpectStatus 200 -ExpectCode 0

# ---------- users ----------
Run-Case -Name "user list" -Method GET -Path "/admin/v1/users" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "user list filters" -Method GET -Path "/admin/v1/users?page=1&page_size=1000" -ExpectStatus 200 -ExpectCode 0 -Note "page_size should be clamped to 100"
Run-Case -Name "user detail not found" -Method GET -Path "/admin/v1/users/999999" -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "user freeze not found" -Method POST -Path "/admin/v1/users/999999/freeze" -Body @{ reason = "test"; remark = "test" } -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "user unfreeze not found" -Method POST -Path "/admin/v1/users/999999/unfreeze" -Body @{ reason = "test"; remark = "test" } -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "user invalid path id" -Method GET -Path "/admin/v1/users/abc" -ExpectStatus 400 -ExpectCode 40001

# ---------- driver certifications ----------
Run-Case -Name "certification list" -Method GET -Path "/admin/v1/driver-certifications" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "certification list filters" -Method GET -Path "/admin/v1/driver-certifications?audit_status=1&page=1&page_size=10" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "certification detail not found" -Method GET -Path "/admin/v1/driver-certifications/999999" -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "certification approve not found" -Method POST -Path "/admin/v1/driver-certifications/999999/approve" -Body @{ remark = "test" } -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "certification reject not found" -Method POST -Path "/admin/v1/driver-certifications/999999/reject" -Body @{ remark = "test" } -ExpectStatus 404 -ExpectCode 40401

# ---------- orders ----------
Run-Case -Name "order list" -Method GET -Path "/admin/v1/orders" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "order list filters" -Method GET -Path "/admin/v1/orders?status=1&page=1&page_size=5" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "order detail not found" -Method GET -Path "/admin/v1/orders/999999" -ExpectStatus 404 -ExpectCode 40401

# If there is real order data, run a positive detail test (read-only).
$orderListResp = Invoke-ApiRequest -Method "GET" -Path "/admin/v1/orders?page=1&page_size=1" -Token $script:Token
if ($orderListResp.Json -and $orderListResp.Json.data -and $orderListResp.Json.data.list -and $orderListResp.Json.data.list.Count -gt 0) {
    $realOrderId = [int64]$orderListResp.Json.data.list[0].id
    Run-Case -Name "order detail real id=$realOrderId" -Method GET -Path "/admin/v1/orders/$realOrderId" -ExpectStatus 200 -ExpectCode 0
} else {
    Write-Host "NOTE: no order rows found, skipping positive order detail case"
}

Run-Case -Name "abnormal orders" -Method GET -Path "/admin/v1/orders/abnormal" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "abnormal orders cancel type" -Method GET -Path "/admin/v1/orders/abnormal?abnormal_type=cancel" -ExpectStatus 200 -ExpectCode 0
if ($script:OrdersvcMode -eq "real") {
    Run-Case -Name "order cancel not found" -Method POST -Path "/admin/v1/orders/999999/cancel" -Body @{ reason = "autotest"; request_id = "autotest-cancel-$script:Timestamp" } -ExpectStatus "4xx" -Note "真实 ordersvc 跨服务错误映射验证"
} else {
    Write-Host "NOTE: ordersvc mode=$script:OrdersvcMode; skipping order write-flow assertion because dummy/unknown cannot prove real closure"
}

# ---------- coupons ----------
Run-Case -Name "coupon list" -Method GET -Path "/admin/v1/coupons" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "coupon list filters" -Method GET -Path "/admin/v1/coupons?type=3&status=1" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "coupon create invalid body" -Method POST -Path "/admin/v1/coupons" -Body @{} -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "coupon update not found" -Method PUT -Path "/admin/v1/coupons/999999" -Body @{
    name = "not-exist"; type = 1; face_value = "1.00"; discount = "1.00"; threshold_amount = "0.00";
    total_count = 10; per_user_limit = 1; valid_start_at = "2026-08-18 00:00:00"; valid_end_at = "2026-12-31 23:59:59"; status = 1
} -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "coupon disable not found" -Method POST -Path "/admin/v1/coupons/999999/disable" -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "coupon issue not found" -Method POST -Path "/admin/v1/coupons/999999/issue" -Body @{ target_type = "user"; target_config = '{"user_ids":[999999]}' } -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "coupon issue invalid config" -Method POST -Path "/admin/v1/coupons/999999/issue" -Body @{ target_type = "user"; target_config = "not-json" } -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "coupon issue tasks" -Method GET -Path "/admin/v1/coupon-issue-tasks" -ExpectStatus 200 -ExpectCode 0

# ---------- promotion activities ----------
Run-Case -Name "activity list" -Method GET -Path "/admin/v1/promotion-activities" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "activity create invalid body" -Method POST -Path "/admin/v1/promotion-activities" -Body @{} -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "activity update not found" -Method PUT -Path "/admin/v1/promotion-activities/999999" -Body @{
    name = "not-exist"; type = 1; config = '{"city_code":"110000"}';
    start_at = "2026-08-18 00:00:00"; end_at = "2026-12-31 23:59:59"; status = 1
} -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "activity publish not found" -Method POST -Path "/admin/v1/promotion-activities/999999/publish" -Body @{} -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "activity rollback not found" -Method POST -Path "/admin/v1/promotion-activities/999999/rollback" -Body @{} -ExpectStatus 404 -ExpectCode 40401

# ---------- statistics / export / risk ----------
Run-Case -Name "statistics overview" -Method GET -Path "/admin/v1/statistics/overview" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "statistics orders" -Method GET -Path "/admin/v1/statistics/orders" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "statistics users" -Method GET -Path "/admin/v1/statistics/users" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "statistics coupons" -Method GET -Path "/admin/v1/statistics/coupons" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "export tasks list" -Method GET -Path "/admin/v1/export-tasks" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "export create invalid body" -Method POST -Path "/admin/v1/export-tasks" -Body @{} -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "blacklist list" -Method GET -Path "/admin/v1/blacklist" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "blacklist add invalid body" -Method POST -Path "/admin/v1/blacklist" -Body @{} -ExpectStatus 400 -ExpectCode 40001
Run-Case -Name "blacklist release not found" -Method POST -Path "/admin/v1/blacklist/999999/release" -Body @{ reason = "autotest" } -ExpectStatus 404 -ExpectCode 40401
Run-Case -Name "risk hit records" -Method GET -Path "/admin/v1/risk/hit-records" -ExpectStatus 200 -ExpectCode 0

# ---------- positive write tests (optional) ----------
if ($WriteOps) {
    Write-Host ""
    Write-Host "WriteOps enabled - running positive write tests..."

    # coupon: create -> issue -> update -> disable
    $couponName = "AUTOTEST_$script:Timestamp"
    Run-Case -Name "coupon create" -Method POST -Path "/admin/v1/coupons" -Body @{
        name = $couponName; type = 3; face_value = "8.00"; discount = "1.00"; threshold_amount = "0.00";
        total_count = 100; per_user_limit = 1; valid_start_at = "2026-08-18 00:00:00"; valid_end_at = "2026-12-31 23:59:59"; status = 1
    } -ExpectStatus 200 -ExpectCode 0
    $script:CouponId = Find-ListedId -Path "/admin/v1/coupons?keyword=$couponName" -Key "name" -Value $couponName
    Write-Host "  created coupon id=$script:CouponId"
    if ($script:CouponId -gt 0) {
        Run-Case -Name "coupon issue" -Method POST -Path "/admin/v1/coupons/$script:CouponId/issue" -Body @{ target_type = "user"; target_config = '{"user_ids":[999999]}' } -ExpectStatus 200 -ExpectCode 0
        Run-Case -Name "coupon update" -Method PUT -Path "/admin/v1/coupons/$script:CouponId" -Body @{
            name = "$couponName-updated"; type = 3; face_value = "8.00"; discount = "1.00"; threshold_amount = "0.00";
            total_count = 100; per_user_limit = 1; valid_start_at = "2026-08-18 00:00:00"; valid_end_at = "2026-12-31 23:59:59"; status = 1
        } -ExpectStatus 200 -ExpectCode 0
        Run-Case -Name "coupon disable" -Method POST -Path "/admin/v1/coupons/$script:CouponId/disable" -ExpectStatus 200 -ExpectCode 0
    } else {
        Write-Host "  WARN: could not locate created coupon, skipping coupon write flow"
    }

    # promotion activity: create -> update -> publish -> rollback
    $activityName = "AUTOTEST_ACT_$script:Timestamp"
    Run-Case -Name "activity create" -Method POST -Path "/admin/v1/promotion-activities" -Body @{
        name = $activityName; type = 3; config = '{"city_code":"110000","discount":"8.00"}';
        start_at = "2026-08-18 00:00:00"; end_at = "2026-12-31 23:59:59"; status = 1
    } -ExpectStatus 200 -ExpectCode 0
    $script:ActivityId = Find-ListedId -Path "/admin/v1/promotion-activities?keyword=$activityName" -Key "name" -Value $activityName
    Write-Host "  created activity id=$script:ActivityId"
    if ($script:ActivityId -gt 0) {
        Run-Case -Name "activity update" -Method PUT -Path "/admin/v1/promotion-activities/$script:ActivityId" -Body @{
            name = "$activityName-updated"; type = 3; config = '{"city_code":"110000","discount":"7.00"}';
            start_at = "2026-08-18 00:00:00"; end_at = "2026-12-31 23:59:59"; status = 1
        } -ExpectStatus 200 -ExpectCode 0
        Run-Case -Name "activity publish" -Method POST -Path "/admin/v1/promotion-activities/$script:ActivityId/publish" -Body @{ publish_scope = "all"; target_config = '{}' } -ExpectStatus 200 -ExpectCode 0
        Run-Case -Name "activity rollback" -Method POST -Path "/admin/v1/promotion-activities/$script:ActivityId/rollback" -Body @{ target_config = '{}' } -ExpectStatus 200 -ExpectCode 0
    } else {
        Write-Host "  WARN: could not locate created activity, skipping activity write flow"
    }

    # blacklist: add -> release
    $testTargetId = 9999998
    Run-Case -Name "blacklist add" -Method POST -Path "/admin/v1/blacklist" -Body @{ target_type = "user"; target_id = $testTargetId; reason = "autotest" } -ExpectStatus 200 -ExpectCode 0
    $script:BlacklistId = Find-ListedId -Path "/admin/v1/blacklist?target_type=user&target_id=$testTargetId" -Key "target_id" -Value ([string]$testTargetId)
    Write-Host "  created blacklist id=$script:BlacklistId"
    if ($script:BlacklistId -gt 0) {
        Run-Case -Name "blacklist release" -Method POST -Path "/admin/v1/blacklist/$script:BlacklistId/release" -Body @{ reason = "autotest release" } -ExpectStatus 200 -ExpectCode 0
    } else {
        Write-Host "  WARN: could not locate created blacklist, skipping release"
    }

    # export task
    Run-Case -Name "export create" -Method POST -Path "/admin/v1/export-tasks" -Body @{ export_type = "orders"; filters = '{"start_time":"2026-08-01 00:00:00"}' } -ExpectStatus 200 -ExpectCode 0
}

# ---------- logout ----------
Run-Case -Name "logout" -Method POST -Path "/admin/v1/auth/logout" -ExpectStatus 200 -ExpectCode 0
Run-Case -Name "me after logout" -Method GET -Path "/admin/v1/auth/me" -ExpectStatus 401 -ExpectCode 40004

# ---------- report ----------
$passCount = @($script:Results | Where-Object { $_.Pass }).Count
$failCount = @($script:Results | Where-Object { -not $_.Pass }).Count
$totalCount = $script:Results.Count

Write-Host ""
Write-Host ("SUMMARY: {0} total, {1} passed, {2} failed" -f $totalCount, $passCount, $failCount)

$report = [PSCustomObject]@{
    generated_at = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
    base_url = $BaseUrl
    username = $Username
    write_ops = [bool]$WriteOps
    ordersvc_mode = $script:OrdersvcMode
    total = $totalCount
    passed = $passCount
    failed = $failCount
    results = $script:Results
}
$report | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $ReportPath -Encoding UTF8
Write-Host "Report saved: $ReportPath"

if ($failCount -gt 0) { exit 1 }
exit 0
