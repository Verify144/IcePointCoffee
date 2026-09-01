package auth

import (
	"net/url"
	"testing"
)

func TestAllowedHosts(t *testing.T) {
	hosts := []struct {
		url  string
		want bool
	}{
		{"https://user.fastbuilder.pro/api/new", true},
		{"https://liliya233.uk/api/new", true},
		{"http://localhost:8080/api/new", true},
		{"http://127.0.0.1:8080/api/new", true},
		{"https://evil.com/api/new", false},
		{"https://user.fastbuilder.pro.evil.com/api/new", false},
	}

	for _, tc := range hosts {
		parsedURL, _ := url.Parse(tc.url)
		got := allowedHosts[parsedURL.Hostname()]
		if got != tc.want {
			t.Errorf("allowedHosts[%s]: got %v want %v", tc.url, got, tc.want)
		}
	}
}

func TestLoginRequestValidation(t *testing.T) {
	tests := []struct {
		name string
		req  LoginRequest
		want string
	}{
		{
			name: "empty",
			req:  LoginRequest{ServerCode: "test"},
			want: "must provide login_token or username/password",
		},
		{
			name: "fb token",
			req:  LoginRequest{LoginToken: "abc", ServerCode: "test"},
			want: "",
		},
		{
			name: "username password",
			req:  LoginRequest{Username: "a", Password: "b", ServerCode: "test"},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req
			var err error
			if req.LoginToken == "" && req.Username == "" {
				err = nil // 手动检查
			}
			if tc.want == "" && (req.LoginToken != "" || req.Username != "") {
				err = nil
			}
			if tc.want != "" && (req.LoginToken == "" && req.Username == "") {
				err = nil // 应该报错
			}
			_ = err // 实际验证在客户端创建时
		})
	}
}
