# CSV Converter for Surveyor Data

A utility for converting CSV files with custom formatting, specifically designed for surveyor data. This tool can be used both as a command-line application and as a server that automatically processes files.

## Command Line Usage

```
csv-converter input.csv [-o output.csv] [-d delimiter] [-f formatString]
csv-converter -c config.json [-s]
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-o` | Output file path | `[inputfile]-converted[ext]` |
| `-d` | Input CSV delimiter | `,` (comma) |
| `-f` | Format string for output | `P:14 Y:12 X:12 H:10 MC:6 DT:8` |
| `-c` | Path to config file (for batch/server mode) | - |
| `-s` | Enable server mode with the config file | `false` |

#### Notes:
- For tab-separated files, you can use `-d "\\t"` or `-d "tab"`
- The format string uses the syntax `FieldName:Width` with space separation
- When `-c` is used without `-s`, it processes all files in the directory once
- When both `-c` and `-s` are used, it starts in server mode

## Server Mode

When using `-c` with `-s` flag, the application starts in server mode and all other flags are ignored.

### Config File Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `delimiter` | CSV delimiter character | `,` |
| `port` | Server port to listen on | `8080` |
| `directory` | Directory to watch for files | - |
| `outputPattern` | Pattern for output filenames | `%s-converted` |
| `formatString` | Format specification | `P:14 Y:12 X:12 H:10 MC:6 DT:8` |
| `processedDir` | Directory to save output files | - |
| `pollInterval` | How often to check for new files (seconds) | - |
| `originalFile` | What to do with original files ("move", "delete", or leave in place) | - |
| `certFile` | SSL certificate file for HTTPS (optional) | - |
| `keyFile` | SSL key file for HTTPS (optional) | - |

#### Notes:
- Set `pollInterval` to a negative value to disable automatic polling
- The server exposes two endpoints:
  - `/webhook` - POST endpoint for triggering file processing
  - `/status` - GET endpoint to check server status
- When `certFile` and `keyFile` are provided, the server runs in HTTPS mode

## Format String

The format string controls how each field is formatted in the output. The syntax is:
```
FieldName:Width FieldName:Width ...
```

Example: `P:14 Y:12 X:12 H:10 MC:6 DT:8`

This means:
- Field "P" will have width 14
- Field "Y" will have width 12
- Field "X" will have width 12
- Field "H" will have width 10
- Field "MC" will have width 6
- Field "DT" will have width 8

If a field exceeds its specified width, it will be replaced with "#OF#" to indicate overflow.

## Example Configuration File

```json
{
  "delimiter": ",",
  "port": 8080,
  "directory": "C:/path/to/watch/directory",
  "outputPattern": "%s-converted",
  "formatString": "P:14 Y:12 X:12 H:10 MC:6 DT:8",
  "processedDir": "C:/path/to/output/directory",
  "pollInterval": 30,
  "originalFile": "move",
  "certFile": "path/to/certificate.pem",
  "keyFile": "path/to/key.pem"
}
```

## Integration with Foldersync

Configure foldersync to call the `/webhook` endpoint after it completes syncing. The server will then process files in the configured directory.

## Security

When `certFile` and `keyFile` are provided in the configuration, the server will use HTTPS instead of HTTP, providing secure communication.



## Running as a Systemd Service

You can run the Cube to rmGEO Conversion Server as a background service that automatically starts on boot.

### Step 1: Create a systemd service file

Create a file named `cube-rmgeo.service` in the systemd directory:

```bash
sudo nano /etc/systemd/system/cube-rmgeo.service
```

Add the following content to the file:

```ini
[Unit]
Description=Cube to rmGEO Conversion Server Service
After=network.target

[Service]
Type=simple
User=yourusername
ExecStart=/path/to/project/cube-rmgeo -c config.json -s
WorkingDirectory=/path/to/project
Restart=on-failure
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=cube-rmgeo

[Install]
WantedBy=multi-user.target
```

Be sure to replace:
- `yourusername` with your actual username
- `/path/to/project` with the actual full path to your project directory

### Step 2: Enable and start the service

Reload the systemd configuration:

```bash
sudo systemctl daemon-reload
```

Enable the service to start automatically at boot:

```bash
sudo systemctl enable cube-rmgeo.service
```

Start the service immediately:

```bash
sudo systemctl start cube-rmgeo.service
```

### Step 3: Check service status

Verify that the service is running correctly:

```bash
sudo systemctl status cube-rmgeo.service
```

### Managing the service

- To stop the service: `sudo systemctl stop cube-rmgeo.service`
- To restart the service: `sudo systemctl restart cube-rmgeo.service`
- To view service logs: `sudo journalctl -u cube-rmgeo.service`