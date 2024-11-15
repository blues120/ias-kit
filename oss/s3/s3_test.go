package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/blues120/ias-kit/oss"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3TestSuite struct {
	suite.Suite

	s3      oss.Oss
	ossKey  string
	ossData []byte

	multiPartsKey       string
	partsNum            int64
	multiPartsData      [][]byte
	totalMultiPartsData []byte
}

func genBytes(bytesNum int) []byte {
	var buf bytes.Buffer
	for i := 0; i < bytesNum; i++ {
		buf.WriteString(fmt.Sprintf("%d", rand.Intn(10)))
	}
	return buf.Bytes()
}

func (s *S3TestSuite) SetupTest() {
	s.ossKey = "rjp-test.mov"
	s.ossData = []byte{'1', '2', '3'}

	// 分片上传每个分片最小为 5MB
	s.multiPartsKey = "multiparts-test.zip"
	s.partsNum = 10
	s.multiPartsData = make([][]byte, s.partsNum)
	var buf bytes.Buffer
	for i := 0; i < int(s.partsNum); i++ {
		singlePartBytes := genBytes(5 * 1024 * 1024) // 5MB
		s.multiPartsData[i] = singlePartBytes
		buf.Write(singlePartBytes)
	}
	s.totalMultiPartsData = buf.Bytes()
	s.T().Logf("total file bytes: %d", len(s.totalMultiPartsData))
}

func TestS3TestSuite(t *testing.T) {
	ts := new(S3TestSuite)

	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("ak", "sk", ""),
		Endpoint:         aws.String("http://127.0.0.1:9000"),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
		LogLevel:         aws.LogLevel(aws.LogDebug),
		Region:           aws.String("cn"),
	}
	s3, err := NewS3("ecloud", cfg, "")
	require.NoError(t, err)
	ts.s3 = s3

	suite.Run(t, ts)
}

func (s *S3TestSuite) TestS3_Upload() {
	err := s.s3.Upload(context.Background(), s.ossKey, bytes.NewReader(s.ossData))
	require.NoError(s.T(), err)
}

func (s *S3TestSuite) TestS3_Download() {
	s.TestS3_Upload()

	readCloser, err := s.s3.Download(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	defer readCloser.Close()

	actual, err := io.ReadAll(readCloser)
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.ossData, actual)
}

func (s *S3TestSuite) TestS3_Exists() {
	_ = s.s3.Delete(context.Background(), s.ossKey)
	exists, err := s.s3.Exists(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	require.Equal(s.T(), false, exists)

	s.TestS3_Upload()
	exists, err = s.s3.Exists(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	require.Equal(s.T(), true, exists)
}

func (s *S3TestSuite) TestS3_GenerateUrl() {
	url, err := s.s3.GenerateUrl(context.Background(), s.ossKey, time.Second*5)
	require.NoError(s.T(), err)
	s.T().Logf("generate url: %s", url)
}

func (s *S3TestSuite) TestS3_GeneratePermanentUrl() {
	url, err := s.s3.GeneratePermanentUrl(context.Background(), s.ossKey)
	require.NoError(s.T(), err)
	s.T().Logf("generate url: %s", url)
}

func (s *S3TestSuite) TestS3_UploadMultiPart() {
	uploadId, err := s.s3.CreateMultipartUpload(context.Background(), s.multiPartsKey)
	require.NoError(s.T(), err)
	s.T().Logf("uploadId: %s", uploadId)

	for i := 0; i < int(s.partsNum); i++ {
		reader := bytes.NewReader(s.multiPartsData[i])
		etag, err := s.s3.UploadPart(context.Background(), s.multiPartsKey, uploadId, int64(i+1), reader)
		require.NoError(s.T(), err)
		s.T().Logf("partNum: %d, etag: %s", i+1, etag)
	}

	parts, err := s.s3.ListParts(context.Background(), s.multiPartsKey, uploadId, 4)
	require.NoError(s.T(), err)
	partsBytes, _ := json.Marshal(parts)
	s.T().Logf("uploaded parts: %s, total: %d", string(partsBytes), len(parts))
	require.Equal(s.T(), int(s.partsNum), len(parts))

	etag, err := s.s3.CompleteMultipartUpload(context.Background(), s.multiPartsKey, uploadId, 0)
	require.NoError(s.T(), err)
	s.T().Logf("whole file etag: %s", etag)

	readCloser, err := s.s3.Download(context.Background(), s.multiPartsKey)
	require.NoError(s.T(), err)
	defer readCloser.Close()

	actual, err := io.ReadAll(readCloser)
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.totalMultiPartsData, actual)
}
