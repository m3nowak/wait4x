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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	tcmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
	"wait4x.dev/v3/checker"
)

// MSSQLSuite is a test suite for MSSQL checker.
type MSSQLSuite struct {
	suite.Suite
	container *tcmssql.MSSQLServerContainer
	password  string
}

// SetupSuite starts an MSSQL container.
func (s *MSSQLSuite) SetupSuite() {
	ctx := context.Background()

	password := "Strong@Passw0rd1"
	container, err := tcmssql.Run(
		ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		tcmssql.WithAcceptEULA(),
		tcmssql.WithPassword(password),
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)
	s.Require().NoError(err)

	s.container = container
	s.password = password
}

// TearDownSuite stops the MSSQL container.
func (s *MSSQLSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoError(err)
}

// TestIdentity tests checker identity.
func (s *MSSQLSuite) TestIdentity() {
	chk := New("127.0.0.1:1433")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Equal("127.0.0.1:1433", identity)
}

// TestInvalidConnection tests an invalid MSSQL connection.
func (s *MSSQLSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	chk := New("127.0.0.1:8075", WithUsername("sa"), WithPassword("Strong@Passw0rd1"))

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidConnection tests a valid MSSQL connection.
func (s *MSSQLSuite) TestValidConnection() {
	ctx := context.Background()

	endpoint, err := s.container.PortEndpoint(ctx, "1433/tcp", "")
	s.Require().NoError(err)

	chk := New(
		endpoint,
		WithUsername("sa"),
		WithPassword(s.password),
		WithInsecureSkipTLSVerify(true),
	)

	s.Assert().NoError(chk.Check(ctx))
}

// TestMSSQL runs the MSSQL test suite.
func TestMSSQL(t *testing.T) {
	suite.Run(t, new(MSSQLSuite))
}
