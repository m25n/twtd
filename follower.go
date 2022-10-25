package twt

import (
	"fmt"
	"net/url"
	"regexp"
)

type Follower interface {
	String() string
}

type SingleFollower struct {
	Nick string
	URL  *url.URL
}

func (f *SingleFollower) String() string {
	return fmt.Sprintf("follower\t%s\t%s", f.Nick, f.URL.String())
}

type MultiFollower struct {
	ListURL    *url.URL
	ContactURL *url.URL
}

func (f *MultiFollower) String() string {
	return fmt.Sprintf("list\t%s\t%s", f.ListURL.String(), f.ContactURL.String())
}

var singleFollowerRegex = regexp.MustCompile("^[^\\/]+\\/[^\\(]+ \\(+([^;]+); @([^\\(]+)\\)$")
var multiFollowerRegex = regexp.MustCompile("^[^\\/]+\\/[^\\(]+ \\(~([^;]+); contact=([^\\(]+)\\)$")

func FollowerUserAgent(userAgent string) bool {
	return singleFollowerRegex.MatchString(userAgent) || multiFollowerRegex.MatchString(userAgent)
}

func ParseFollower(userAgent string) Follower {
	matches := singleFollowerRegex.FindStringSubmatch(userAgent)
	if len(matches) == 3 {
		u, err := url.Parse(matches[1])
		if err != nil {
			return nil
		}
		return &SingleFollower{
			Nick: matches[2],
			URL:  u,
		}
	}

	matches = multiFollowerRegex.FindStringSubmatch(userAgent)
	if len(matches) == 3 {
		listURL, err := url.Parse(matches[1])
		if err != nil {
			return nil
		}
		contactURL, err := url.Parse(matches[2])
		if err != nil {
			return nil
		}
		return &MultiFollower{
			ListURL:    listURL,
			ContactURL: contactURL,
		}
	}

	return nil
}
