package oss

import (
	"context"
	"io"
	"time"
)

type CompletedPart struct {
	PartNumber int64
	ETag       string
}

type Oss interface {
	// Upload 上传文件
	Upload(ctx context.Context, key string, reader io.Reader) error

	// Download 下载文件
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, key string) error

	// Exists 判断文件是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// GenerateUrl 生成文件下载链接
	// expire 链接过期时间
	// Deprecated: use GenerateTemporaryUrl or GeneratePermanentUrl instead
	GenerateUrl(ctx context.Context, key string, expire time.Duration) (string, error)

	// GenerateTemporaryUrl 生成临时访问链接
	// expire 链接过期时间，最长有效期7天
	GenerateTemporaryUrl(ctx context.Context, key string, expire time.Duration) (string, error)

	// GeneratePermanentUrl 生成永久访问链接
	GeneratePermanentUrl(ctx context.Context, key string) (string, error)

	// 初始化分片上传，用于大文件分片上传
	CreateMultipartUpload(ctx context.Context, key string) (uploadId string, err error)

	// 上传分片
	UploadPart(ctx context.Context, key, uploadId string, partNumber int64, reader io.ReadSeeker) (etag string, err error)

	// 终止分片上传
	AbortMultipartUpload(ctx context.Context, key, uploadId string) error

	// 完成分片上传
	CompleteMultipartUpload(ctx context.Context, key, uploadId string, partsNum int64) (etag string, err error)

	// 列举已上传分片
	ListParts(ctx context.Context, key, uploadId string, maxParts int64) (parts []*CompletedPart, err error)
}
