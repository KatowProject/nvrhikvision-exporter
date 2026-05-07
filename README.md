# Hikvision NVR Exporter

Prometheus exporter for Hikvision NVR devices.

## Features

- HDD SMART health, temperature, and power-on days
- Camera status and camera information
- CPU, memory, and uptime metrics
- Support for multiple NVRs in one exporter
- Configuration via YAML files or CLI flags

## Installation

### Requirements

- Go 1.26+
- Access to a Hikvision NVR API

### Clone the repository

```bash
git clone https://github.com/KatowProject/nvrhikvision-exporter.git
cd nvrhikvision-exporter
```

## Build

### Build the current platform

```bash
make build
```

On Windows, the same command works as well:

```cmd
make build
```

### Build all platforms

Linux/macOS:

```bash
bash ./scripts/build.sh 1.0.0 all
```

Windows PowerShell:

```powershell
.\scripts\build.ps1 -Version "1.0.0" -Target "all"
```

Windows CMD:

```batch
scripts\build.bat 1.0.0 all
```

### Build a specific platform

```bash
bash ./scripts/build.sh 1.0.0 linux/amd64
bash ./scripts/build.sh 1.0.0 linux/arm64
bash ./scripts/build.sh 1.0.0 darwin/arm64
```

## Usage

### YAML configuration file

Create a `config.yaml` file:

```yaml
nvrs:
  - ip: "192.168.1.2"
    username: "admin1"
    password: "cctv12345"
    name: "NVR-1"
  - ip: "192.168.1.3"
    username: "admin2"
    password: "cctv123"
    name: "NVR-2"

server:
  port: "9102"
```

Run the exporter:

```bash
make run
```

### CLI flags

Legacy single-NVR mode is still supported:

```bash
./dist/nvrhikvision-exporter -ip=192.168.1.3 -user=your-user -pass=your-pass -port=9102
```

## Metrics

### HDD metrics

- `hikvision_hdd_health_percent`
- `hikvision_hdd_temperature_celsius`
- `hikvision_hdd_power_on_days`
- `hikvision_hdd_status`
- `hikvision_hdd_smart_attribute`
- `hikvision_hdd_smart_normalized`

### Camera metrics

- `hikvision_camera_status`
- `hikvision_camera_info`

### System metrics

- `hikvision_cpu_usage_percent`
- `hikvision_memory_usage_percent`
- `hikvision_uptime_seconds`

## Development

```bash
make build
make build-all
make build-linux
make build-windows
make build-darwin
make run
make test
make fmt
make clean
```

## Docker

```bash
make docker-build
make docker-run
docker-compose -f deploy/docker-compose.yml up
```

## Running as a Service

If you want the exporter to start automatically and keep running in the background, you can install it as a service.

### Windows with NSSM

NSSM (Non-Sucking Service Manager) is a simple way to run the exporter as a Windows service.

1. Build the Windows binary:

```bat
make build-windows
```

2. Prepare a `config.yaml` file, for example:

```yaml
nvrs:
  - ip: "192.168.1.2"
    username: "admin1"
    password: "cctv12345"
    name: "NVR-1"
  - ip: "192.168.1.3"
    username: "admin2"
    password: "cctv123"
    name: "NVR-2"

server:
  port: "9102"
```


3. Install the service with NSSM:

```bat
nssm install HikvisionExporter C:\path\to\nvrhikvision-exporter.exe
nssm set HikvisionExporter AppDirectory C:\path\to
nssm set HikvisionExporter AppParameters -config=config.yaml
nssm start HikvisionExporter
```

If you want to manage the service from Windows Services, run NSSM as Administrator. You can also configure stdout/stderr logs from the NSSM GUI if needed.

### Linux with systemd

On Linux, `systemd` is the recommended way to run the exporter as a daemon.

1. Build or copy the Linux binary to a fixed location, for example `/usr/local/bin/nvrhikvision-exporter`.

2. Put your `config.yaml` in a stable directory such as `/etc/nvrhikvision-exporter/config.yaml`.

3. Create a service file, for example `/etc/systemd/system/nvrhikvision-exporter.service`:

```ini
[Unit]
Description=Hikvision NVR Exporter
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nvrhikvision-exporter -config=/etc/nvrhikvision-exporter/config.yaml
WorkingDirectory=/etc/nvrhikvision-exporter
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

4. Reload `systemd`, enable the service, and start it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now nvrhikvision-exporter
sudo systemctl status nvrhikvision-exporter
```

5. View logs if needed:

```bash
journalctl -u nvrhikvision-exporter -f
```

## Troubleshooting

- Make sure the NVR is reachable from the exporter host
- Verify the username and password
- Check the logs if metrics do not appear

## License

MIT
