$baseUrl = "http://localhost:8080"

Write-Host "Test 1: Health Check"
Invoke-WebRequest -Uri "$baseUrl/health" -Method Get -UseBasicParsing | Select-Object StatusCode, Content

Write-Host "`nTest 2: Create Account 1"
$body = @{id="acc001"; initial_balance=1000} | ConvertTo-Json
$r = Invoke-WebRequest -Uri "$baseUrl/accounts" -Method Post -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 3: Create Account 2"
$body = @{id="acc002"; initial_balance=500} | ConvertTo-Json
$r = Invoke-WebRequest -Uri "$baseUrl/accounts" -Method Post -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 4: Get Account 1"
$r = Invoke-WebRequest -Uri "$baseUrl/accounts/acc001" -Method Get -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 5: Get Account 2"
$r = Invoke-WebRequest -Uri "$baseUrl/accounts/acc002" -Method Get -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 6: Transfer 250 from acc001 to acc002"
$body = @{source_account_id="acc001"; destination_account_id="acc002"; amount=250} | ConvertTo-Json
$r = Invoke-WebRequest -Uri "$baseUrl/transactions" -Method Post -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 7: Verify acc001 balance (should be 750)"
$r = Invoke-WebRequest -Uri "$baseUrl/accounts/acc001" -Method Get -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 8: Verify acc002 balance (should be 750)"
$r = Invoke-WebRequest -Uri "$baseUrl/accounts/acc002" -Method Get -UseBasicParsing
Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"

Write-Host "`nTest 9: Try duplicate account (should fail with 409)"
$body = @{id="acc001"; initial_balance=500} | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/accounts" -Method Post -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing -ErrorAction Stop
    Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"
} catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode), Error: $($_.Exception.Response.StatusCode.Value__)"
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    Write-Host "Body: $($reader.ReadToEnd())"
}

Write-Host "`nTest 10: Try insufficient balance transfer (should fail with 400)"
$body = @{source_account_id="acc001"; destination_account_id="acc002"; amount=10000} | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/transactions" -Method Post -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing -ErrorAction Stop
    Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"
} catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode), Error:"
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    Write-Host "Body: $($reader.ReadToEnd())"
}

Write-Host "`nTest 11: Try same source/dest (should fail with 400)"
$body = @{source_account_id="acc001"; destination_account_id="acc001"; amount=100} | ConvertTo-Json
try {
    $r = Invoke-WebRequest -Uri "$baseUrl/transactions" -Method Post -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing -ErrorAction Stop
    Write-Host "Status: $($r.StatusCode), Body: $($r.Content)"
} catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode)"
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    Write-Host "Body: $($reader.ReadToEnd())"
}

Write-Host "`n=== All tests completed ==="
