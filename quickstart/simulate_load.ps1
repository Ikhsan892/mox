$max = 100
Write-Host "Running $max requests..."
1..$max | ForEach-Object {
    $res = Invoke-WebRequest -Uri "http://localhost:80" -UseBasicParsing
    $body = $res.Content.Trim()
    $servedBy = $res.Headers["X-Served-By"]
    Write-Host "[$_] Handled by: $servedBy | Response: $body"
}
