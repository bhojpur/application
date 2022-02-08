package utils

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
	"os"

	orm "github.com/bhojpur/orm/pkg/engine"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// TestDB initialize a db for testing
func TestDB() *orm.DB {
	var db *orm.DB
	var err error
	var dbuser, dbpwd, dbname, dbhost = "bhojpur", "bhojpur", "bhojpur_test", "localhost"

	if os.Getenv("DB_USER") != "" {
		dbuser = os.Getenv("DB_USER")
	}

	if os.Getenv("DB_PWD") != "" {
		dbpwd = os.Getenv("DB_PWD")
	}

	if os.Getenv("DB_NAME") != "" {
		dbname = os.Getenv("DB_NAME")
	}

	if os.Getenv("DB_HOST") != "" {
		dbhost = os.Getenv("DB_HOST")
	}

	if os.Getenv("TEST_DB") == "mysql" {
		// CREATE USER 'bhojpur'@'localhost' IDENTIFIED BY 'bhojpur';
		// CREATE DATABASE bhojpur_test;
		// GRANT ALL ON bhojpur_test.* TO 'bhojpur'@'localhost';
		db, err = orm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))
	} else {
		db, err = orm.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbuser, dbpwd, dbhost, dbname))
	}

	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") != "" {
		db.LogMode(true)
	}

	return db
}

var db *orm.DB

func GetTestDB() *orm.DB {
	if db != nil {
		return db
	}

	db = TestDB()

	return db
}

// PrepareDBAndTables prepare given tables cleanly and return a test database instance
func PrepareDBAndTables(tables ...interface{}) *orm.DB {
	db := GetTestDB()

	ResetDBTables(db, tables...)

	return db
}

// ResetDBTables reset given tables.
func ResetDBTables(db *orm.DB, tables ...interface{}) {
	Truncate(db, tables...)
	AutoMigrate(db, tables...)
}

// Truncate receives table arguments and truncate their content in database.
func Truncate(db *orm.DB, givenTables ...interface{}) {
	// We need to iterate throught the list in reverse order of
	// creation, since later tables may have constraints or
	// dependencies on earlier tables.
	len := len(givenTables)
	for i := range givenTables {
		table := givenTables[len-i-1]
		db.DropTableIfExists(table)
	}
}

// AutoMigrate receives table arguments and create or update their
// table structure in database.
func AutoMigrate(db *orm.DB, givenTables ...interface{}) {
	for _, table := range givenTables {
		db.AutoMigrate(table)
		if migratable, ok := table.(Migratable); ok {
			exec(func() error { return migratable.AfterMigrate(db) })
		}
	}
}

// Migratable defines interface for implementing post-migration
// actions such as adding constraints that arent's supported by ORM's
// struct tags. This function must be idempotent, since it will most
// likely be executed multiple times.
type Migratable interface {
	AfterMigrate(db *orm.DB) error
}

func exec(c func() error) {
	if err := c(); err != nil {
		panic(err)
	}
}
