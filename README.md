
This tool listens for HTTP Post requests coming from [Gogs](gogs.io) and on `tag creation`
[event](https://gogs.io/docs/features/webhook.html), it will execute Docker, targeting the
Dockerfile specified in the config, passing the TAG build argument with the received `ref`.


# Features

* Notifications of build start, success and failure are pushed to telegram.
  * The notifications include a link to see the build logs.
* Build artifacts are pushed to S3.

![](screenshots/telegram-link.png)

# Requirements for the Dockerfiles

1. Image must be called `builder` (`FROM blah:latest as builder`)
2. It must list its artifacts, one artifact per line, in `/artifacts`

# TODO
* Per-Build lock to avoid concurrent builds (or clone once per tag?)

# Configuration

## Environment variables

The following environment variables are used

```
S3_ACCESS_KEY=access_key
S3_SECRET_KEY=secret_key
TELEGRAM_BOT_KEY=XXXXXXXXX:YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
```

## Config file

```json


{
    "LogPath": "/tmp/",
    "repoCloneBase": "/tmp/",
    "Repos": [
        {
            "Name":                      "Recipes",
            "GitUrl":                    "ssh://git@gogs:2222/tati/kitchn.git",
            "RelativePathForDockerfile": "build/Dockerfile",
            "Bucket":                    "recipes",
            "TelegramChatId":            -311945893
        },
        {
            "Name":                      "Test Repo",
            "GitUrl":                    "ssh://git@gogs:2222/david/test.git",
            "RelativePathForDockerfile": "build/Dockerfile",
            "Bucket":                    "testrepo",
            "TelegramChatId":            1719831
        }
    ]
}

```

