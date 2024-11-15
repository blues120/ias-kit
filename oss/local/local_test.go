package local

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/blues120/ias-kit/oss"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LocalTestSuite struct {
	suite.Suite

	local   oss.Oss
	ossKey  string
	ossData []byte
}

func (s *LocalTestSuite) SetupTest() {
	s.ossKey = "test-local"
	s.ossData = []byte{'1', '2', '3'}
}

func TestLocalTestSuite(t *testing.T) {
	ts := new(LocalTestSuite)

	tempDir := t.TempDir()
	local, err := NewLocal(tempDir, "")
	require.NoError(t, err)
	ts.local = local

	suite.Run(t, ts)
}

func (s *LocalTestSuite) TestLocal_Upload() {
	err := s.local.Upload(context.Background(), s.ossKey, bytes.NewReader(s.ossData))
	require.NoError(s.T(), err)
}

func (s *LocalTestSuite) TestLocal_Download() {
	s.TestLocal_Upload()

	readCloser, err := s.local.Download(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	defer readCloser.Close()

	actual, err := io.ReadAll(readCloser)
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.ossData, actual)
}

func (s *LocalTestSuite) TestLocal_Exists() {
	_ = s.local.Delete(context.Background(), s.ossKey)
	exists, err := s.local.Exists(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	require.Equal(s.T(), false, exists)

	s.TestLocal_Upload()
	exists, err = s.local.Exists(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	require.Equal(s.T(), true, exists)
}

func (s *LocalTestSuite) TestLocal_GenerateUrl() {
	url, err := s.local.GenerateUrl(context.Background(), s.ossKey, time.Minute*10)
	require.NoError(s.T(), err)
	s.T().Logf("generate url: %s", url)
}
