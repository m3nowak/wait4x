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

// Package cmd provides the command-line interface for the Wait4X application.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"wait4x.dev/v3/internal/test"
)

func TestNATSCommandInvalidArgument(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewNATSCommand())

	_, err := test.ExecuteCommand(rootCmd, "nats")

	assert.Equal(t, "ADDRESS is required argument for the nats command", err.Error())
}

func TestNATSCommandInvalidAuthCombination(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewNATSCommand())

	_, err := test.ExecuteCommand(rootCmd, "nats", "nats://127.0.0.1:4222", "--username", "user")

	assert.Equal(t, "both --username and --password must be provided together", err.Error())
}

func TestNATSCommandMutuallyExclusiveAuth(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewNATSCommand())

	_, err := test.ExecuteCommand(rootCmd, "nats", "nats://127.0.0.1:4222", "--username", "user", "--password", "pass", "--credential-file", "/tmp/cred")

	assert.Equal(t, "--credential-file can't be used together with --username/--password", err.Error())
}

func TestNATSCommandHelp(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewNATSCommand())

	output, err := test.ExecuteCommand(rootCmd, "nats", "--help")

	assert.NoError(t, err)
	assert.Contains(t, output, "Check NATS connection")
	assert.Contains(t, output, "--credential-file")
}
