# rame-nyan

line bot for ra-men recomendation

# files
```
$ tree demo
demo
├── app.go
├── app.yaml
└── line.env

$ cat demo/app.yaml
application: (google cloud platform application id <projectID>)
version: 1
runtime: go
api_version: go1

handlers:
- url: /task.*
  script: _go_app
  login: admin
  secure: always
- url: /.*
  script: _go_app
  secure: always

$ cat demo/line.env
LINE_BOT_CHANNEL_SECRET=(line bot channnel secret)
LINE_BOT_CHANNEL_TOKEN=(line bot channel token)
```

# init

```
brew install go-app-engine-64
export GOPATH=`pwd`
go get -u github.com/line/line-bot-sdk-go/linebot
go get -u github.com/joho/godotenv
go get -u google.golang.org/appengine
```

# deoloy
modify code
```
cd demo
goapp deploy
```


# line developer setup

set variable Webhook as "https://<projectID>.appspot.com/callback"

