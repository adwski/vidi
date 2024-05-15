## Vidi

![Coverage](https://img.shields.io/badge/Coverage-66.1%25-yellow)

WIP

This is a final project for Yandex Practicum's Go Course.

Vidi is a media server that supports:
 - simple user registration
 - video uploading by registered users
 - On-demand streaming of uploaded videos using MPEG-DASH

Uploaded mp4 files are pre-processed, so they could be streamed to dash clients. Preprocessing includes:
 - Segmentation (using awesome [Eyevinn/mp4ff](https://github.com/Eyevinn/mp4ff) package)
 - MPD generation (with [Eyevinn/dash-mpd](https://github.com/Eyevinn/dash-mpd))

Also project uses:
- PostgreSQL for video object storage and user storage
- S3 compatible storage for video data (compose project uses minio)
- Redis for media sessions storage

## Development

### Docker

```bash
# start compose project
make docker-dev 

# stop compose project and cleanup volumes
make docker-dev-clean
```

### Tests

```bash
# run unit tests
make unittests

# run unit and end2end tests
# docker compose project should be stopped (it will be ran by the command)
make test-all
```

### Happy Path

User interaction flow is described [here](./docs/happy_path.md).
