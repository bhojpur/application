package validations_test

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
	"errors"
	"fmt"
	"regexp"
	"testing"

	validations "github.com/bhojpur/application/pkg/validations"
	"github.com/bhojpur/application/test/utils"
	errsvr "github.com/bhojpur/errors/pkg/validation"
	orm "github.com/bhojpur/orm/pkg/engine"
	_ "github.com/mattn/go-sqlite3"
)

var db *orm.DB

type User struct {
	orm.Model
	Name           string `valid:"required"`
	Password       string `valid:"length(6|20)"`
	SecurePassword string `valid:"numeric"`
	Email          string `valid:"email,uniqEmail~Email already be token"`
	CompanyID      int
	Company        Company
	CreditCard     CreditCard
	Addresses      []Address
	Languages      []Language `orm:"many2many:user_languages"`
}

func (user *User) Validate(db *orm.DB) {
	errsvr.CustomTypeTagMap.Set("uniqEmail", errsvr.CustomTypeValidator(func(email interface{}, context interface{}) bool {
		var count int
		if db.Model(&User{}).Where("email = ?", email).Count(&count); count == 0 || email == "" {
			return true
		}
		return false
	}))
	if user.Name == "invalid" {
		db.AddError(validations.NewError(user, "Name", "invalid user name"))
	}
}

type Company struct {
	orm.Model
	Name string
}

func (company *Company) Validate(db *orm.DB) {
	if company.Name == "invalid" {
		db.AddError(errors.New("invalid company name"))
	}
}

type CreditCard struct {
	orm.Model
	UserID int
	Number string
}

func (card *CreditCard) Validate(db *orm.DB) {
	if !regexp.MustCompile("^(\\d){13,16}$").MatchString(card.Number) {
		db.AddError(validations.NewError(card, "Number", "invalid card number"))
	}
}

type Address struct {
	orm.Model
	UserID  int
	Address string
}

func (address *Address) Validate(db *orm.DB) {
	if address.Address == "invalid" {
		db.AddError(validations.NewError(address, "Address", "invalid address"))
	}
}

type Language struct {
	orm.Model
	Code string
}

func (language *Language) Validate(db *orm.DB) error {
	if language.Code == "invalid" {
		return validations.NewError(language, "Code", "invalid language")
	}
	return nil
}

func init() {
	db = utils.TestDB()
	validations.RegisterCallbacks(db)
	tables := []interface{}{&User{}, &Company{}, &CreditCard{}, &Address{}, &Language{}}
	for _, table := range tables {
		if err := db.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
		db.AutoMigrate(table)
	}
}

func TestGoValidation(t *testing.T) {
	user := User{Name: "", Password: "123123", Email: "a@gmail.com"}

	result := db.Save(&user)
	if result.Error == nil {
		t.Errorf("Should get error when save empty user")
	}

	if result.Error.Error() != "Name can't be blank" {
		t.Errorf("Error message should be equal `Name can't be blank`")
	}

	user = User{Name: "", Password: "123", SecurePassword: "AB123", Email: "aagmail.com"}
	result = db.Save(&user)
	messages := []string{"Name can't be blank",
		"Password is the wrong length (should be 6~20 characters)",
		"SecurePassword is not a number",
		"Email is not a valid email address"}
	for i, err := range result.GetErrors() {
		if messages[i] != err.Error() {
			t.Errorf(fmt.Sprintf("Error message should be equal `%v`, but it is `%v`", messages[i], err.Error()))
		}
	}

	user = User{Name: "A", Password: "123123", Email: "a@gmail.com"}
	result = db.Save(&user)
	user = User{Name: "B", Password: "123123", Email: "a@gmail.com"}
	if result := db.Save(&user); result.Error.Error() != "Email already be token" {
		t.Errorf("Should get email alredy be token error")
	}
}

func TestSaveInvalidUser(t *testing.T) {
	user := User{Name: "invalid"}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid user")
	}
}

func TestSaveInvalidCompany(t *testing.T) {
	user := User{
		Name:    "valid",
		Company: Company{Name: "invalid"},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid company")
	}
}

func TestSaveInvalidCreditCard(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "invalid"},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid credit card")
	}
}

func TestSaveInvalidAddresses(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "4111111111111111"},
		Addresses:  []Address{{Address: "invalid"}},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid addresses")
	}
}

func TestSaveInvalidLanguage(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "4111111111111111"},
		Addresses:  []Address{{Address: "valid"}},
		Languages:  []Language{{Code: "invalid"}},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid language")
	}
}

func TestSaveAllValidData(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "4111111111111111"},
		Addresses:  []Address{{Address: "valid1"}, {Address: "valid2"}},
		Languages:  []Language{{Code: "valid1"}, {Code: "valid2"}},
	}

	if result := db.Save(&user); result.Error != nil {
		t.Errorf("Should get no error when save valid data, but got: %v", result.Error)
	}
}
