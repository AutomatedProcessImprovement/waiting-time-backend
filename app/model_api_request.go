package app

import (
	"encoding/json"
	"net/url"
)

type ApiRequest struct {
	EventLog         *URL `json:"event_log,omitempty"`
	CallbackEndpoint *URL `json:"callback_endpoint,omitempty"`
}

func (r *ApiRequest) UnmarshalJSON(data []byte) error {
	var jsonData = map[string]string{}

	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return err
	}

	eventLog := jsonData["event_log"]
	u, err := url.Parse(eventLog)
	if err != nil {
		return err
	}
	r.EventLog = &URL{url: u}

	callbackEndpoint := jsonData["callback_endpoint"]
	u, err = url.Parse(callbackEndpoint)
	if err != nil {
		return err
	}
	r.CallbackEndpoint = &URL{url: u}

	return nil
}
