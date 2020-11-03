package http

import (
	"GameNerdzMonitor/models"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	RequestTypeGET  = "GET"
	RequestTypePOST = "POST"

	HomepageURL   = "https://www.gamenerdz.com/"
	GetProductURL = "https://www.gamenerdz.com/remote/v1/product-attributes/"

	DiscordWebhookURL = "https://discord.com/api/webhooks/767776124251930634/c88jDv0m759Cn5EPK7d7junmSjo2zvx8FfRgnMJec2oJNEgjqJXJhp7ZuDNZkGsng4Pq"
)

type Client struct {
	client *http.Client
	logger *zap.Logger
	cookie string
	user   string
	pass   string
}

func InitializeClient(proxy models.Proxy, logger *zap.Logger) (Client, error) {
	c := Client{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&url.URL{
					Scheme: "http",
					User:   url.UserPassword(proxy.Un, proxy.Pw),
					Host:   proxy.Host,
				}),
			},
		},
		logger: logger,
		user:   proxy.Un,
		pass:   proxy.Pw,
	}

	cookie, err := c.GetCookies()
	if err != nil {
		return c, err
	}

	c.cookie = cookie

	return c, nil
}

func (c *Client) GetCookies() (string, error) {
	req, err := c.createRequest(RequestTypeGET, HomepageURL, "", false)
	if err != nil {
		c.logger.Error("GetCookies - httpNewRequest", zap.Error(err))
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("GetCookies - sendRequest", zap.Error(err))
		return "", err
	}

	cookie := resp.Header.Get("set-cookie")
	fmt.Println(cookie)

	if cookie == "" {
		return "", errors.New("No cookie generated")
	}

	return cookie, nil
}

func (c *Client) GetProductAvailability(sku string) (bool, models.Product, error) {
	getProductURL := fmt.Sprintf(`%s%s`, GetProductURL, sku)
	req, err := c.createRequest(RequestTypeGET, getProductURL, "", true)
	if err != nil {
		c.logger.Error("GetProductAvailability - httpNewRequest", zap.Error(err))
		return false, models.Product{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("GetProductAvailability - sendRequest", zap.Error(err))
		return false, models.Product{}, err
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("GetProductAvailability - readResponseBody", zap.Error(err))
		return false, models.Product{}, err
	}
	fmt.Println(string(bodyText))

	var product models.Product
	err = json.Unmarshal(bodyText, &product)
	if err != nil {
		c.logger.Error("GetProductAvailability - unmarshal", zap.Error(err))
		return false, models.Product{}, err
	}

	productIsInStock := product.Data.Instock
	if productIsInStock {
		return true, product, nil
	}

	return false, product, nil
}

func (c *Client) SendDiscordMessage(body models.Message) (int, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		c.logger.Error("SendDiscordMessage - marshalProfile", zap.Error(err))
		return 0, err
	}

	req, err := c.createRequest(RequestTypePOST, DiscordWebhookURL, string(requestBody), false)
	if err != nil {
		c.logger.Error("SendDiscordMessage - httpNewRequest", zap.Error(err))
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("SendDiscordMessage - sendRequest", zap.Error(err))
		return 0, err
	}

	// bodyText, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	c.logger.Error("SendDiscordMessage - readResponseBody", zap.Error(err))
	// 	return 0, err
	// }

	// var buf *bytes.Buffer = new(bytes.Buffer)

	// fmt.Println(string(bodyText))
	// resp.Header.Write(buf)
	// lines := strings.Split(buf.String(), "\n")
	// if len(lines) > 0 {
	// 	fmt.Println("Response headers:")
	// 	for _, line := range lines {
	// 		fmt.Printf("\t%s\n", line)
	// 	}
	// }
	// buf.Reset()

	// fmt.Println(retryTimer)
	// fmt.Println(resp.StatusCode)
	// fmt.Println(http.StatusText(resp.StatusCode))

	retryTimer := resp.Header.Get("Retry-After")
	if retryTimer == "" {
		return 0, nil
	} else {
		retry, _ := strconv.Atoi(retryTimer)
		return retry, nil
	}
}

func (c *Client) createRequest(requestType, url, body string, withCookies bool) (*http.Request, error) {
	var req *http.Request
	var err error

	if requestType == RequestTypeGET {
		req, err = http.NewRequest(requestType, url, nil)
	} else {
		req, err = http.NewRequest(requestType, url, strings.NewReader(body))
	}
	if err != nil {
		return nil, err
	}

	return c.setHeaders(req, withCookies), nil
}

func (c *Client) setHeaders(req *http.Request, withCookies bool) *http.Request {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:80.0) Gecko/20100101 Firefox/80.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Store-Scope", "pokemon")
	req.Header.Set("Origin", "https://www.gamenerdz.com")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Te", "Trailers")

	if withCookies {
		req.Header.Set("Cookie", c.cookie)
	}

	return req
}

// req.Header.Set("X-Xsrf-Token", "6c299956a35972a46e074fcc5be0c0961c96eb82eaec070d9d07654f5043de68, 6c299956a35972a46e074fcc5be0c0961c96eb82eaec070d9d07654f5043de68")
