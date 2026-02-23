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

// Package nats provides NATS checker.
package nats

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"wait4x.dev/v3/checker"
)

const (
	// DefaultConnectionTimeout is the default connection timeout duration.
	DefaultConnectionTimeout = 3 * time.Second
)

// Option configures a NATS checker.
type Option func(n *NATS)

// NATS is a NATS checker.
type NATS struct {
	address               string
	username              string
	password              string
	credentialFile        string
	insecureSkipTLSVerify bool
	timeout               time.Duration
}

// New creates a new NATS checker.
func New(address string, opts ...Option) checker.Checker {
	n := &NATS{
		address: address,
		timeout: DefaultConnectionTimeout,
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// WithUsername configures username authentication.
func WithUsername(username string) Option {
	return func(n *NATS) {
		n.username = username
	}
}

// WithPassword configures password authentication.
func WithPassword(password string) Option {
	return func(n *NATS) {
		n.password = password
	}
}

// WithCredentialFile configures a username/password credentials file.
func WithCredentialFile(path string) Option {
	return func(n *NATS) {
		n.credentialFile = path
	}
}

// WithInsecureSkipTLSVerify configures whether to disable TLS certificate verification.
func WithInsecureSkipTLSVerify(insecureSkipTLSVerify bool) Option {
	return func(n *NATS) {
		n.insecureSkipTLSVerify = insecureSkipTLSVerify
	}
}

// WithTimeout configures connection timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(n *NATS) {
		n.timeout = timeout
	}
}

// Identity returns checker identity.
func (n *NATS) Identity() (string, error) {
	u, err := url.Parse(n.address)
	if err != nil {
		return "", fmt.Errorf("can't retrieve the checker identity: %w", err)
	}

	if u.Host == "" {
		return "", fmt.Errorf("can't retrieve the checker identity: invalid address")
	}

	return u.Host, nil
}

func (n *NATS) authOption() (nats.Option, error) {
	if n.credentialFile != "" {
		username, password, err := readCredentialFile(n.credentialFile)
		if err != nil {
			return nil, err
		}

		return nats.UserInfo(username, password), nil
	}

	if n.username != "" || n.password != "" {
		return nats.UserInfo(n.username, n.password), nil
	}

	return nil, nil
}

func readCredentialFile(path string) (string, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("can't read credentials file: %w", err)
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return "", "", fmt.Errorf("can't parse credentials file: empty file")
	}

	if strings.Contains(trimmed, ":") {
		splitted := strings.SplitN(trimmed, ":", 2)
		if splitted[0] == "" {
			return "", "", fmt.Errorf("can't parse credentials file: username is required")
		}

		return splitted[0], splitted[1], nil
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("can't parse credentials file: expected username:password or two-line format")
	}

	username := strings.TrimSpace(lines[0])
	password := strings.TrimSpace(lines[1])
	if username == "" {
		return "", "", fmt.Errorf("can't parse credentials file: username is required")
	}

	return username, password, nil
}

// Check checks NATS connection.
func (n *NATS) Check(ctx context.Context) error {
	opts := []nats.Option{nats.Timeout(n.timeout)}

	authOpt, err := n.authOption()
	if err != nil {
		return err
	}

	if authOpt != nil {
		opts = append(opts, authOpt)
	}

	if n.insecureSkipTLSVerify {
		opts = append(opts, nats.Secure(&tls.Config{InsecureSkipVerify: true})) //nolint:gosec
	}

	nc, err := nats.Connect(n.address, opts...)
	if err != nil {
		if isExpectedConnectionError(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the nats server",
				err,
				"address", n.address,
			)
		}

		return err
	}

	defer nc.Close()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, n.timeout)
		defer cancel()
	}

	if err = nc.FlushWithContext(ctx); err != nil {
		if isExpectedConnectionError(err) {
			return checker.NewExpectedError(
				"failed to establish a connection to the nats server",
				err,
				"address", n.address,
			)
		}

		return err
	}

	if err = nc.LastError(); err != nil {
		return err
	}

	return nil
}

func isExpectedConnectionError(err error) bool {
	return checker.IsConnectionRefused(err) || strings.Contains(err.Error(), "no servers available for connection")
}
