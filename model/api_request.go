package model

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ApiRequest is a request's body for POST /jobs.
//
// swagger:model
type ApiRequest struct {
	EventLogURL          string            `json:"event_log,omitempty"`
	EventLogURL_         *URL              `json:"-"`
	CallbackEndpointURL  string            `json:"callback_endpoint,omitempty"`
	CallbackEndpointURL_ *URL              `json:"-"`
	ColumnMapping        map[string]string `json:"column_mapping,omitempty"`
}

func (r *ApiRequest) UnmarshalJSON(data []byte) error {
	var jsonData = map[string]interface{}{}

	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return err
	}

	eventLog, ok := jsonData["event_log"].(string)
	if !ok {
		return fmt.Errorf("event_log is not a string")
	}
	r.EventLogURL = eventLog
	u, err := url.Parse(eventLog)
	if err != nil {
		return err
	}
	r.EventLogURL_ = &URL{URL: u}

	// callback_endpoint is optional
	callbackEndpoint, ok := jsonData["callback_endpoint"]
	if ok {
		callbackEndpointStr, ok := callbackEndpoint.(string)
		if !ok {
			return fmt.Errorf("callback_endpoint is not a string")
		}
		r.CallbackEndpointURL = callbackEndpointStr
		u, err = url.Parse(callbackEndpointStr)
		if err != nil {
			return err
		}
		r.CallbackEndpointURL_ = &URL{URL: u}
	}

	// column_mapping is optional
	mapping, ok := jsonData["column_mapping"]
	if ok {
		mappingMap, ok := mapping.(map[string]interface{})
		if !ok {
			return fmt.Errorf("column_mapping is not a valid dictionary")
		}
		mappingMapStr := map[string]string{}
		for k, v := range mappingMap {
			vStr, ok := v.(string)
			if !ok {
				return fmt.Errorf("column_mapping value is not a string: %v", v)
			}
			mappingMapStr[k] = vStr
		}
		r.ColumnMapping = mappingMapStr
	}

	return nil
}

func (r *ApiRequest) MarshalJSON() ([]byte, error) {
	jsonData := map[string]string{}

	jsonData["event_log"] = r.EventLogURL_.String()
	jsonData["callback_endpoint"] = r.CallbackEndpointURL_.String()

	return json.Marshal(jsonData)
}
