param (
    [ValidateSet("up", "down", "logs", "rebuild")]
    [string]$Command,

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
    foreach ($image in $DockerImages) {
        Write-Host "Getting the logs from $image..."
        docker logs --tail 5 $image
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
    } else {
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
