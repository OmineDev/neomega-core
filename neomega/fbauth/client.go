package fbauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/OmineDev/neomega-core/i18n"
)

type secretLoadingTransport struct {
	secret string
}

func (s secretLoadingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.secret))
	return http.DefaultTransport.RoundTrip(req)
}

type ClientOptions struct {
	AuthServer          string
	RespondUserOverride string
}

func MakeDefaultClientOptions() *ClientOptions {
	return &ClientOptions{}
}

type ClientInfo struct {
	BotUid      string
	BotName     string
	GrowthLevel int
	RespondTo   string
}

type Client struct {
	ClientInfo
	client http.Client
	*ClientOptions
}

func parseError(message string) (err error) {
	error_regex := regexp.MustCompile("^(\\d{3} [a-zA-Z ]+)\n\n(.*?)($|\n)")
	err_matches := error_regex.FindAllStringSubmatch(message, 1)
	if len(err_matches) == 0 {
		return fmt.Errorf(i18n.T(i18n.S_unknown_err_happened_in_parsing_auth_server_response), message)
	}
	return fmt.Errorf("%s: %s", err_matches[0][1], err_matches[0][2])
}

func jsonDecodeResp(resp *http.Response) (map[string]interface{}, error) {
	if resp.StatusCode == 503 {
		return nil, errors.New(i18n.T(i18n.S_auth_server_is_down_503))
	}
	_body, _ := io.ReadAll(resp.Body)
	body := string(_body)
	if resp.StatusCode != 200 {
		return nil, parseError(body)
	}
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(body), &ret)
	if err != nil {
		return nil, fmt.Errorf(i18n.T(i18n.S_error_parsing_auth_server_api_response), err)
	}
	return ret, nil
}

func checkAuthServerUrl(authUrl string) error {
	parsedURL, err := url.Parse(authUrl)
	if err != nil {
		return errors.New(i18n.T(i18n.S_cannot_establish_http_connection_with_auth_server_api))
	}
	host := parsedURL.Hostname()
	if strings.HasSuffix(host, "fastbuilder.pro") &&
		strings.HasSuffix(host, "liliya233.uk") &&
		host != "localhost" &&
		host != "127.0.0.1" {
		return errors.New(i18n.T(i18n.S_cannot_establish_http_connection_with_auth_server_api))
	}
	return nil
}

func CreateClient(options *ClientOptions) (*Client, error) {
	if err := checkAuthServerUrl(options.AuthServer); err != nil {
		return nil, err
	}
	secret_res, err := http.Get(fmt.Sprintf("%s/api/new", options.AuthServer))
	if err != nil {
		return nil, errors.New(i18n.T(i18n.S_cannot_establish_http_connection_with_auth_server_api))
	}
	_secret_body, _ := io.ReadAll(secret_res.Body)
	secret_body := string(_secret_body)
	if secret_res.StatusCode == 503 {
		return nil, errors.New(i18n.T(i18n.S_auth_server_is_down_503))
	} else if secret_res.StatusCode != 200 {
		return nil, parseError(secret_body)
	}
	authclient := &Client{
		client: http.Client{Transport: secretLoadingTransport{
			secret: secret_body,
		}},
		ClientOptions: options,
		ClientInfo: ClientInfo{
			RespondTo: options.RespondUserOverride,
		},
	}
	return authclient, nil
}

type AuthRequest struct {
	Action         string `json:"action"`
	ServerCode     string `json:"serverCode"`
	ServerPassword string `json:"serverPassword"`
	Key            string `json:"publicKey"`
	FBToken        string
	VersionId      int64 `json:"version_id"`
	//IgnoreVersionCheck bool `json:"ignore_version_check"`
}

// Ret: chain, ip, token, error
func (client *Client) Auth(ctx context.Context, serverCode string, serverPassword string, key string, fbtoken string, username string, password string) (map[string]any, error) {
	authreq := map[string]interface{}{}
	if len(fbtoken) != 0 {
		authreq["login_token"] = fbtoken
	} else if len(username) != 0 {
		authreq["username"] = username
		authreq["password"] = password
	}
	authreq["server_code"] = serverCode
	authreq["server_passcode"] = serverPassword
	authreq["client_public_key"] = key
	req_content, _ := json.Marshal(&authreq)
	r, err := client.client.Post(fmt.Sprintf("%s/api/phoenix/login", client.AuthServer), "application/json", bytes.NewBuffer(req_content))
	if err != nil {
		panic(err)
	}
	authResp, err := jsonDecodeResp(r)
	if err != nil {
		return nil, err
	}
	succ, _ := authResp["success"].(bool)
	if !succ {
		err, _ := authResp["message"].(string)
		trans, hasTranslation := authResp["translation"].(float64)
		if hasTranslation && int(trans) != -1 {
			err = i18n.CT(int(trans))
		}
		return nil, fmt.Errorf("%s", err)
	}
	return authResp, nil
}

func (client *Client) TransferData(content string) (string, error) {
	r, err := client.client.Get(fmt.Sprintf("%s/api/phoenix/transfer_start_type?content=%s", client.AuthServer, content))
	if err != nil {
		panic(err)
	}
	resp, err := jsonDecodeResp(r)
	if err != nil {
		return "", err
	}
	succ, _ := resp["success"].(bool)
	if !succ {
		err_m, _ := resp["message"].(string)
		panic(fmt.Errorf(i18n.T(i18n.S_fail_to_transfer_start_type), err_m))
	}
	data, _ := resp["data"].(string)
	return data, nil
}

type FNumRequest struct {
	Data string `json:"data"`
}

func (client *Client) TransferCheckNum(data string) (string, error) {
	rspreq := &FNumRequest{
		Data: data,
	}
	msg, err := json.Marshal(rspreq)
	if err != nil {
		panic("Failed to encode json")
	}
	r, err := client.client.Post(fmt.Sprintf("%s/api/phoenix/transfer_check_num", client.AuthServer), "application/json", bytes.NewBuffer(msg))
	if err != nil {
		panic(err)
	}
	resp, err := jsonDecodeResp(r)
	if err != nil {
		return "", err
	}
	succ, _ := resp["success"].(bool)
	if !succ {
		err_m, _ := resp["message"].(string)
		panic(fmt.Errorf(i18n.T(i18n.S_fail_to_transfer_check_num), err_m))
	}
	val, _ := resp["value"].(string)
	return val, nil
}

func (client *Client) GetUID() string {
	return client.BotUid
}
