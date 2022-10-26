package twt

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type DB interface {
	Get() (io.ReadCloser, error)
	PostStatus(io.Reader) error
	LogFollower(string) error
}

type FileDB struct {
	twtxtFilepath string
	twtxtMu       sync.RWMutex
	twtxtCache    []byte

	followers *log.Logger
}

func NewFileDB(basedir string) (*FileDB, error) {
	followers, err := createFollowersLog(basedir)
	if err != nil {
		return nil, err
	}
	return &FileDB{
		twtxtFilepath: filepath.Join(basedir, "twtxt.txt"),
		followers:     followers,
	}, nil
}

func createFollowersLog(basedir string) (*log.Logger, error) {
	followersFilepath := filepath.Join(basedir, "followers.log")
	fh, err := os.OpenFile(followersFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return log.New(fh, "", log.Ldate|log.Ltime), nil
}

func (f *FileDB) LogFollower(userAgent string) error {
	f.followers.Println(userAgent)
	return nil
}

func (f *FileDB) Get() (io.ReadCloser, error) {
	f.twtxtMu.RLock()
	if len(f.twtxtCache) == 0 {
		f.twtxtMu.RUnlock()
		f.twtxtMu.Lock()
		if len(f.twtxtCache) == 0 {
			fh, err := os.Open(f.twtxtFilepath)
			if err != nil {
				f.twtxtMu.Unlock()
				return nil, err
			}
			buf := bytes.NewBuffer(f.twtxtCache)
			_, err = io.Copy(buf, fh)
			_ = fh.Close()
			if err != nil {
				f.twtxtMu.Unlock()
				return nil, err
			}
			f.twtxtCache = buf.Bytes()
		}
		f.twtxtMu.Unlock()
		f.twtxtMu.RLock()
	}
	return &unlockableReader{bytes.NewBuffer(f.twtxtCache), &f.twtxtMu}, nil
}

func (f *FileDB) PostStatus(statusLine io.Reader) error {
	f.twtxtMu.Lock()
	defer f.twtxtMu.Unlock()
	f.twtxtCache = f.twtxtCache[:0]
	fh, err := os.OpenFile(f.twtxtFilepath, os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(fh, statusLine)
	_ = fh.Close()
	return err
}

type runlocker interface {
	RUnlock()
}

type unlockableReader struct {
	r  io.Reader
	ul runlocker
}

func (u *unlockableReader) Read(p []byte) (n int, err error) {
	return u.r.Read(p)
}

func (u *unlockableReader) Close() error {
	u.ul.RUnlock()
	return nil
}
