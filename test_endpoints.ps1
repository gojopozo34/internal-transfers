# Test endpoints for internal-transfers API

$baseUrl = "http://localhost:8080"

# Colors for output
$Green = "`e[32m"
$Red = "`e[31m"  
$Yellow = "`e[33m"
$Reset = "`e[0m"

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Uri,
        [string]$Body
    )
    
    Write-Host "`n$Yellow >> Testing: $Name$Reset"
    try {
        if ($Body) {
            $response = Invoke-WebRequest -Uri $Uri -Method $Method -Headers @{"Content-Type"="application/json"} -Body $Body -UseBasicParsing -ErrorAction Stop
        } else {
            $response = Invoke-WebRequest -Uri $Uri -Method $Method -UseBasicParsing -ErrorAction Stop
        }
        Write-Host "$Green✓ Status: $($response.StatusCode)$Reset"
        Write-Host "$Green Response: $($response.Content)$Reset"
        return $response
    }
    catch {
        Write-Host "$Red✗ Error: $($_.Exception.Message)$Reset"
        if ($_.Exception.Response) {
            Write-Host "$Red Response: $($_.Exception.Response.StatusCode)$Reset"
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            Write-Host "$Red Body: $($reader.ReadToEnd())$Reset"
        }
        return $null
    }
}

# Test 1: Health check
Test-Endpoint -Name "Health Check" -Method Get -Uri "$baseUrl/health"

# Test 2: Create account 1
Test-Endpoint -Name "Create Account 1" -Method Post -Uri "$baseUrl/accounts" `
    -Body '{"id":"acc001","initial_balance":1000}'

# Test 3: Create account 2
Test-Endpoint -Name "Create Account 2" -Method Post -Uri "$baseUrl/accounts" `
    -Body '{"id":"acc002","initial_balance":500}'

# Test 4: Get account 1
Test-Endpoint -Name "Get Account 1" -Method Get -Uri "$baseUrl/accounts/acc001"

# Test 5: Get account 2
Test-Endpoint -Name "Get Account 2" -Method Get -Uri "$baseUrl/accounts/acc002"

# Test 6: Transfer money
Test-Endpoint -Name "Transfer 250 from acc001 to acc002" -Method Post -Uri "$baseUrl/transactions" `
    -Body '{"source_account_id":"acc001","destination_account_id":"acc002","amount":250}'

# Test 7: Verify acc001 balance (should be 750)
Test-Endpoint -Name "Verify Account 1 Balance" -Method Get -Uri "$baseUrl/accounts/acc001"

# Test 8: Verify acc002 balance (should be 750)
Test-Endpoint -Name "Verify Account 2 Balance" -Method Get -Uri "$baseUrl/accounts/acc002"

# Test 9: Try duplicate account (should fail)
Test-Endpoint -Name "Try Duplicate Account (should fail)" -Method Post -Uri "$baseUrl/accounts" `
    -Body '{"id":"acc001","initial_balance":500}'

# Test 10: Try insufficient balance transfer (should fail)
Test-Endpoint -Name "Try Insufficient Balance (should fail)" -Method Post -Uri "$baseUrl/transactions" `
    -Body '{"source_account_id":"acc001","destination_account_id":"acc002","amount":10000}'

# Test 11: Try same source/dest (should fail)
Test-Endpoint -Name "Try Same Source/Destination (should fail)" -Method Post -Uri "$baseUrl/transactions" `
    -Body '{"source_account_id":"acc001","destination_account_id":"acc001","amount":100}'

Write-Host "`n$Green=== All tests completed ===$Reset"
