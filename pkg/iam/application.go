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

import "encoding/json"

// Application has the same definition as Bhojpur IAM application object
type Application struct {
	Owner       string `orm:"varchar(100) notnull pk" json:"owner"`
	Name        string `orm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `orm:"varchar(100)" json:"createdTime"`

	DisplayName         string `orm:"varchar(100)" json:"displayName"`
	Logo                string `orm:"varchar(100)" json:"logo"`
	HomepageUrl         string `orm:"varchar(100)" json:"homepageUrl"`
	Description         string `orm:"varchar(100)" json:"description"`
	Organization        string `orm:"varchar(100)" json:"organization"`
	Cert                string `orm:"varchar(100)" json:"cert"`
	EnablePassword      bool   `json:"enablePassword"`
	EnableSignUp        bool   `json:"enableSignUp"`
	EnableSigninSession bool   `json:"enableSigninSession"`
	EnableCodeSignin    bool   `json:"enableCodeSignin"`

	ClientId             string   `orm:"varchar(100)" json:"clientId"`
	ClientSecret         string   `orm:"varchar(100)" json:"clientSecret"`
	RedirectUris         []string `orm:"varchar(1000)" json:"redirectUris"`
	TokenFormat          string   `orm:"varchar(100)" json:"tokenFormat"`
	ExpireInHours        int      `json:"expireInHours"`
	RefreshExpireInHours int      `json:"refreshExpireInHours"`
	SignupUrl            string   `orm:"varchar(200)" json:"signupUrl"`
	SigninUrl            string   `orm:"varchar(200)" json:"signinUrl"`
	ForgetUrl            string   `orm:"varchar(200)" json:"forgetUrl"`
	AffiliationUrl       string   `orm:"varchar(100)" json:"affiliationUrl"`
	TermsOfUse           string   `orm:"varchar(100)" json:"termsOfUse"`
	SignupHtml           string   `orm:"mediumtext" json:"signupHtml"`
	SigninHtml           string   `orm:"mediumtext" json:"signinHtml"`
}

func AddApplication(application *Application) (bool, error) {
	if application.Owner == "" {
		application.Owner = "admin"
	}
	postBytes, err := json.Marshal(application)
	if err != nil {
		return false, err
	}

	resp, err := doPost("add-application", nil, postBytes, false)
	if err != nil {
		return false, err
	}

	return resp.Data == "Affected", nil
}

func DeleteApplication(name string) (bool, error) {
	application := Application{
		Owner: "admin",
		Name:  name,
	}
	postBytes, err := json.Marshal(application)
	if err != nil {
		return false, err
	}

	resp, err := doPost("delete-application", nil, postBytes, false)
	if err != nil {
		return false, err
	}

	return resp.Data == "Affected", nil
}
