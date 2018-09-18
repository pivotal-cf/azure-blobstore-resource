package api

import (
	"net/url"
	"time"
)

func URLAppendTimeStamp(baseURL string, snapshot time.Time) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("snapshot", snapshot.UTC().Format("2006-01-02T15:04:05.0000000Z"))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
