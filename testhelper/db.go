package testhelper

import (
	"bytes"
	"io"
)

type FakeDB struct {
	statusLines []string
	followers   []string
}

func NewFakeDB() *FakeDB {
	return &FakeDB{}
}

func (db *FakeDB) Get() (io.ReadCloser, error) {
	buf := bytes.NewBuffer(nil)
	for _, statusLine := range db.statusLines {
		buf.WriteString(statusLine)
	}
	return io.NopCloser(buf), nil
}

func (db *FakeDB) PostStatus(statusLine io.Reader) error {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, statusLine)
	if err != nil {
		return err
	}
	db.statusLines = append(db.statusLines, buf.String())
	return nil
}

func (db *FakeDB) LogFollower(follower string) error {
	db.followers = append(db.followers, follower)
	return nil
}

type StubDB struct {
	GetReadCloser io.ReadCloser
	GetErr        error

	PostStatusErr error

	LogFollowerErr error
}

func EmptyStubDB() *StubDB {
	return &StubDB{
		GetReadCloser: EmptyReadCloser,
	}
}

func (db *StubDB) Get() (io.ReadCloser, error) {
	return db.GetReadCloser, db.GetErr
}

func (db *StubDB) PostStatus(_ io.Reader) error {
	return db.PostStatusErr
}

func (db *StubDB) LogFollower(_ string) error {
	return db.LogFollowerErr
}

type MockDB struct {
	StatusLines []string
	Followers   []string
}

func NewMockDB() *MockDB {
	return &MockDB{}
}

func (db *MockDB) Get() (io.ReadCloser, error) {
	return EmptyReadCloser, nil
}

func (db *MockDB) PostStatus(reader io.Reader) error {
	statusLine := bytes.NewBuffer(nil)
	_, _ = io.Copy(statusLine, reader)
	db.StatusLines = append(db.StatusLines, statusLine.String())
	return nil
}

func (db *MockDB) LogFollower(follower string) error {
	db.Followers = append(db.Followers, follower)
	return nil
}
