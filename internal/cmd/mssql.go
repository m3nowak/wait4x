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

//go:build !disable_mssql

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"wait4x.dev/v3/checker"
	"wait4x.dev/v3/checker/mssql"
	"wait4x.dev/v3/internal/contextutil"
	"wait4x.dev/v3/waiter"
)

// NewMSSQLCommand creates a new mssql sub-command.
func NewMSSQLCommand() *cobra.Command {
	mssqlCommand := &cobra.Command{
		Use:   "mssql ADDRESS... [flags] [-- command [args...]]",
		Short: "Check MSSQL connection",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("ADDRESS is required argument for the mssql command")
			}

			return nil
		},
		Example: `
  # Checking MSSQL connection with username and password
  wait4x mssql localhost:1433 --username sa --password 'Strong@Passw0rd!'

  # Checking MSSQL connection with skipped TLS verification
  wait4x mssql localhost:1433 --username sa --password 'Strong@Passw0rd!' --insecure-skip-tls-verify
`,
		RunE: runMSSQL,
	}

	mssqlCommand.Flags().String("username", "", "MSSQL username")
	mssqlCommand.Flags().String("password", "", "MSSQL password")
	mssqlCommand.Flags().Bool("insecure-skip-tls-verify", false, "Disable TLS certificate verification")

	return mssqlCommand
}

func runMSSQL(cmd *cobra.Command, args []string) error {
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return fmt.Errorf("failed to parse --username flag: %w", err)
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return fmt.Errorf("failed to parse --password flag: %w", err)
	}

	if username == "" || password == "" {
		return errors.New("both --username and --password are required for the mssql command")
	}

	insecureSkipTLSVerify, err := cmd.Flags().GetBool("insecure-skip-tls-verify")
	if err != nil {
		return fmt.Errorf("failed to parse --insecure-skip-tls-verify flag: %w", err)
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
		checkers[i] = mssql.New(
			arg,
			mssql.WithUsername(username),
			mssql.WithPassword(password),
			mssql.WithInsecureSkipTLSVerify(insecureSkipTLSVerify),
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
