package logger

import "log"

type Logger log.Logger

func New(logger *log.Logger) *Logger {
	return (*Logger)(logger)
}

func (l *Logger) logger() *log.Logger {
	return (*log.Logger)(l)
}

func (l *Logger) WritingBodyErr(err error) {
	l.logger().Println("error writing body:", err.Error())
}

func (l *Logger) GettingTwtxtErr(err error) {
	l.logger().Println("error getting twtxt.txt", err)
}

func (l *Logger) FollowerLoggingErr(err error) {
	l.logger().Println("error logging follower:", err.Error())
}

func (l *Logger) PostingStatusErr(err error) {
	l.logger().Println("error posting status:", err.Error())
}
