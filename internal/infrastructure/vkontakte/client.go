package vkontakte

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"parser/internal/domain/model"
)

const (
	accountTypeGroup   = "group"
	accountTypeProfile = "profile"
)

var ErrAccountNotFound = errors.New("vk account not found")

type Client struct {
	httpClient  *http.Client
	baseURL     string
	apiVersion  string
	accessToken string
}

func NewClient(baseURL, apiVersion, accessToken string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		httpClient:  &http.Client{Timeout: timeout},
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiVersion:  apiVersion,
		accessToken: accessToken,
	}
}

func (c *Client) GetAccountInfo(ctx context.Context, username string) (model.VKAccount, error) {
	username = cleanUsername(username)
	if username == "" {
		return model.VKAccount{}, ErrAccountNotFound
	}

	group, err := c.getGroup(ctx, username)
	if err == nil {
		return group, nil
	}
	if !errors.Is(err, ErrAccountNotFound) {
		return model.VKAccount{}, err
	}

	profile, err := c.getProfile(ctx, username)
	if err == nil {
		return profile, nil
	}
	return model.VKAccount{}, err
}

func (c *Client) getGroup(ctx context.Context, username string) (model.VKAccount, error) {
	params := url.Values{}
	params.Set("group_ids", username)
	params.Set("fields", "screen_name,name,description,is_closed,members_count,verified,photo_200")

	var response groupsGetByIDResponse
	if err := c.get(ctx, "groups.getById", params, &response); err != nil {
		return model.VKAccount{}, err
	}
	if response.Error != nil {
		if response.Error.Code == 100 || response.Error.Code == 15 {
			return model.VKAccount{}, ErrAccountNotFound
		}
		return model.VKAccount{}, fmt.Errorf("vk groups.getById: %s", response.Error.Message)
	}
	if len(response.Response) == 0 {
		return model.VKAccount{}, ErrAccountNotFound
	}

	item := response.Response[0]
	return model.VKAccount{
		SocialID:       item.ID,
		AccountType:    accountTypeGroup,
		FullName:       item.Name,
		Username:       fallbackUsername(item.ScreenName, username),
		FollowersCount: item.MembersCount,
		AvatarURL:      item.Photo200,
		Private:        item.IsClosed != 0,
		Verified:       item.Verified != 0,
	}, nil
}

func (c *Client) getProfile(ctx context.Context, username string) (model.VKAccount, error) {
	params := url.Values{}
	params.Set("user_ids", username)
	params.Set("fields", "screen_name,first_name,last_name,is_closed,followers_count,verified,photo_200")

	var response usersGetResponse
	if err := c.get(ctx, "users.get", params, &response); err != nil {
		return model.VKAccount{}, err
	}
	if response.Error != nil {
		if response.Error.Code == 100 || response.Error.Code == 15 || response.Error.Code == 18 {
			return model.VKAccount{}, ErrAccountNotFound
		}
		return model.VKAccount{}, fmt.Errorf("vk users.get: %s", response.Error.Message)
	}
	if len(response.Response) == 0 || response.Response[0].Deactivated != "" {
		return model.VKAccount{}, ErrAccountNotFound
	}

	item := response.Response[0]
	return model.VKAccount{
		SocialID:       item.ID,
		AccountType:    accountTypeProfile,
		FullName:       strings.TrimSpace(item.FirstName + " " + item.LastName),
		Username:       fallbackUsername(item.ScreenName, username),
		FollowersCount: item.FollowersCount,
		AvatarURL:      item.Photo200,
		Private:        item.IsClosed,
		Verified:       item.Verified != 0,
	}, nil
}

func (c *Client) get(ctx context.Context, method string, params url.Values, target any) error {
	params.Set("v", c.apiVersion)
	if c.accessToken != "" {
		params.Set("access_token", c.accessToken)
	}

	endpoint := c.baseURL + "/" + method + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("vk %s returned http %d", method, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode vk %s response: %w", method, err)
	}
	return nil
}

func cleanUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.TrimPrefix(username, "https://vk.com/")
	username = strings.TrimPrefix(username, "http://vk.com/")
	username = strings.TrimPrefix(username, "vk.com/")
	username = strings.Trim(username, "/")
	return username
}

func fallbackUsername(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

type vkError struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_msg"`
}

type groupsGetByIDResponse struct {
	Response []vkGroup `json:"response"`
	Error    *vkError  `json:"error"`
}

type vkGroup struct {
	ID           int64  `json:"id"`
	ScreenName   string `json:"screen_name"`
	Name         string `json:"name"`
	IsClosed     int    `json:"is_closed"`
	MembersCount int    `json:"members_count"`
	Verified     int    `json:"verified"`
	Photo200     string `json:"photo_200"`
}

type usersGetResponse struct {
	Response []vkUser `json:"response"`
	Error    *vkError `json:"error"`
}

type vkUser struct {
	ID             int64  `json:"id"`
	ScreenName     string `json:"screen_name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	IsClosed       bool   `json:"is_closed"`
	FollowersCount int    `json:"followers_count"`
	Verified       int    `json:"verified"`
	Photo200       string `json:"photo_200"`
	Deactivated    string `json:"deactivated"`
}
