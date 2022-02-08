package iam

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"net/url"
	"strings"
)

func GetSignupUrl(enablePassword bool, redirectUri string) string {
	// redirectUri can be empty string if enablePassword == true (only password
	// enabled signup page is required)
	if enablePassword {
		return fmt.Sprintf("%s/signup/%s", authConfig.Endpoint, authConfig.ApplicationName)
	} else {
		return strings.ReplaceAll(GetSigninUrl(redirectUri), "/login/oauth/authorize", "/signup/oauth/authorize")
	}
}

func GetSigninUrl(redirectUri string) string {
	// origin := "https://iam.bhojpur.net"
	// redirectUri := fmt.Sprintf("%s/callback", origin)
	scope := "read"
	state := authConfig.ApplicationName
	return fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s",
		authConfig.Endpoint, authConfig.ClientId, url.QueryEscape(redirectUri), scope, state)
}

func GetUserProfileUrl(userName string, accessToken string) string {
	param := ""
	if accessToken != "" {
		param = fmt.Sprintf("?access_token=%s", accessToken)
	}
	return fmt.Sprintf("%s/users/%s/%s%s", authConfig.Endpoint, authConfig.OrganizationName, userName, param)
}

func GetMyProfileUrl(accessToken string) string {
	param := ""
	if accessToken != "" {
		param = fmt.Sprintf("?access_token=%s", accessToken)
	}
	return fmt.Sprintf("%s/account%s", authConfig.Endpoint, param)
}
