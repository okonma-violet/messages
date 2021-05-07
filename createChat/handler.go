package main

import (
	"context"
	"errors"
	"flag"
	"lib"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/mgo/bson"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

type CreateChat struct {
	mgoSession *mgo.Session
	mgoColl    *mgo.Collection
}
type chatInfo struct {
	Id    bson.ObjectId `bson:"_id"`
	Users []string      `bson:"users"`
	Name  string        `bson:"name"`
	Type  int           `bson:"type"`
}

func NewSendMessage(mgoAddr string, mgoColl string) (*CreateChat, error) {

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		logger.Error("Mongo conn", err)
		return nil, err
	}

	mgoCollection := mgoSession.DB("main").C(mgoColl)

	return &CreateChat{mgoSession: mgoSession, mgoColl: mgoCollection}, nil

}

func (conf *CreateChat) Close() error {
	conf.mgoSession.Close()
	return nil
}

func (conf *CreateChat) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	w = &suckhttp.Response{}

	cookie := lib.GetCookie(r.GetHeader(suckhttp.Cookie), "koki")
	if cookie == nil {
		response := suckhttp.NewResponse(400, "Bad request")
		return response, errors.New("Not set cookie token")
	}

	queryValues, err := url.ParseQuery(r.Uri.RawQuery)
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return
	}
	chatType := queryValues.Get("type")

	switch chatType {
	case "0":

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		ownerHash := queryValues.Get("u1") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯

		uid := suckutils.GetRandUID(314)
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": uid, "users": []string{ownerHash}, "type": 0}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 0, "users": ownerHash}

		foundedChat := &chatInfo{}
		insertedChat, err := conf.mgoColl.Find(selector).Apply(change, foundedChat)
		if err != nil { // TODO: ПРОВЕРИТЬ ОШИБКУ mgo.ErrNotFound
			logger.Error("Mongo insert", err)
			return nil, err
		}

		if foundedChat.Id == "" { // TODO: И ЕСЛИ ВЫШИБАЕТ ОШИБКУ ПИХАЕМ ЕЕ СЮДА
			insertedChatId, ok := insertedChat.UpsertedId.([]byte)
			if ok && insertedChatId != nil {
				w.SetStatusCode(200, "OK")
				w.SetBody(insertedChatId) // TODO: КУДА WRITE? Редирект?
				return w, nil
			} else {
				logger.Err
			}
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody(foundChat.Id) // TODO: КУДА WRITE? Редирект?
		return responce, nil

	case "group":

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		ownerHash := queryValues.Get("u1") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯

		chatName := queryValues.Get("name")

		if chatName == "" {
			chatName = "Group chat"
		}

		uid := suckutils.GetRandUID(314)
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": uid, "users": ownerHash, "type": 2}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 2, "users.0": ownerHash, "name": chatName}
		foundChat := &ChatInfo{}

		insertedChat, err := handler.mongoColl.Find(selector).Apply(change, foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == nil {
			insertedChatId, ok := insertedChat.UpsertedId.([]byte)
			if ok && insertedChatId != nil {
				responce := suckhttp.NewResponse(200, "OK")
				responce.SetBody(insertedChatId) // TODO: КУДА WRITE?
				return responce, nil
			} else {
				return nil, nil // ??
			}
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody(foundChat.Id) // TODO: КУДА WRITE?
		return responce, nil

	case "ls":

		secondUserHash := queryValues.Get("u2")
		if secondUserHash == "" {
			return suckhttp.NewResponse(400, "Miss second user"), nil
		}

		// КУДА ТО ТАМ АВТОРИЗУЕМСЯ ¯\_(ツ)_/¯

		ownerHash := queryValues.Get("u1") // TODO: ОТКУДА ТО ПОЛУЧАЕМ ХЭШ ¯\_(ツ)_/¯

		uid := suckutils.GetRandUID(314)
		change := mgo.Change{
			Update:    bson.M{"$setOnInsert": bson.M{"_id": uid, "users": []string{ownerHash, secondUserHash}, "type": 1}},
			Upsert:    true,
			ReturnNew: false,
			Remove:    false,
		}
		selector := &bson.M{"type": 1, "$or": []bson.M{{"users": []string{ownerHash, secondUserHash}}, {"users": []string{secondUserHash, ownerHash}}}}
		foundChat := &ChatInfo{}

		insertedChat, err := handler.mongoColl.Find(selector).Apply(change, foundChat)
		if err != nil {
			return nil, err
		}

		if foundChat.Id == nil {
			insertedChatId, ok := insertedChat.UpsertedId.([]byte)
			if ok && insertedChatId != nil {
				responce := suckhttp.NewResponse(200, "OK")
				responce.SetBody(insertedChatId) // TODO: КУДА WRITE?
				return responce, nil
			} else {
				return nil, nil // ??
			}
		}

		responce := suckhttp.NewResponse(200, "OK")
		responce.SetBody(foundChat.Id) // TODO: КУДА WRITE?
		return responce, nil

	default:
		return suckhttp.NewResponse(400, "Bad request"), errors.New("Type error")
	}
}

func main() {
	port := flag.Int("port", 0, "WebService port")
	flag.Parse()

	if *port <= 0 {
		println("Port not set")
		os.Exit(1)
	}

	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
	loggerctx, loggercancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		loggercancel()
		<-logger.AllLogsFlushed
	}()

	logger.SetupLogger(loggerctx, time.Second*10, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

	handler, err := NewCreateChatHandler("127.0.0.1:27017")
	if err != nil {
		logger.Error("Mongo connection", err)
		return
	}

	logger.Error("HTTP service", httpservice.ServeHTTPService(ctx, suckutils.ConcatTwo(":", strconv.Itoa(*port)), handler)) // TODO: отхардкодить порт?
}

type CreateChatHandler struct {
	mongoColl *mgo.Collection
}

func (handler *CreateChatHandler) Close() {
	handler.mongoColl.Database.Session.Close()
}

func NewCreateChatHandler(connectionString string) (*CreateChatHandler, error) { // порнография?
	mongoSession, err := mgo.Dial(connectionString)

	if err != nil {
		return nil, err
	}

	return &CreateChatHandler{mongoColl: mongoSession.DB("main").C("chats")}, nil
}
