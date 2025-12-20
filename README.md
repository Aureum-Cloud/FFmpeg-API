# Media Processing HTTP API (FFmpeg API)

A **HTTP API** for processing video, audio, and images using **FFmpeg**.

## What it does

* Convert media formats (video ↔ audio, MP4 → GIF, WAV → MP3)
* Resize, trim, or transcode files
* Extract audio or frames from video
* Run media workflows without local file handling

You provide input files and FFmpeg commands; the API returns the results via **S3**, **Base64**, or a **direct HTTP stream**.

## Endpoint

`POST /v1/process`

## Request Structure

```json
{
  "s3Config": { ... },
  "input": { ... },
  "commands": [ ... ],
  "output": { ... }
}
```

## Inputs

A map of **filename → source** (used by FFmpeg).

Supported sources:

* `s3`: `s3://bucket/key`
* `http`: public URL
* `base64`: Base64-encoded content

```json
{
  "input.mp4": {
    "http": "https://example.com/video.mp4"
  }
}
```

## FFmpeg Commands

An array of FFmpeg argument lists, executed sequentially.

```json
[
  ["-i", "input.mp4", "-vn", "output.mp3"]
]
```

Filenames reference input files or outputs from previous commands.

## Output Options

### Upload to S3

```json
{ "s3": "s3://my-bucket/outputs/" }
```

### Return Base64

```json
{ "base64": true }
```

### Stream via HTTP

```json
{ "inlineContentType": "audio/mpeg" }
```

> When streaming, the first output file is returned directly and no JSON is sent.

## JSON Response

```json
{
  "results": {
    "output.mp3": {
      "url": "...",
      "base64": "..."
    }
  }
}
```

## Example

```bash
curl -X POST http://localhost:8080/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "input.mp4": { "http": "https://example.com/video.mp4" }
    },
    "commands": [
      ["-i", "input.mp4", "-vn", "output.mp3"]
    ],
    "output": { "base64": true }
  }'
```

## Docker

### Image

```
ghcr.io/aureum-cloud/ffmpeg-api:latest
```

### Run

```bash
docker run -d -p 8080:8080 ghcr.io/aureum-cloud/ffmpeg-api:latest
```

API available at:

```
http://localhost:8080
```

## Notes

* FFmpeg is bundled
* No external dependencies
* Temporary files are cleaned automatically
* S3 credentials are supplied per request
* Go binary on scratch image
