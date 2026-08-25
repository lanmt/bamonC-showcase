$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$docsRoot = Join-Path $repoRoot 'docs'
$sourceRoot = Join-Path $docsRoot 'source'
$downloadRoot = Join-Path $docsRoot 'downloads'
$archivePath = Join-Path $downloadRoot 'bamonC-anonymized.zip'

$sourceFiles = @(
    'main.go',
    'routers/initrouter.go',
    'task/corn.go',
    'task/captcha_task.go',
    'request/helper.go',
    'request/submit_reservation.go',
    'service/user_service.go',
    'model/user.go'
)

New-Item -ItemType Directory -Force -Path $sourceRoot, $downloadRoot | Out-Null
foreach ($relativePath in $sourceFiles) {
    $destination = Join-Path $sourceRoot ($relativePath + '.txt')
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $destination) | Out-Null
    Copy-Item -LiteralPath (Join-Path $repoRoot $relativePath) -Destination $destination -Force
}

$stagingRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("bamonc-portfolio-" + [guid]::NewGuid())
$packageRoot = Join-Path $stagingRoot 'bamonC-anonymized'
New-Item -ItemType Directory -Force -Path $packageRoot | Out-Null

$packageItems = @(
    'controller', 'middleware', 'model', 'request', 'routers', 'service', 'static',
    'task', 'templates', 'Dockerfile.flask', 'Dockerfile.gin', 'Dockerfile.umi-ocr',
    'docker-compose.yml', 'flask-ddddocr.py', 'go.mod', 'go.sum', 'LICENSE',
    'main.go', 'PORTFOLIO_NOTICE.txt', 'qodana.yaml'
)

foreach ($item in $packageItems) {
    Copy-Item -LiteralPath (Join-Path $repoRoot $item) -Destination $packageRoot -Recurse -Force
}

if (Test-Path -LiteralPath $archivePath) {
    Remove-Item -LiteralPath $archivePath -Force
}
Compress-Archive -Path $packageRoot -DestinationPath $archivePath -CompressionLevel Optimal
Remove-Item -LiteralPath $stagingRoot -Recurse -Force

Write-Host "Portfolio assets generated at $docsRoot"
