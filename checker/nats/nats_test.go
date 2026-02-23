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
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	tcnats "github.com/testcontainers/testcontainers-go/modules/nats"
	"wait4x.dev/v3/checker"
)

// NATSSuite is a test suite for NATS checker.
type NATSSuite struct {
	suite.Suite
	noAuthContainer *tcnats.NATSContainer
	authContainer   *tcnats.NATSContainer
	credentialFile  string
}

// SetupSuite starts NATS containers.
func (s *NATSSuite) SetupSuite() {
	ctx := context.Background()

	noAuthContainer, err := tcnats.Run(
		ctx,
		"nats:2.11.7",
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)
	s.Require().NoError(err)
	s.noAuthContainer = noAuthContainer

	authContainer, err := tcnats.Run(
		ctx,
		"nats:2.11.7",
		tcnats.WithUsername("nats-user"),
		tcnats.WithPassword("nats-password"),
		testcontainers.WithLogger(log.TestLogger(s.T())),
	)
	s.Require().NoError(err)
	s.authContainer = authContainer

	file, err := os.CreateTemp("", "nats-credentials-*.txt")
	s.Require().NoError(err)
	_, err = file.WriteString("nats-user:nats-password")
	s.Require().NoError(err)
	s.Require().NoError(file.Close())
	s.credentialFile = file.Name()
}

// TearDownSuite stops NATS containers.
func (s *NATSSuite) TearDownSuite() {
	if s.noAuthContainer != nil {
		err := s.noAuthContainer.Terminate(context.Background())
		s.Require().NoError(err)
	}

	if s.authContainer != nil {
		err := s.authContainer.Terminate(context.Background())
		s.Require().NoError(err)
	}

	if s.credentialFile != "" {
		s.Require().NoError(os.Remove(s.credentialFile))
	}
}

// TestIdentity tests checker identity.
func (s *NATSSuite) TestIdentity() {
	chk := New("nats://127.0.0.1:4222")
	identity, err := chk.Identity()

	s.Require().NoError(err)
	s.Equal("127.0.0.1:4222", identity)
}

// TestInvalidIdentity tests invalid identity.
func (s *NATSSuite) TestInvalidIdentity() {
	chk := New("127.0.0.1:4222")
	_, err := chk.Identity()

	s.Require().ErrorContains(err, "can't retrieve the checker identity")
}

// TestInvalidConnection tests invalid connection.
func (s *NATSSuite) TestInvalidConnection() {
	var expectedError *checker.ExpectedError
	chk := New("nats://127.0.0.1:8075")

	s.Assert().ErrorAs(chk.Check(context.Background()), &expectedError)
}

// TestValidConnectionNoLogin tests no-login mode.
func (s *NATSSuite) TestValidConnectionNoLogin() {
	ctx := context.Background()
	endpoint, err := s.noAuthContainer.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint)
	s.Assert().NoError(chk.Check(ctx))
}

// TestValidConnectionUsernamePassword tests username/password login.
func (s *NATSSuite) TestValidConnectionUsernamePassword() {
	ctx := context.Background()
	endpoint, err := s.authContainer.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint, WithUsername("nats-user"), WithPassword("nats-password"))
	s.Assert().NoError(chk.Check(ctx))
}

// TestValidConnectionCredentialFile tests credentials-file login.
func (s *NATSSuite) TestValidConnectionCredentialFile() {
	ctx := context.Background()
	endpoint, err := s.authContainer.ConnectionString(ctx)
	s.Require().NoError(err)

	chk := New(endpoint, WithCredentialFile(s.credentialFile))
	s.Assert().NoError(chk.Check(ctx))
}

// TestNATS runs the NATS test suite.
func TestNATS(t *testing.T) {
	suite.Run(t, new(NATSSuite))
}
