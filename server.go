package twt

import (
	"context"
	"github.com/m25n/twt/task"
	"io"
	"mime"
	"net/http"
	"time"
)

type Logger interface {
	WritingBodyErr(err error)
	FollowerLoggingErr(err error)
	PostingStatusErr(err error)
	GettingTwtxtErr(err error)
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Handler(logger Logger, db DB, auth Middleware, enqueueTask task.EnqueueFunc) http.Handler {
	get := getHandler(logger, db, enqueueTask)
	patch := auth(patchHandler(logger, db))
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch {
		case req.URL.Path == "/twtxt.txt":
			switch req.Method {
			case http.MethodGet:
				get(res, req)
			case http.MethodPatch:
				patch(res, req)
			default:
				http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.NotFound(res, req)
		}
	})
}

func getHandler(logger Logger, db DB, enqueueTask task.EnqueueFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/vnd.twtxt+plain")
		file, err := db.Get()
		if err != nil {
			logger.GettingTwtxtErr(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(res, file)
		_ = file.Close()
		if err != nil {
			logger.WritingBodyErr(err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = enqueueTask(ctx, func(ctx context.Context) {
			userAgent := req.Header.Get("User-Agent")
			if FollowerUserAgent(userAgent) {
				if err := db.LogFollower(userAgent); err != nil {
					logger.FollowerLoggingErr(err)
				}
			}
		})
		if err != nil {
			logger.FollowerLoggingErr(err)
		}
	}
}

func patchHandler(logger Logger, db DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		if mediaType != "text/vnd.twtxt+plain" && mediaType != "text/vnd.twtxt" {
			res.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		err = db.PostStatus(req.Body)
		if err != nil {
			logger.PostingStatusErr(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
