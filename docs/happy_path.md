## Happy path

This docu describes user interaction flow with vidi platform. To follow these steps you should have docker-compose project started with all containers running.

### User registration

Register user with following call:
```shell
curl -v -XPOST -H "Content-Type: application/json" \
    http://localhost:8080/api/users/register \
    -d '{"username":"johnS","password":"qwerty123"}'
```

Response should be `200 OK` with authentication cookie.

To double-check that user is created you can call login with same credentials:

```shell
curl -v -XPOST -H "Content-Type: application/json" \
    http://localhost:8081/api/users/login \
    -d '{"username":"johnS","password":"qwerty123"}'
```

Response should also be `200 OK` with authentication cookie. Both cookies will work until expiration time, which is 12 hours by default.

From this point we can make authenticated requests using received cookie. For convenience, we can make an alias:

```shell
alias curl_auth='curl --cookie "vidiSessID=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJRT0RBUW5xM1NkNjg2bmtsQ1RHZkhnIiwibmFtZSI6ImpvaG5TIiwiZXhwIjoxNzA1MTg2MjgyfQ.KnwB89hW9tQRvW18e_Xno45nvc_7aMcJimJ2ehVuMIM"'
```

### Video upload

#### Prepare file

Prepare mp4 video file. At this time, vidi supports only mp4 files with `AVC (h.264)` video codec and `mp4a` audio codec. 

To check if file is supported, you can use vidi cli:
```shell
go run cmd/vidi/vidi.go mp4 dump -f ./testfiles/test_seq_h264_high.mp4
```
At the end of output should be: "Codecs are supported!"

#### Upload video

First, create video object using video API call:
```shell
curl_auth -v -XPOST http://localhost:8080/api/video/user/
```
```shell
{"id":"PMO7_KJTRO2xntQyQdfRAw","status":"created","created_at":"2024-01-15 22:16:06.318666558 +0000 UTC m=+59.538474225","upload_url":"http://localhost:8080/upload/4F0lCtrmSl27B2bRFQNvDg"}
```
You should receive `201 Created` and video object parameters. Most important one is `upload_url`, which should be used to upload mp4 video file. This url is unique and valid for a short period of time (default upload session time is 5 min).

Now you can actually upload video file:

```shell
curl -v -XPOST http://localhost:8080/upload/4F0lCtrmSl27B2bRFQNvDg -H "Content-Type: video/mp4" --data-binary "@./testfiles/test_seq_h264_high.mp4"
```

After successful upload response should be `204 No Content`. You can check video status by using video ID from video creation response. Status could be `uploaded` or `ready` if video is very short and was already process.

```shell
curl_auth -v -XGET http://localhost:8080/api/video/user/PMO7_KJTRO2xntQyQdfRAw
```
```shell
{"id":"PMO7_KJTRO2xntQyQdfRAw","status":"ready","created_at":"0001-01-01 00:00:00 +0000 UTC"}
```

If video is ready, you can get watch URL by calling:
```shell
curl_auth -v -XPOST http://localhost:8080/api/video/user/PMO7_KJTRO2xntQyQdfRAw/watch
```
```shell
{"watch_url":"http://localhost:8080/watch/bsz-zpKZTFqOXgfLYSSgmg/manifest.mpd"}
```

Response will contain generated unique URL that can be fed to any player that understands MPEG-DASH.

If video is no longer needed you may delete it using:
```shell
curl_auth -v -XDELETE http://localhost:8080/api/video/user/PMO7_KJTRO2xntQyQdfRAw
```
