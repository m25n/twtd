package twt_test

import (
	"bytes"
	"errors"
	"github.com/m25n/twt"
	"github.com/m25n/twt/testhelper"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const status = "2022-01-01T00:00:00Z\tI have a thought\n"

func TestServer(t *testing.T) {
	t.Run("invalid methods respond with method not allowed", func(t *testing.T) {
		h := twt.Handler(testhelper.DummyLogger{}, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)

		req, _ := http.NewRequest("DELETE", "/twtxt.txt", nil)
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)

		require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	})

	t.Run("invalid paths respond with not found", func(t *testing.T) {
		h := twt.Handler(testhelper.DummyLogger{}, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)

		req, _ := http.NewRequest("GET", "/doesnotexist", nil)
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)

		require.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("posting status", func(t *testing.T) {
		t.Run("responds bad request with invalid media type ", func(t *testing.T) {
			h := twt.Handler(testhelper.DummyLogger{}, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/twtxt.txt", nil)
			req.Header.Set("Content-Type", "")

			h.ServeHTTP(res, req)

			require.Equal(t, http.StatusBadRequest, res.Code)
		})

		t.Run("responds unsupported media type", func(t *testing.T) {
			h := twt.Handler(testhelper.DummyLogger{}, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/twtxt.txt", nil)
			req.Header.Set("Content-Type", "text/plain")

			h.ServeHTTP(res, req)

			require.Equal(t, http.StatusUnsupportedMediaType, res.Code)
		})

		t.Run("responds internal server error when database fails", func(t *testing.T) {
			postErr := errors.New("post err")
			h := twt.Handler(testhelper.DummyLogger{}, &testhelper.StubDB{PostStatusErr: postErr}, twt.NoAuth(), testhelper.NoopEnqueueTask)

			res := postStatus(h, status)

			require.Equal(t, http.StatusInternalServerError, res.Code)
		})

		t.Run("logs error when database fails", func(t *testing.T) {
			postErr := errors.New("post err")
			logger := testhelper.NewMockLogger()
			h := twt.Handler(logger, &testhelper.StubDB{PostStatusErr: postErr}, twt.NoAuth(), testhelper.NoopEnqueueTask)

			_ = postStatus(h, status)

			require.Contains(t, logger.PostingStatusErrs, postErr)
		})
	})

	t.Run("twtxt.txt", func(t *testing.T) {
		t.Run("responds with correct content type", func(t *testing.T) {
			h := twt.Handler(testhelper.DummyLogger{}, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			h.ServeHTTP(res, req)

			require.Equal(t, "text/vnd.twtxt+plain", res.Header().Get("Content-Type"))
		})

		t.Run("responds with status OK", func(t *testing.T) {
			h := twt.Handler(testhelper.DummyLogger{}, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			h.ServeHTTP(res, req)

			require.Equal(t, http.StatusOK, res.Code)
		})

		t.Run("responds with internal server error when the database has an error", func(t *testing.T) {
			h := twt.Handler(testhelper.DummyLogger{}, &testhelper.StubDB{GetErr: errors.New("db error")}, twt.NoAuth(), testhelper.NoopEnqueueTask)

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			h.ServeHTTP(res, req)

			require.Equal(t, http.StatusInternalServerError, res.Code)
		})

		t.Run("logs error when the database has an error", func(t *testing.T) {
			logger := testhelper.NewMockLogger()
			dbErr := errors.New("db error")
			h := twt.Handler(logger, &testhelper.StubDB{GetErr: dbErr}, twt.NoAuth(), testhelper.NoopEnqueueTask)

			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			h.ServeHTTP(httptest.NewRecorder(), req)

			require.Contains(t, logger.GettingTwtxtErrs, dbErr)
		})

		t.Run("logs error when there is an error writing the response", func(t *testing.T) {
			logger := testhelper.NewMockLogger()
			readErr := errors.New("read error")
			h := twt.Handler(logger, &testhelper.StubDB{GetReadCloser: io.NopCloser(&testhelper.StubReader{ReadErr: readErr})}, twt.NoAuth(), testhelper.NoopEnqueueTask)

			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			h.ServeHTTP(httptest.NewRecorder(), req)

			require.Contains(t, logger.WritingBodyErrs, readErr)
		})

		t.Run("logs error there is an error enqueuing the task to log a follower", func(t *testing.T) {
			logger := testhelper.NewMockLogger()
			loggingFollowerErr := errors.New("enqueue error")
			h := twt.Handler(logger, testhelper.EmptyStubDB(), twt.NoAuth(), testhelper.StubEnqueueTask(loggingFollowerErr))

			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			h.ServeHTTP(httptest.NewRecorder(), req)

			require.Contains(t, logger.FollowerLoggingErrs, loggingFollowerErr)
		})

		t.Run("logs followers", func(t *testing.T) {
			db := testhelper.NewMockDB()
			h := twt.Handler(testhelper.DummyLogger{}, db, twt.NoAuth(), testhelper.SyncEnqueueTask)

			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			req.Header.Set("User-Agent", "twtxt/1.2.3 (+https://example.com/twtxt.txt; @somebody)")
			h.ServeHTTP(httptest.NewRecorder(), req)

			require.Contains(t, db.Followers, "twtxt/1.2.3 (+https://example.com/twtxt.txt; @somebody)")
		})

		t.Run("logs error there is an error enqueuing the task to log a follower", func(t *testing.T) {
			logger := testhelper.NewMockLogger()
			followerErr := errors.New("error logging follower")
			db := &testhelper.StubDB{GetReadCloser: testhelper.EmptyReadCloser, LogFollowerErr: followerErr}
			h := twt.Handler(logger, db, twt.NoAuth(), testhelper.SyncEnqueueTask)

			req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
			req.Header.Set("User-Agent", "twtxt/1.2.3 (+https://example.com/twtxt.txt; @somebody)")
			h.ServeHTTP(httptest.NewRecorder(), req)

			require.Contains(t, logger.FollowerLoggingErrs, followerErr)
		})
	})

	t.Run("posted statuses can be read back", func(t *testing.T) {
		h := twt.Handler(testhelper.DummyLogger{}, testhelper.NewFakeDB(), twt.NoAuth(), testhelper.NoopEnqueueTask)

		_ = postStatus(h, status)
		twtxt := getTwtxt(h)

		require.Equal(t, status, twtxt)
	})
}

func getTwtxt(h http.Handler) string {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/twtxt.txt", nil)
	h.ServeHTTP(res, req)
	buf := bytes.NewBuffer(nil)
	_, _ = io.Copy(buf, res.Body)
	return buf.String()
}

func postStatus(h http.Handler, status string) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	buf := bytes.NewBufferString(status)
	req, _ := http.NewRequest("PATCH", "/twtxt.txt", buf)
	req.Header.Set("Content-Type", "text/vnd.twtxt+plain")
	h.ServeHTTP(res, req)
	return res
}
