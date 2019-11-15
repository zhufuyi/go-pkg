package gohttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Request HTTP request
type Request struct {
	customRequest func(req *http.Request, data *bytes.Buffer) // 用于定义HEADER, 如添加sign等
	url           string
	params        map[string]interface{} // URL后的参数
	body          string                 // Body数据
	bodyJSON      interface{}            // 可JSON Marshal 的Body的数据
	timeout       time.Duration          // Client timeout
	headers       map[string]string

	request  *http.Request
	response *Response
	method   string
	err      error

	jaegerTrace func() error // jaeger调用链跟踪
}

// Response HTTP response
type Response struct {
	*http.Response
	err error
}

// -----------------------------------  Request  -----------------------------------

// Reset set all fields to default value, use at pool
func (req *Request) Reset() {
	req.params = nil
	req.body = ""
	req.bodyJSON = nil
	req.timeout = 0
	req.headers = nil

	req.request = nil
	req.response = nil
	req.method = ""
	req.err = nil
	req.jaegerTrace = nil
}

// SetURL 设置URL
func (req *Request) SetURL(path string) *Request {
	req.url = path
	return req
}

// SetParams 设置URL后的参数
func (req *Request) SetParams(params map[string]interface{}) *Request {
	if req.params == nil {
		req.params = params
	} else {
		for k, v := range params {
			req.params[k] = v
		}
	}
	return req
}

// SetParam 设置URL后的参数
func (req *Request) SetParam(k string, v interface{}) *Request {
	if req.params == nil {
		req.params = make(map[string]interface{})
	}
	req.params[k] = v
	return req
}

// SetBody 设置Body数据
func (req *Request) SetBody(body string) *Request {
	req.body = body
	return req
}

// SetJSONBody 设置Body数据, JSON格式
func (req *Request) SetJSONBody(body interface{}) *Request {
	req.bodyJSON = body
	return req
}

// SetTimeout 超时时间
func (req *Request) SetTimeout(t time.Duration) *Request {
	req.timeout = t
	return req
}

// SetContentType 设置ContentType
func (req *Request) SetContentType(a string) *Request {
	req.SetHeader("Content-Type", a)
	return req
}

// SetHeader 设置Request Header 的值
func (req *Request) SetHeader(k, v string) *Request {
	if req.headers == nil {
		req.headers = make(map[string]string)
	}
	req.headers[k] = v
	return req
}

// SetJaegerTrace 是否开启open tracing功能
func (req *Request) SetJaegerTrace(traceID string, ginKeys map[string]interface{}) *Request {
	// 通过gin.Context的keys字段传进来字段内容
	if isEnable, ok := ginKeys["isEnableTrace"]; !ok || !isEnable.(bool) {
		return req
	}

	vTracer := ginKeys["tracer"]
	vParentSpan := ginKeys["parentSpan"]

	req.jaegerTrace = func() error {
		if vTracer == nil {
			return errors.New("tracer is nil, maybe it hasn't been initialized yet.")
		}
		tracer := vTracer.(opentracing.Tracer)

		if vParentSpan == nil {
			return errors.New("jaeger.ParentSpan is nil")
		}
		parentSpan := vParentSpan.(opentracing.Span)

		if traceID != "" {
			req.SetHeader("Uber-Trace-Id", traceID)
			req.request.Header.Add("Uber-Trace-Id", traceID)
		}

		span := opentracing.StartSpan("http client call",
			opentracing.ChildOf(parentSpan.Context()),
			opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
			opentracing.Tag{Key: string(ext.HTTPMethod), Value: req.method},
			opentracing.Tag{Key: string(ext.HTTPUrl), Value: req.url},
			ext.SpanKindRPCClient,
		)

		span.Finish()

		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.request.Header))
		if err != nil {
			return err
		}

		return nil
	}

	return req
}

// CustomRequest 自定义Request, 如添加sign, 设置header等
func (req *Request) CustomRequest(f func(req *http.Request, data *bytes.Buffer)) *Request {
	req.customRequest = f
	return req
}

// GET 发送GET请求
func (req *Request) GET() (*Response, error) {
	req.method = "GET"
	return req.pull()
}

// DELETE 发送DELETE请求
func (req *Request) DELETE() (*Response, error) {
	req.method = "DELETE"
	return req.pull()
}

// POST 发送POST请求
func (req *Request) POST() (*Response, error) {
	req.method = "POST"
	return req.push()
}

// PUT 发送PUT请求
func (req *Request) PUT() (*Response, error) {
	req.method = "PUT"
	return req.push()
}

// PATCH 发送PATCH请求
func (req *Request) PATCH() (*Response, error) {
	req.method = "PATCH"
	return req.push()
}

// Do a request
func (req *Request) Do(method string, data interface{}) (*Response, error) {
	req.method = method

	switch method {
	case "GET", "DELETE":
		if data != nil {
			if params, ok := data.(map[string]interface{}); ok {
				req.SetParams(params)
			} else {
				req.err = errors.New("params is not a map[string]interface{}")
				return nil, req.err
			}
		}

		return req.pull()

	case "POST", "PUT", "PATCH":
		if data != nil {
			req.SetJSONBody(data)
		}

		return req.push()
	}

	req.err = errors.New("unknow method " + method)
	return nil, req.err
}

func (req *Request) pull() (*Response, error) {
	// 添加

	val := ""
	if len(req.params) > 0 {
		values := url.Values{}
		for k, v := range req.params {
			values.Add(k, fmt.Sprintf("%v", v))
		}
		val += values.Encode()
	}

	if val != "" {
		if strings.Contains(req.url, "?") {
			req.url += "&" + val
		} else {
			req.url += "?" + val
		}
	}

	var buf *bytes.Buffer
	if req.customRequest != nil {
		buf = bytes.NewBufferString(val)
	}

	return req.send(nil, buf)
}

func (req *Request) push() (*Response, error) {
	var buf = new(bytes.Buffer)

	if req.bodyJSON != nil {
		body, err := json.Marshal(req.bodyJSON)
		if err != nil {
			req.err = err
			return nil, req.err
		}

		buf = bytes.NewBuffer(body)

	} else {
		buf = bytes.NewBufferString(req.body)
	}

	return req.send(buf, buf)
}

func (req *Request) send(body io.Reader, buf *bytes.Buffer) (*Response, error) {
	req.request, req.err = http.NewRequest(req.method, req.url, body)
	if req.err != nil {
		return nil, req.err
	}

	if req.customRequest != nil {
		req.customRequest(req.request, buf)
	}

	if req.headers != nil {
		for k, v := range req.headers {
			req.request.Header.Add(k, v)
		}
	}

	if req.timeout < 1 {
		req.timeout = 1 * time.Minute
	}

	// 跟踪信息
	if req.jaegerTrace != nil {
		if err := req.jaegerTrace(); err != nil {
			fmt.Printf("trace Info error: %s\n", err.Error())
		}
	}

	client := http.Client{Timeout: req.timeout}

	resp := new(Response)
	resp.Response, resp.err = client.Do(req.request)

	req.response = resp
	req.err = resp.err

	return resp, resp.err
}

// Response return response
func (req *Request) Response() (*Response, error) {
	if req.err != nil {
		return nil, req.err
	}
	return req.response, req.response.Error()
}

// -----------------------------------  Response -----------------------------------

// Error return err
func (resp *Response) Error() error {
	return resp.err
}

// BodyString 返回HttpResponse的body数据
func (resp *Response) BodyString() (string, error) {
	if resp.err != nil {
		return "", resp.err
	}
	body, err := resp.ReadBody()
	return string(body), err
}

// ReadBody 返回HttpResponse的body数据
func (resp *Response) ReadBody() ([]byte, error) {
	if resp.err != nil {
		return []byte{}, resp.err
	}

	if resp.Response == nil {
		return []byte{}, errors.New("nil")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// BindJSON parses the response's body as JSON
func (resp *Response) BindJSON(v interface{}) error {
	if resp.err != nil {
		return resp.err
	}
	body, err := resp.ReadBody()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// -------------------------------------------------------------------------------------------------

// JSONResponse 输出格式
type JSONResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// HTTPGetJson get请求，返回是固定json格式
func HTTPGetJson(url string, params map[string]interface{}) (*JSONResponse, error) {
	req := &Request{}
	req.SetURL(url)
	req.SetParams(params)

	resp, err := req.GET()
	if err != nil {
		return nil, fmt.Errorf("get请求失败, error=%s, req=%+v", err.Error(), *req)
	}
	defer resp.Body.Close()

	return parseRespone(resp)
}

// HTTPPostJson post请求，返回是固定json格式
func HTTPPostJson(url string, body interface{}, params ...map[string]interface{}) (*JSONResponse, error) {
	req := &Request{}
	req.SetURL(url) // url地址固定
	req.SetContentType("application/json")
	if len(params) > 0 {
		req.SetParams(params[0])
	}
	req.SetJSONBody(body)

	resp, err := req.POST()
	if err != nil {
		return nil, fmt.Errorf("post请求失败, error=%s, req=%+v", err.Error(), *req)
	}
	defer resp.Body.Close()

	return parseRespone(resp)
}

func parseRespone(resp *Response) (*JSONResponse, error) {
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("返回状态非200, StatusCode=%d", resp.StatusCode)
	}

	result := &JSONResponse{}
	err := resp.BindJSON(result)
	if err != nil {
		return nil, fmt.Errorf("解析参数错误, error=%s, req=%+v", err.Error(), *resp)
	}

	return result, nil
}
