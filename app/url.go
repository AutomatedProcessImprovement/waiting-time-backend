package app

import (
	"encoding/json"
	"net/url"
)

type URL struct {
	url *url.URL
}

func (u *URL) String() string {
	if u == nil {
		return ""
	}

	return u.url.String()
}

func (u *URL) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	u.url, err = url.Parse(s)
	if err != nil {
		return err
	}

	return nil
}

func (u *URL) MarshalJSON() ([]byte, error) {
	if u == nil {
		return []byte("null"), nil
	}

	return json.Marshal(u.url.String())
}
