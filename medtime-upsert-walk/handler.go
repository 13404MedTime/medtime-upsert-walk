package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	botToken        = "8032551691:AAGOvf0vFQANorJyMcW8ZVa72PDi5hgfJoE"
	chatID          = "6805374430"
	baseUrl         = "https://api.admin.u-code.io"
	logFunctionName = "ucode-template"
	IsHTTP          = true // if this is true benchmark test works.
)

// Request structures
type (
	// Handle request body
	NewRequestBody struct {
		RequestData HttpRequest `json:"request_data"`
		Auth        AuthData    `json:"auth"`
		Data        Data        `json:"data"`
	}

	HttpRequest struct {
		Method  string      `json:"method"`
		Path    string      `json:"path"`
		Headers http.Header `json:"headers"`
		Params  url.Values  `json:"params"`
		Body    []byte      `json:"body"`
	}

	AuthData struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	// Function request body >>>>> GET_LIST, GET_LIST_SLIM, CREATE, UPDATE
	Request struct {
		Data map[string]interface{} `json:"data"`
	}

	// most common request structure -> UPDATE, MULTIPLE_UPDATE, CREATE, DELETE
	Data struct {
		AppId      string                 `json:"app_id"`
		Method     string                 `json:"method"`
		ObjectData map[string]interface{} `json:"object_data"`
		ObjectIds  []string               `json:"object_ids"`
		TableSlug  string                 `json:"table_slug"`
		UserId     string                 `json:"user_id"`
	}

	FunctionRequest struct {
		BaseUrl     string  `json:"base_url"`
		TableSlug   string  `json:"table_slug"`
		AppId       string  `json:"app_id"`
		Request     Request `json:"request"`
		DisableFaas bool    `json:"disable_faas"`
	}

	SlimFunctionRequest struct {
		BaseUrl     string                 `json:"base_url"`
		TableSlug   string                 `json:"table_slug"`
		AppId       string                 `json:"app_id"`
		Request     map[string]interface{} `json:"request"`
		DisableFaas bool                   `json:"disable_faas"`
		DateFilter  DateFilter             `json:"date_filter"`
	}
	DateFilter struct {
		Gte      string `json:"$gte"`
		Lt       string `json:"$lt"`
		ClientId string `json:"client_id"`
	}
)

// Response structures
type (
	// Create function response body >>>>> CREATE
	Datas struct {
		Data struct {
			Data struct {
				Data map[string]interface{} `json:"data"`
			} `json:"data"`
		} `json:"data"`
	}

	// ClientApiResponse This is get single api response >>>>> GET_SINGLE_BY_ID, GET_SLIM_BY_ID
	ClientApiResponse struct {
		Data ClientApiData `json:"data"`
	}

	ClientApiData struct {
		Data ClientApiResp `json:"data"`
	}

	ClientApiResp struct {
		Response map[string]interface{} `json:"response"`
	}

	Response struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}

	// GetListClientApiResponse This is get list api response >>>>> GET_LIST, GET_LIST_SLIM
	GetListClientApiResponse struct {
		Data GetListClientApiData `json:"data"`
	}

	GetListClientApiData struct {
		Data GetListClientApiResp `json:"data"`
	}

	GetListClientApiResp struct {
		Response []map[string]interface{} `json:"response"`
	}

	// ClientApiUpdateResponse This is single update api response >>>>> UPDATE
	ClientApiUpdateResponse struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		Data        struct {
			TableSlug string                 `json:"table_slug"`
			Data      map[string]interface{} `json:"data"`
		} `json:"data"`
	}

	// ClientApiMultipleUpdateResponse This is multiple update api response >>>>> MULTIPLE_UPDATE
	ClientApiMultipleUpdateResponse struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		Data        struct {
			Data struct {
				Objects []map[string]interface{} `json:"objects"`
			} `json:"data"`
		} `json:"data"`
	}

	ResponseStatus struct {
		Status string `json:"status"`
	}
)

// Testing types
type (
	Asserts struct {
		Request  NewRequestBody
		Response Response
	}

	FunctionAssert struct{}
)

type GetSlimListRequest struct {
	Date      Date   `json:"date"`
	CleintsID string `json:"cleints_id"`
	FromOfs   bool   `json:"from-ofs"`
}
type Date struct {
	Gte string `json:"$gte"`
	Lt  string `json:"$lt"`
}

func (f FunctionAssert) GetAsserts() []Asserts {
	var appId = os.Getenv("APP_ID")

	return []Asserts{
		{
			Request: NewRequestBody{
				Data: Data{
					AppId:     appId,
					ObjectIds: []string{"96b6c9e0-ec0c-4297-8098-fa9341c40820"},
				},
			},
			Response: Response{
				Status: "done",
			},
		},
		{
			Request: NewRequestBody{
				Data: Data{
					AppId:     appId,
					ObjectIds: []string{"96b6c9e0-ec0c-4297-8098"},
				},
			},
			Response: Response{Status: "error"},
		},
	}
}

func (f FunctionAssert) GetBenchmarkRequest() Asserts {
	var appId = os.Getenv("APP_ID")
	return Asserts{
		Request: NewRequestBody{
			Data: Data{
				AppId:     appId,
				ObjectIds: []string{"96b6c9e0-ec0c-4297-8098-fa9341c40820"},
			},
		},
		Response: Response{
			Status: "done",
		},
	}
}

// Handle a serverless request
func Handle(req []byte) string {

	var (
		response Response
		request  NewRequestBody
	)

	defer func() {
		responseByte, _ := json.Marshal(response)
		Send(string(responseByte))
	}()

	err := json.Unmarshal(req, &request)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling request", "error": err.Error()}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	Send(fmt.Sprintf("request coming %v", req))

	currentTime := time.Now()

	layout := "02.01.2006 15:04"

	gte := currentTime

	createtObjectRequest := DateFilter{
		Gte:      strings.Split(gte.Format(layout), " ")[0] + `\t00:00`,
		ClientId: request.Data.ObjectData["cleints_id"].(string),
	}

	res, response, err := GetListSlimObject(SlimFunctionRequest{
		BaseUrl:   baseUrl,
		TableSlug: "walk",
		AppId:     request.Data.AppId,
		DateFilter: createtObjectRequest,
		DisableFaas: true,
	})

	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling request", "error": err.Error()}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}

	ids := []string{}
	for _, v := range res.Data.Data.Response {
		ids = append(ids, v.GUID)
	}

	if len(ids) != 0 {
		_, err = MultipleDelete(FunctionRequest{
			BaseUrl:     baseUrl,
			TableSlug:   "walk",
			AppId:       request.Data.AppId,
			Request:     Request{Data: map[string]interface{}{"ids": ids}},
			DisableFaas: true,
		})
		if err != nil {
			response.Data = map[string]interface{}{"message": "Error while deleting clients previous walks", "error": err.Error()}
			response.Status = "error"
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
	}

	responseByte, _ := json.Marshal(res)

	return string(responseByte)
}

func GetListObject(in FunctionRequest) (GetListClientApiResponse, Response, error) {
	response := Response{}

	var getListObject GetListClientApiResponse
	getListResponseInByte, err := DoRequest(fmt.Sprintf("%s/v1/object/get-list/%s?from-ofs=%t", in.BaseUrl, in.TableSlug, in.DisableFaas), "POST", in.Request, in.AppId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting list object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("error")
	}
	err = json.Unmarshal(getListResponseInByte, &getListObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("error")
	}
	return getListObject, response, nil
}

func GetListSlimObject(in SlimFunctionRequest) (AutoGenerated, Response, error) {
	response := Response{}

	var getListSlimObject AutoGenerated
	url := fmt.Sprintf(`%s/v1/object-slim/get-list/%s?data={"cleints_id":"%s","date":{"$gte":"%s"}}`, in.BaseUrl, in.TableSlug, in.DateFilter.ClientId, in.DateFilter.Gte)
	getListSlimResponseInByte, err := DoRequest(url, "GET", nil, in.AppId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting list slim object", "error": err.Error()}
		response.Status = "error"
		return AutoGenerated{}, response, errors.New("error")
	}

	err = json.Unmarshal(getListSlimResponseInByte, &getListSlimObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list slim object", "error": err.Error()}
		response.Status = "error"
		return AutoGenerated{}, response, errors.New("error")
	}

	return getListSlimObject, response, nil
}

// Send to Telegram Bot for logging
func Send(msg string) {
	if !IsHTTP {
		return
	}

	if botToken == "" {
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	req, _ := http.NewRequest("POST", url, nil)
	query := req.URL.Query()
	query.Add("chat_id", chatID)
	query.Add("text", msg)
	req.URL.RawQuery = query.Encode()
	_, _ = http.DefaultClient.Do(req)
}
