# CSV Converter for Surveyor Data

A utility for converting CSV files with custom formatting, specifically designed for surveyor data. This tool can be used both as a command-line application and as a server that automatically processes files.

## Command Line Usage

```
csv-converter input.csv [-o output.csv] [-s separator] [-f formatString]
csv-converter -c config.json (server mode)
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-o` | Output file path | `[inputfile]-converted[ext]` |
| `-s` | Input CSV separator | `,` (comma) |
| `-f` | Format string for output | `P:14 Y:12 X:12 H:10 MC:6 DT:8` |
| `-c` | Path to config file (activates server mode) | - |

#### Notes:
- For tab-separated files, you can use `-s "\\t"` or `-s "tab"`
- The format string uses the syntax `FieldName:Width` with space separation  

## Server Mode

When using `-c` flag, the application starts in server mode and all other flags are ignored.

### Config File Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `delimiter` | CSV delimiter character | `,` |
| `port` | Server port to listen on | `8080` |
| `directory` | Directory to watch for files | - |
| `outputPattern` | Pattern for output filenames | `%s-converted` |
| `formatString` | Format specification | `P:14 Y:12 X:12 H:10 MC:6 DT:8` |
| `processedDir` | Directory to move processed files (optional) | - |
| `pollInterval` | How often to check for new files (seconds) | `30` |

#### Notes:
- Set `pollInterval` to `-1` to disable automatic polling
- The server exposes two endpoints:
  - `/webhook` - POST endpoint for triggering file processing
  - `/status` - GET endpoint to check server status

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

## Example Configuration File

```json
{
  "delimiter": ",",
  "port": 8080,
  "directory": "C:/path/to/watch/directory",
  "outputPattern": "%s-converted",
  "formatString": "P:14 Y:12 X:12 H:10 MC:6 DT:8",
  "processedDir": "C:/path/to/processed/directory",
  "pollInterval": 30
}
```

## Integration with Foldersync

Configure foldersync to call the `/webhook` endpoint after it completes syncing. The server will then process any new CSV files in the configured directory.