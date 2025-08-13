# Google Storage Dumper

A simple Go tool to download all files from a public Google Cloud Storage bucket.

## Features

- Lists and downloads all files from a specified GCS bucket.
- Handles paginated bucket listings.
- Downloads files concurrently for speed.
- Preserves directory structure.

## Install

You can install the tool using Go 1.25+:

```sh
go install github.com/topi314/google-storage-dumper@latest
```

This will place the `google-storage-dumper` binary in your `$GOPATH/bin` or `$GOBIN` directory.

## Usage

```sh
go run main.go -storage-url=<storage-url> -bucket-name=<bucket-name> [data-folder]
```

- `-storage-url`: Base URL for Google Cloud Storage (default: `https://storage.googleapis.com/`)
- `-bucket-name`: Name of the bucket to download from (default: `pokemongolive`)
- `data-folder` (optional, positional): Local directory to save files (default: `./data`)

### Example

```sh
go run main.go ./output -bucket-name=my-bucket
```

## Requirements

- Go 1.25 or newer

## Notes

- Only works with public buckets (no authentication).
- Downloads are limited to 10 concurrent files.

## License

This project is licensed under the [Apache License 2.0](LICENSE).