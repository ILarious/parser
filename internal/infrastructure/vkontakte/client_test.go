package vkontakte

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGetAccountInfoGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups.getById" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(groupsGetByIDResponse{Response: []vkGroup{{
			ID:           1,
			ScreenName:   "team",
			Name:         "Team",
			MembersCount: 42,
			Verified:     1,
			Photo200:     "https://example.test/photo.jpg",
		}}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "5.130", "token", time.Second)
	account, err := client.GetAccountInfo(context.Background(), "https://vk.com/team")
	if err != nil {
		t.Fatalf("GetAccountInfo() error = %v", err)
	}
	if account.AccountType != "group" || account.FullName != "Team" || account.FollowersCount != 42 {
		t.Fatalf("unexpected account: %+v", account)
	}
}

func TestClientGetAccountInfoFallsBackToProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/groups.getById":
			_ = json.NewEncoder(w).Encode(groupsGetByIDResponse{Error: &vkError{Code: 100, Message: "not found"}})
		case "/users.get":
			_ = json.NewEncoder(w).Encode(usersGetResponse{Response: []vkUser{{
				ID:             1,
				ScreenName:     "durov",
				FirstName:      "Pavel",
				LastName:       "Durov",
				FollowersCount: 100,
			}}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "5.130", "", time.Second)
	account, err := client.GetAccountInfo(context.Background(), "durov")
	if err != nil {
		t.Fatalf("GetAccountInfo() error = %v", err)
	}
	if account.AccountType != "profile" || account.FullName != "Pavel Durov" || account.FollowersCount != 100 {
		t.Fatalf("unexpected account: %+v", account)
	}
}
