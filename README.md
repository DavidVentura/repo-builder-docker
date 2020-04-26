
This tool listens for HTTP Post requests coming from [Gogs](gogs.io) and on `tag creation` + `push to master`
[events](https://gogs.io/docs/features/webhook.html), it will execute Docker, targeting `build.json` file 
at the root of the target repository, passing the REF build argument, populated with the received `ref`.


# Features

* Notifications of build start, success and failure are pushed to telegram.
  * The notifications include a link to see the build logs.
* Build artifacts are pushed to S3.

This project currently only builds nodejs projects by running `npm install; npm run build` 

![](screenshots/telegram-link.png)

# Requirements for building repositories

1. Populate a `build.json` file at the root of the repository

```
{
    "subprojects": [
    {
      "name": "subproject-1",
      "dir": "recipes",
      "artifacts": ["build/"]
    },
    {
      "name": "subproject-2",
      "dir": "ktchn",
      "artifacts": ["dist/bundle.js"]
    }
    ]
}
```

2. Create a webhook in gogs, you can do this in the UI or by posting to the API (see [here](https://github.com/gogs/docs-api/blob/master/Repositories/Webhooks.md))

```
ACCESS_TOKEN=blah; curl -X POST http://gogs.labs/api/v1/repos/:username/:reponame/hooks\
	-H "Authorization: token ${ACCESS_TOKEN}"\
	-H "Content-Type: application/json"\
	--data '{"type": "gogs",
	"config": {"url": "http://ci.labs:8080/hook", "content_type": "json"},
	"events": ["create", "push"]}'
```

(To create an access token go to your personal account -> settings -> applications -> generate new token)

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
    "BuildDockerfilePath": "/home/ci/Dockerfile",
    "Repos": [
        {
            "Name":                      "Recipes",
            "GitUrl":                    "ssh://git@gogs:2222/tati/kitchn.git",
            "Bucket":                    "recipes",
            "TelegramChatId":            -311945893
        },
        {
            "Name":                      "Test Repo",
            "GitUrl":                    "ssh://git@gogs:2222/david/test.git",
            "Bucket":                    "testrepo",
            "TelegramChatId":            1719831
        }
    ]
}

```

