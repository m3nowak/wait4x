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

func TestMSSQLCommandInvalidArgument(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewMSSQLCommand())

	_, err := test.ExecuteCommand(rootCmd, "mssql")

	assert.Equal(t, "ADDRESS is required argument for the mssql command", err.Error())
}

func TestMSSQLCommandRequiresCredentials(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewMSSQLCommand())

	_, err := test.ExecuteCommand(rootCmd, "mssql", "127.0.0.1:1433")

	assert.Equal(t, "both --username and --password are required for the mssql command", err.Error())
}

func TestMSSQLCommandHelp(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewMSSQLCommand())

	output, err := test.ExecuteCommand(rootCmd, "mssql", "--help")

	assert.NoError(t, err)
	assert.Contains(t, output, "Check MSSQL connection")
	assert.Contains(t, output, "--username")
	assert.Contains(t, output, "--password")
}
