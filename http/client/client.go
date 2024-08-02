package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string, timeout time.Duration, transport http.RoundTripper) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		baseURL: baseURL,
	}
}

func (c *Client) Get(ctx context.Context, endpoint string, in interface{}, headers map[string]string) (*http.Response, error) {
	pathUrl, err := BuildURL(c.baseURL+endpoint, in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pathUrl, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

func (c *Client) Delete(ctx context.Context, endpoint string, in interface{}, headers map[string]string) (*http.Response, error) {
	pathUrl, err := BuildURL(c.baseURL+endpoint, in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, pathUrl, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

func (c *Client) Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

func (c *Client) Put(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		return resp, &ClientResponseNot200Error{
			ClientResponseCode: resp.StatusCode,
			ClientResponseBody: string(body),
			Err:                errors.New("response status code is not 200"),
		}
	}

	return resp, nil
}

func BuildURL(template string, input interface{}) (string, error) {
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("input is not a struct or pointer to a struct")
	}

	isPathParam := strings.Contains(template, "{")
	if isPathParam {
		// path params /user/{id}/profile/{name}
		for i := 0; i < t.NumField(); i++ {
			fieldName := t.Field(i).Name
			placeholder := "{" + strings.ToLower(fieldName) + "}"
			fieldValue := fmt.Sprintf("%v", v.FieldByName(fieldName).Interface())
			template = strings.ReplaceAll(template, placeholder, fieldValue)
		}
	} else {
		// query params url?id=1&name=John
		dataUrl, err := url.Parse(template)
		if err != nil {
			return template, err
		}

		query := dataUrl.Query()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i).Interface()
			query.Add(strings.ToLower(field.Name), fmt.Sprintf("%v", value))
		}

		dataUrl.RawQuery = query.Encode()
		template = dataUrl.String()
	}

	return template, nil
}
