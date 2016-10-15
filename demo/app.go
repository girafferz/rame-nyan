package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/datastore"
	"fmt"
)

var botHandler *httphandler.WebhookHandler

func init() {
	err := godotenv.Load("line.env")
	if err != nil {
		panic(err)
	}

	botHandler, err = httphandler.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	botHandler.HandleEvents(handleCallback)

	http.Handle("/callback", botHandler)
	http.HandleFunc("/task", handleTask)
}

// Webhook を受け取って TaskQueueに詰める関数
func handleCallback(evs []*linebot.Event, r *http.Request) {
	c := newContext(r)
	ts := make([]*taskqueue.Task, len(evs))
	for i, e := range evs {
		j, err := json.Marshal(e)
		if err != nil {
			errorf(c, "json.Marshal: %v", err)
			return
		}
		data := base64.StdEncoding.EncodeToString(j)
		t := taskqueue.NewPOSTTask("/task", url.Values{"data": {data}})
		ts[i] = t
	}
	taskqueue.AddMulti(c, ts, "")
}

type Entity struct {
	Value string
}

// 受け取ったメッセージを処理する関数
func handleTask(w http.ResponseWriter, r *http.Request) {
	c := newContext(r)
	data := r.FormValue("data")
	if data == "" {
		errorf(c, "No data")
		return
	}

	j, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		errorf(c, "base64 DecodeString: %v", err)
		return
	}

	e := new(linebot.Event)
	err = json.Unmarshal(j, e)
	if err != nil {
		errorf(c, "json.Unmarshal: %v", err)
		return
	}

	bot, err := newLINEBot(c)
	if err != nil {
		errorf(c, "newLINEBot: %v", err)
		return
	}

	logf(c, "==source==")
	logf(c, "EventType: %s\nMessage: %#v", e.Source)
	logf(c, "==source==")

	logf(c, "EventType: %s\nMessage: %#v", e.Type, e.Message)

	m := linebot.NewTextMessage("ok")
	if _, err = bot.ReplyMessage(e.ReplyToken, m).WithContext(c).Do(); err != nil {
		errorf(c, "ReplayMessage: %v", err)
		return
	}

	c, r, w = writeDb(c, r, w)

	w.WriteHeader(200)
}

func writeDb (c context.Context, r *http.Request, w http.ResponseWriter) {
	logf(c, "==writedb==")
	k := datastore.NewKey(c, "Entity", "stringID", 0, nil)
	e := new(Entity)
	if err := datastore.Get(c, k, e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	old := e.Value
	e.Value = r.URL.Path
	if _, err := datastore.Put(c, k, e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "old=%q\nnew=%q\n", old, e.Value)

	logf(c, "==writedb return==")
	return c, r, w
}

func newLINEBot(c context.Context) (*linebot.Client, error) {
	return botHandler.NewClient(
		linebot.WithHTTPClient(urlfetch.Client(c)),
	)
}

// newContext は appengine.NewContext を短く書くための関数
func newContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

// logf は log.Infof を短く書くための関数
func logf(c context.Context, format string, args ...interface{}) {
	log.Infof(c, format, args...)
}

// errorf は log.Errorf を短く書くための関数
func errorf(c context.Context, format string, args ...interface{}) {
	log.Errorf(c, format, args...)
}
