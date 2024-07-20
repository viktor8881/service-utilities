package simplehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type SimpleClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewSimpleClient(baseURL string, timeout time.Duration, transport http.RoundTripper) *SimpleClient {
	return &SimpleClient{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		baseURL: baseURL,
	}
}

func (c *SimpleClient) Get(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)

		return resp, &ClientResponseNot200Error{
			ClientResponseCode: resp.StatusCode,
			ClientResponseBody: body.String(),
			Err:                errors.New("Response status code is not 2xx"),
		}
	}

	return resp, nil
}

func (c *SimpleClient) Delete(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)

		return resp, &ClientResponseNot200Error{
			ClientResponseCode: resp.StatusCode,
			ClientResponseBody: body.String(),
			Err:                errors.New("response status code is not 2xx"),
		}
	}

	return resp, nil
}

func (c *SimpleClient) Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
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

	return c.httpClient.Do(req)
}

func (c *SimpleClient) Put(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
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

	return c.httpClient.Do(req)
}

func (c *SimpleClient) Close() {
	c.httpClient.CloseIdleConnections()
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

	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		placeholder := "{" + fieldName + "}"
		placeholderLower := "{" + strings.ToLower(fieldName) + "}"
		fieldValue := fmt.Sprintf("%v", v.FieldByName(fieldName).Interface())
		template = strings.ReplaceAll(template, placeholder, fieldValue)
		template = strings.ReplaceAll(template, placeholderLower, fieldValue)
	}
	return template, nil
}
