package model

import (
	"encoding/json"
	"net/url"
)

// ApiRequest is a request's body for POST /jobs.
//
// swagger:model
type ApiRequest struct {
	EventLogURL          string `json:"event_log,omitempty"`
	EventLogURL_         *URL   `json:"-"`
	CallbackEndpointURL  string `json:"callback_endpoint,omitempty"`
	CallbackEndpointURL_ *URL   `json:"-"`
}

func (r *ApiRequest) UnmarshalJSON(data []byte) error {
	var jsonData = map[string]string{}

	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return err
	}

	eventLog := jsonData["event_log"]
	r.EventLogURL = eventLog
	u, err := url.Parse(eventLog)
	if err != nil {
		return err
	}
	r.EventLogURL_ = &URL{URL: u}

	callbackEndpoint := jsonData["callback_endpoint"]
	r.CallbackEndpointURL = callbackEndpoint
	u, err = url.Parse(callbackEndpoint)
	if err != nil {
		return err
	}
	r.CallbackEndpointURL_ = &URL{URL: u}

	return nil
}

func (r *ApiRequest) MarshalJSON() ([]byte, error) {
	jsonData := map[string]string{}

	jsonData["event_log"] = r.EventLogURL_.String()
	jsonData["callback_endpoint"] = r.CallbackEndpointURL_.String()

	return json.Marshal(jsonData)
}
