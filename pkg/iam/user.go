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
	"encoding/json"
	"fmt"
	"strconv"
)

// User has the same definition as Bhojpur IAM user object
type User struct {
	Owner       string `orm:"varchar(100) notnull pk" json:"owner"`
	Name        string `orm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `orm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `orm:"varchar(100)" json:"updatedTime"`

	Id                string   `orm:"varchar(100)" json:"id"`
	Type              string   `orm:"varchar(100)" json:"type"`
	Password          string   `orm:"varchar(100)" json:"password"`
	PasswordSalt      string   `orm:"varchar(100)" json:"passwordSalt"`
	DisplayName       string   `orm:"varchar(100)" json:"displayName"`
	Avatar            string   `orm:"varchar(255)" json:"avatar"`
	PermanentAvatar   string   `orm:"varchar(255)" json:"permanentAvatar"`
	Email             string   `orm:"varchar(100)" json:"email"`
	Phone             string   `orm:"varchar(100)" json:"phone"`
	Location          string   `orm:"varchar(100)" json:"location"`
	Address           []string `json:"address"`
	Affiliation       string   `orm:"varchar(100)" json:"affiliation"`
	Title             string   `orm:"varchar(100)" json:"title"`
	IdCardType        string   `orm:"varchar(100)" json:"idCardType"`
	IdCard            string   `orm:"varchar(100)" json:"idCard"`
	Homepage          string   `orm:"varchar(100)" json:"homepage"`
	Bio               string   `orm:"varchar(100)" json:"bio"`
	Tag               string   `orm:"varchar(100)" json:"tag"`
	Region            string   `orm:"varchar(100)" json:"region"`
	Language          string   `orm:"varchar(100)" json:"language"`
	Gender            string   `orm:"varchar(100)" json:"gender"`
	Birthday          string   `orm:"varchar(100)" json:"birthday"`
	Education         string   `orm:"varchar(100)" json:"education"`
	Score             int      `json:"score"`
	Ranking           int      `json:"ranking"`
	IsDefaultAvatar   bool     `json:"isDefaultAvatar"`
	IsOnline          bool     `json:"isOnline"`
	IsAdmin           bool     `json:"isAdmin"`
	IsGlobalAdmin     bool     `json:"isGlobalAdmin"`
	IsForbidden       bool     `json:"isForbidden"`
	IsDeleted         bool     `json:"isDeleted"`
	SignupApplication string   `orm:"varchar(100)" json:"signupApplication"`
	Hash              string   `orm:"varchar(100)" json:"hash"`
	PreHash           string   `orm:"varchar(100)" json:"preHash"`

	CreatedIp      string `orm:"varchar(100)" json:"createdIp"`
	LastSigninTime string `orm:"varchar(100)" json:"lastSigninTime"`
	LastSigninIp   string `orm:"varchar(100)" json:"lastSigninIp"`

	Github   string `orm:"varchar(100)" json:"github"`
	Google   string `orm:"varchar(100)" json:"google"`
	QQ       string `orm:"qq varchar(100)" json:"qq"`
	WeChat   string `orm:"wechat varchar(100)" json:"wechat"`
	Facebook string `orm:"facebook varchar(100)" json:"facebook"`
	DingTalk string `orm:"dingtalk varchar(100)" json:"dingtalk"`
	Weibo    string `orm:"weibo varchar(100)" json:"weibo"`
	Gitee    string `orm:"gitee varchar(100)" json:"gitee"`
	LinkedIn string `orm:"linkedin varchar(100)" json:"linkedin"`
	Wecom    string `orm:"wecom varchar(100)" json:"wecom"`
	Lark     string `orm:"lark varchar(100)" json:"lark"`
	Gitlab   string `orm:"gitlab varchar(100)" json:"gitlab"`

	Ldap       string            `orm:"ldap varchar(100)" json:"ldap"`
	Properties map[string]string `json:"properties"`
}

func GetUsers() ([]*User, error) {
	queryMap := map[string]string{
		"owner": authConfig.OrganizationName,
	}

	url := getUrl("get-users", queryMap)

	bytes, err := doGetBytes(url)
	if err != nil {
		return nil, err
	}

	var users []*User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetSortedUsers(sorter string, limit int) ([]*User, error) {
	queryMap := map[string]string{
		"owner":  authConfig.OrganizationName,
		"sorter": sorter,
		"limit":  strconv.Itoa(limit),
	}

	url := getUrl("get-sorted-users", queryMap)

	bytes, err := doGetBytes(url)
	if err != nil {
		return nil, err
	}

	var users []*User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserCount(isOnline string) (int, error) {
	queryMap := map[string]string{
		"owner":    authConfig.OrganizationName,
		"isOnline": isOnline,
	}

	url := getUrl("get-user-count", queryMap)

	bytes, err := doGetBytes(url)
	if err != nil {
		return -1, err
	}

	var count int
	err = json.Unmarshal(bytes, &count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func GetUser(name string) (*User, error) {
	queryMap := map[string]string{
		"id": fmt.Sprintf("%s/%s", authConfig.OrganizationName, name),
	}

	url := getUrl("get-user", queryMap)

	bytes, err := doGetBytes(url)
	if err != nil {
		return nil, err
	}

	var user *User
	err = json.Unmarshal(bytes, &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByEmail(email string) (*User, error) {
	queryMap := map[string]string{
		"owner": authConfig.OrganizationName,
		"email": email,
	}

	url := getUrl("get-user", queryMap)

	bytes, err := doGetBytes(url)
	if err != nil {
		return nil, err
	}

	var user *User
	err = json.Unmarshal(bytes, &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUser(user *User) (bool, error) {
	_, affected, err := modifyUser("update-user", user, nil)
	return affected, err
}

func UpdateUserForColumns(user *User, columns []string) (bool, error) {
	_, affected, err := modifyUser("update-user", user, columns)
	return affected, err
}

func AddUser(user *User) (bool, error) {
	_, affected, err := modifyUser("add-user", user, nil)
	return affected, err
}

func DeleteUser(user *User) (bool, error) {
	_, affected, err := modifyUser("delete-user", user, nil)
	return affected, err
}

func CheckUserPassword(user *User) (bool, error) {
	response, _, err := modifyUser("check-user-password", user, nil)
	return response.Status == "ok", err
}
