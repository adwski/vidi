## Vidi

WIP

This is a final project for Yandex Practicum's Go Course.

Vidi is a media server that supports:
 - simple user registration
 - video uploading by registered users
 - On-demand streaming of uploaded video using MPEG-DASH

Uploaded mp4 files are pre-processed, so they could be streamed to dash clients. Preprocessing includes:
 - Segmentation (using awesome [Eyevinn/mp4ff](https://github.com/Eyevinn/mp4ff) package)
 - MPD generation (with [Eyevinn/dash-mpd](https://github.com/Eyevinn/dash-mpd))

Also project uses:
- PostgreSQL for video object storage and user storage
- S3 compatible storage for video data (compose project uses minio)

## Development

### Docker

```bash
# start compose project
make docker-dev 

# stop compose project and cleanup volumes
make docker-dev-clean
```

### Tips

Check codec support in browser console
```javascript
MediaSource.isTypeSupported('video/mp4;codecs="avc1.64001e"')
```
