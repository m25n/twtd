package twt

import (
	"context"
	"github.com/m25n/twt/task"
	"io"
	"log"
	"mime"
	"net/http"
	"time"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Handler(db DB, auth Middleware, enqueueTask task.EnqueueFunc) http.Handler {
	getTwtxt := handleGetTwtxt(db, enqueueTask)
	patchTwtxt := auth(handlePatchTwtxt(db))
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch {
		case req.URL.Path == "/twtxt.txt":
			switch req.Method {
			case http.MethodGet:
				getTwtxt(res, req)
			case http.MethodPatch:
				patchTwtxt(res, req)
			default:
				http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.NotFound(res, req)
		}
	})
}

func handleGetTwtxt(db DB, enqueueTask task.EnqueueFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		file, err := db.Get()
		if err != nil {
			log.Print(err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(res, file)
		_ = file.Close()
		if err != nil {
			log.Print(err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = enqueueTask(ctx, func(ctx context.Context) {
			userAgent := req.Header.Get("User-Agent")
			if FollowerUserAgent(userAgent) {
				if err := db.LogFollower(userAgent); err != nil {
					log.Println("error logging follower:", err.Error())
				}
			}
		})
		if err != nil {
			log.Printf("error logging follower: %s", err.Error())
		}
	}
}

func handlePatchTwtxt(db DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			log.Println(err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		if mediaType != "text/x-diff" {
			res.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		err = db.Patch(req.Body)
		if err != nil {
			log.Println(err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}