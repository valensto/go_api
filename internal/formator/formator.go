package formator

import "net/http"

func NewJSONData(dt, id string, attr interface{}) JsonData {
	return JsonData{
		Type:       dt,
		ID:         id,
		Attributes: attr,
	}
}

type JsonData struct {
	Type       string      `json:"type"`
	ID         string      `json:"id"`
	Attributes interface{} `json:"attributes"`
}

type RespErr struct {
	Status int    `json:"status"`
	Source source `json:"source"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type source struct {
	URI    string `json:"uri"`
	Method string `json:"method"`
}

func NewRespErr(r *http.Request, status int, title, detail string) RespErr {
	jsonErr := RespErr{
		Status: status,
		Source: source{
			URI:    r.RequestURI,
			Method: r.Method,
		},
		Title:  title,
		Detail: detail,
	}
	return jsonErr
}
