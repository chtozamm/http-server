param (
    [ValidateSet("up", "down", "logs", "rebuild")]
    [string]$Command,

    [ValidateSet("http-server", "http-server-postgres", "all")]
    [string]$ImageName
)

$ComposeFile = "docker/compose.yaml"
$DockerImages = @("http-server", "http-server-postgres")

function up {
    Write-Host "Starting services..."
    docker compose -f "$ComposeFile" up -d
}

function down {
    Write-Host "Stopping services..."
    docker compose -f "$ComposeFile" down
}

function logs {
    if (-not $ImageName) {
        Write-Error "No image name provided for getting logs."
        Write-Host "Usage: .\$PSCommandName logs <IMAGE_NAME>"
        exit 1
    }
    
    if ($ImageName -eq "all") {
        foreach ($image in $DockerImages) {
            Write-Host "`n> Getting the logs from $image...`n" -ForegroundColor Green
            docker logs $image
            Write-Host ""
        }
        return
    }

    if ($DockerImages -contains $ImageName) {
        Write-Host "`n> Getting the logs from $ImageName...`n" -ForegroundColor Green
        docker logs $ImageName
        Write-Host ""
    }
    else {
        Write-Error "Invalid image name: $ImageName. Valid options are: $($DockerImages -join ', ')"
    }
}

function rebuild {
    if (-not $ImageName) {
        Write-Error "No image name provided for rebuild."
        Write-Host "Usage: .\$PSCommandName rebuild <IMAGE_NAME>"
        exit 1
    }
    
    if ($ImageName -eq "all") {
        Write-Host "Rebuilding all images..."
        down
        foreach ($image in $DockerImages) {
            docker image rm $image
        }
        up
        return
    }

    if ($DockerImages -contains $ImageName) {
        Write-Host "Rebuilding image $ImageName..."
        down
        docker image rm $ImageName
        up
    }
    else {
        Write-Error "Invalid image name: $ImageName. Valid options are: $($DockerImages -join ', ')"
    }
}


if (-not $Command) {
    Write-Error "No arguments provided."
    Write-Host "Usage: .\$PSCommandName <COMMAND>"
    Write-Host "Valid commands: up, down, logs, rebuild"
    exit 1
}

switch ($Command) {
    "up" {
        up
    }
    "down" {
        down
    }
    "logs" {
        logs
    }
    "rebuild" {
        rebuild
    }
}
