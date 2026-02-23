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

//go:build !disable_nats

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/nats"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewNATSCommand creates a new nats sub-command.
func NewNATSCommand() *cobra.Command {
	natsCommand := &cobra.Command{
		Use:   "nats ADDRESS... [flags] [-- command [args...]]",
		Short: "Check NATS connection",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the nats command")
			}

			return nil
		},
		Example: `
  # Checking NATS connection without authentication
  wait4x nats nats://localhost:4222

  # Checking NATS connection with username/password authentication
  wait4x nats nats://localhost:4222 --username nats-user --password nats-password

  # Checking NATS connection using a credentials file (username:password)
  wait4x nats nats://localhost:4222 --credential-file ./nats.credentials
`,
		RunE: runNATS,
	}

	natsCommand.Flags().String("username", "", "NATS username")
	natsCommand.Flags().String("password", "", "NATS password")
	natsCommand.Flags().String("credential-file", "", "Path to username/password credentials file")
	natsCommand.Flags().Bool("insecure-skip-tls-verify", false, "Disable TLS certificate verification")
	natsCommand.Flags().Duration("connection-timeout", nats.DefaultConnectionTimeout, "Timeout for establishing a NATS connection")

	return natsCommand
}

func runNATS(cmd *cobra.Command, args []string) error {
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return fmt.Errorf("failed to parse --username flag: %w", err)
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return fmt.Errorf("failed to parse --password flag: %w", err)
	}

	credentialFile, err := cmd.Flags().GetString("credential-file")
	if err != nil {
		return fmt.Errorf("failed to parse --credential-file flag: %w", err)
	}

	if credentialFile != "" && (username != "" || password != "") {
		return errors.New("--credential-file can't be used together with --username/--password")
	}

	if (username == "") != (password == "") {
		return errors.New("both --username and --password must be provided together")
	}

	insecureSkipTLSVerify, err := cmd.Flags().GetBool("insecure-skip-tls-verify")
	if err != nil {
		return fmt.Errorf("failed to parse --insecure-skip-tls-verify flag: %w", err)
	}

	connectionTimeout, err := cmd.Flags().GetDuration("connection-timeout")
	if err != nil {
		return fmt.Errorf("failed to parse --connection-timeout flag: %w", err)
	}

	logger, err := logr.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get logger from context: %w", err)
	}

	// ArgsLenAtDash returns -1 when -- was not specified
	if i := cmd.ArgsLenAtDash(); i != -1 {
		args = args[:i]
	}

	checkers := make([]checker.Checker, len(args))
	for i, arg := range args {
		checkers[i] = nats.New(
			arg,
			nats.WithUsername(username),
			nats.WithPassword(password),
			nats.WithCredentialFile(credentialFile),
			nats.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
			nats.WithTimeout(connectionTimeout),
		)
	}

	return waiter.WaitParallelContext(
		cmd.Context(),
		checkers,
		waiter.WithTimeout(contextutil.GetTimeout(cmd.Context())),
		waiter.WithInterval(contextutil.GetInterval(cmd.Context())),
		waiter.WithInvertCheck(contextutil.GetInvertCheck(cmd.Context())),
		waiter.WithBackoffPolicy(contextutil.GetBackoffPolicy(cmd.Context())),
		waiter.WithBackoffCoefficient(contextutil.GetBackoffCoefficient(cmd.Context())),
		waiter.WithBackoffExponentialMaxInterval(contextutil.GetBackoffExponentialMaxInterval(cmd.Context())),
		waiter.WithLogger(logger),
	)
}
