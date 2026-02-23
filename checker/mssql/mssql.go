// Copyright 2019-2025 The Wait4X Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package mssql provides the MSSQL checker for the Wait4X application.
package mssql

import (
	"context"
	"database/sql"
	"net/url"
	"regexp"

	"wait4x.dev/v3/checker"

	// Needed for the MSSQL driver
	_ "github.com/microsoft/go-mssqldb"
)

var hidePasswordRegexp = regexp.MustCompile(`^(sqlserver://[^/:]+):[^:@]+@`)

// Option configures an MSSQL checker.
type Option func(m *MSSQL)

// MSSQL is an MSSQL checker.
type MSSQL struct {
	address               string
	username              string
	password              string
	insecureSkipTLSVerify bool
}

// New creates a new MSSQL checker.
func New(address string, opts ...Option) checker.Checker {
	m := &MSSQL{address: address}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithUsername configures username authentication.
func WithUsername(username string) Option {
	return func(m *MSSQL) {
		m.username = username
	}
}

// WithPassword configures password authentication.
func WithPassword(password string) Option {
	return func(m *MSSQL) {
		m.password = password
	}
}

// WithInsecureSkipTLSVerify configures whether to disable TLS certificate verification.
func WithInsecureSkipTLSVerify(insecureSkipTLSVerify bool) Option {
	return func(m *MSSQL) {
		m.insecureSkipTLSVerify = insecureSkipTLSVerify
	}
}

// Identity returns the identity of the MSSQL checker.
func (m *MSSQL) Identity() (string, error) {
	return m.address, nil
}

func (m *MSSQL) connectionString() string {
	u := &url.URL{
		Scheme: "sqlserver",
		Host:   m.address,
	}

	if m.username != "" {
		u.User = url.UserPassword(m.username, m.password)
	}

	q := u.Query()
	if m.insecureSkipTLSVerify {
		q.Set("encrypt", "true")
		q.Set("trustservercertificate", "true")
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// Check checks the MSSQL connection.
func (m *MSSQL) Check(ctx context.Context) (err error) {
	dsn := m.connectionString()

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return err
	}

	defer func(db *sql.DB) {
		if dberr := db.Close(); dberr != nil {
			err = dberr
		}
	}(db)

	err = db.PingContext(ctx)
	if err != nil {
		if checker.IsConnectionRefused(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the mssql server", err,
				"dsn", hidePasswordRegexp.ReplaceAllString(dsn, `$1:***@`),
			)
		}

		return err
	}

	return nil
}
