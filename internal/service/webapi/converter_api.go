package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	url "net/url"
	"os"
)

const rubles = "RUB"

type ConverterAPI struct {
	client *http.Client
}

func NewConverterAPI(client *http.Client) *ConverterAPI {
	return &ConverterAPI{client}
}

type responseHTTP struct {
	Result float64 `json:"result"`
}

func (c *ConverterAPI) ConvertToCurrency(ctx context.Context, currency string, amount float64) (float64, error) {
	converterUrl := os.Getenv("CONVERTER_URL")

	req, err := http.NewRequestWithContext(ctx, "GET", converterUrl, nil)
	req.Header.Set("apikey", os.Getenv("CONVERTER_API_KEY"))

	// add parameters
	req.URL.RawQuery = url.Values{
		"from":   {currency},
		"to":     {rubles},
		"amount": {"1"},
	}.Encode()

	if err != nil {
		return 0, fmt.Errorf("webapi - ConvertToCurrency - http.NewRequest: %w", err)
	}

	// do request
	res, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("webapi - ConvertToCurrency - c.client.Do: %w", err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("webapi - ConvertToCurrency - ioutil.ReadAll: %w", err)
	}

	// parse response and pull out the result
	r := bytes.NewReader(body)
	decoder := json.NewDecoder(r)

	var val responseHTTP
	err = decoder.Decode(&val)
	if err != nil {
		return 0, fmt.Errorf("webapi - ConvertToCurrency - decoder.Decode: %w", err)
	}

	log.Info(val.Result)

	return val.Result, nil
}
