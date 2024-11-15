package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/blues120/ias-kit/oss"
)

const (
	AclPrivate         = "private"
	AclPublicRead      = "public-read"
	AclPublicReadWrite = "public-read-write"
)

type awsS3 struct {
	bucket string
	core   *s3.S3

	endpointAlias string
}

func NewS3(bucket string, cfg *aws.Config, endpointAlias string) (oss.Oss, error) {
	se, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return &awsS3{
		core:          s3.New(se, cfg),
		bucket:        bucket,
		endpointAlias: endpointAlias,
	}, nil
}

func (r *awsS3) Upload(ctx context.Context, key string, reader io.Reader) error {
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	obj := &s3.PutObjectInput{
		Body:   bytes.NewReader(fileBytes),
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	_, err = r.core.PutObjectWithContext(ctx, obj)
	return err
}

func (r *awsS3) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	out, err := r.core.GetObjectWithContext(ctx, obj)
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (r *awsS3) Delete(ctx context.Context, key string) error {
	obj := &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	_, err := r.core.DeleteObjectWithContext(ctx, obj)
	return err
}

func (r *awsS3) Exists(ctx context.Context, key string) (bool, error) {
	obj := &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	_, err := r.core.HeadObject(obj)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case "NotFound":
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

func (r *awsS3) GenerateUrl(ctx context.Context, key string, expire time.Duration) (string, error) {
	obj := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	req, _ := r.core.GetObjectRequest(obj)
	url, err := req.Presign(expire)
	if err != nil {
		return "", err
	}
	return r.genUrl(url), nil
}

func (r *awsS3) GenerateTemporaryUrl(ctx context.Context, key string, expire time.Duration) (string, error) {
	// 过期时间最长7天
	if expire > time.Hour*24*7 {
		return "", fmt.Errorf("the expiration time is up to 7 days")
	}
	obj := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	req, _ := r.core.GetObjectRequest(obj)
	url, err := req.Presign(expire)
	if err != nil {
		return "", err
	}
	return r.genUrl(url), nil
}

func (r *awsS3) GeneratePermanentUrl(ctx context.Context, key string) (string, error) {
	_, err := r.core.PutObjectAclWithContext(ctx, &s3.PutObjectAclInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
		ACL:    aws.String(AclPublicRead),
	})
	if err != nil {
		return "", err
	}
	url := r.core.Endpoint + "/" + r.bucket + "/" + key

	return r.genUrl(url), nil
}

func (r *awsS3) genUrl(resource string) string {
	if r.endpointAlias == "" {
		return resource
	}

	return strings.ReplaceAll(resource, r.core.Endpoint, r.endpointAlias)
}

// 分片上传每个分片最小为 5MB，如果不适用需要用普通上传
func (r *awsS3) CreateMultipartUpload(ctx context.Context, key string) (uploadId string, err error) {
	resp, err := r.core.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	return *resp.UploadId, nil
}

func (r *awsS3) UploadPart(ctx context.Context, key, uploadId string, partNumber int64, reader io.ReadSeeker) (etag string, err error) {
	resp, err := r.core.UploadPart(&s3.UploadPartInput{
		Bucket:     aws.String(r.bucket),
		Key:        aws.String(key),
		PartNumber: &partNumber,
		UploadId:   &uploadId,
		Body:       reader,
	})
	if err != nil {
		return "", err
	}
	return *resp.ETag, nil
}

func (r *awsS3) AbortMultipartUpload(ctx context.Context, key, uploadId string) error {
	_, err := r.core.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(key),
		UploadId: &uploadId,
	})
	return err
}

func transformCompletedParts(parts []*oss.CompletedPart) []*s3.CompletedPart {
	var newParts = make([]*s3.CompletedPart, len(parts))
	for i := range parts {
		newParts[i] = &s3.CompletedPart{
			PartNumber: &parts[i].PartNumber,
			ETag:       &parts[i].ETag,
		}
	}
	return newParts
}

func (r *awsS3) CompleteMultipartUpload(ctx context.Context, key, uploadId string, partsNum int64) (etag string, err error) {
	parts, err := r.ListParts(ctx, key, uploadId, 0)
	if err != nil {
		return
	}

	resp, err := r.core.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(key),
		UploadId: &uploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: transformCompletedParts(parts),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.ETag, nil
}

// 循环获取所有分片
// ListMultipartUploads 上传进行中但还没完成的分片
// ListParts 上传完成的分片
func (r *awsS3) ListParts(ctx context.Context, key, uploadId string, maxParts int64) (parts []*oss.CompletedPart, err error) {
	if maxParts == 0 {
		maxParts = 1000
	}
	var partNumberMarker *int64

	for {
		resp, err1 := r.core.ListParts(&s3.ListPartsInput{
			Bucket:           aws.String(r.bucket),
			Key:              aws.String(key),
			UploadId:         &uploadId,
			MaxParts:         &maxParts,
			PartNumberMarker: partNumberMarker,
		})
		if err1 != nil {
			err = err1
			return
		}

		for i := range resp.Parts {
			parts = append(parts, &oss.CompletedPart{
				PartNumber: *resp.Parts[i].PartNumber,
				ETag:       *resp.Parts[i].ETag,
			})
		}

		if *resp.IsTruncated && resp.NextPartNumberMarker != nil {
			partNumberMarker = resp.NextPartNumberMarker
		} else {
			break
		}
	}

	return
}
