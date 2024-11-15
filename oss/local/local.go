package local

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/blues120/ias-kit/oss"
	"github.com/google/uuid"
)

type local struct {
	storePath string
	path      string
}

func NewLocal(storePath, path string) (oss.Oss, error) {
	if err := os.MkdirAll(storePath, os.ModePerm); err != nil {
		return nil, err
	}
	return &local{storePath: storePath, path: path}, nil
}

func (r *local) Upload(ctx context.Context, key string, reader io.Reader) error {
	p := r.getSavePath(key)

	out, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return err
	}
	return nil
}

func (r *local) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	p := r.getSavePath(key)
	return os.Open(p)
}

func (r *local) Delete(ctx context.Context, key string) error {
	p := r.getSavePath(key)
	return os.Remove(p)
}

// Exists 判断本地文件是否存在
func (r *local) Exists(ctx context.Context, key string) (bool, error) {
	// 获取绝对路径
	p := r.getSavePath(key)

	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		//如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}

	return false, err
}

func (r *local) GenerateUrl(ctx context.Context, key string, expire time.Duration) (string, error) {
	//name := r.getMD5Name(key)
	if r.path == "" {
		return path.Join(r.storePath, key), nil
	}

	return path.Join(r.path, key), nil
}

func (r *local) GenerateTemporaryUrl(ctx context.Context, key string, expire time.Duration) (string, error) {
	return r.getSavePath(key), nil
}

func (r *local) GeneratePermanentUrl(ctx context.Context, key string) (string, error) {
	return r.getSavePath(key), nil
}

// // getMD5Name 获取md5格式的文件名
// func (r *local) getMD5Name(key string) string {
// 	// 读取文件后缀
// 	ext := path.Ext(key)

// 	// 读取文件名并加密
// 	name := strings.TrimSuffix(key, ext)
// 	hash := md5.Sum([]byte(name))
// 	name = hex.EncodeToString(hash[:])

// 	// 拼接新文件名
// 	return name + ext
// }

// getSavePath 获取存储路径
func (r *local) getSavePath(key string) string {
	//name := r.getMD5Name(key)
	return path.Join(r.storePath, key)
}

// 分片上传每个分片最小为 5MB，如果不适用需要用普通上传
func (r *local) CreateMultipartUpload(ctx context.Context, key string) (uploadId string, err error) {
	return uuid.New().String(), nil
}

func (r *local) UploadPart(ctx context.Context, key, uploadId string, partNumber int64, reader io.ReadSeeker) (etag string, err error) {

	// 创建本地文件
	targetFileName := getUploadTmpPath(uploadId, partNumber)
	err = os.MkdirAll(filepath.Dir(targetFileName), os.ModePerm)
	if err != nil {
		return "", err
	}
	localFile, err := os.Create(targetFileName)
	if err != nil {
		return "", err
	}
	defer localFile.Close()

	// 将输入参数的指针重新定位到起始位置
	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// 读取数据并写入文件
	_, err = io.Copy(localFile, reader)
	if err != nil {
		err = os.RemoveAll(getUploadTmpPath(uploadId, partNumber))
		return "", err
	}

	return "", nil
}

func (r *local) AbortMultipartUpload(ctx context.Context, key, uploadId string) error {
	err := os.RemoveAll("/tmp/" + uploadId)
	return err
}

func (r *local) CompleteMultipartUpload(ctx context.Context, key, uploadId string, partsNum int64) (etag string, err error) {
	// 生成最终文件
	targetPath := r.getSavePath(key)
	finalFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer finalFile.Close()

	// 获取所有分片
	parts, err := r.ListParts(ctx, key, uploadId, partsNum)
	if err != nil {
		return "", err
	}
	if len(parts) != int(partsNum) {
		return "", errors.New("有分片缺失")
	}

	// 将所有分片写入最终文件
	for _, part := range parts {
		partPath := getUploadTmpPath(uploadId, part.PartNumber)
		tempFile, err := os.Open(partPath)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(finalFile, tempFile)
		tempFile.Close()
		if err != nil {
			return "", err
		}
	}

	// 删除临时文件
	err = os.RemoveAll("/tmp/" + uploadId)

	return "", err
}

// 循环获取所有分片
func (r *local) ListParts(ctx context.Context, key, uploadId string, maxParts int64) (parts []*oss.CompletedPart, err error) {

	ret := make([]*oss.CompletedPart, 0)

	if maxParts == 0 {
		maxParts = 1000
	}

	for i := 1; i <= int(maxParts); i++ {
		filepath := getUploadTmpPath(uploadId, int64(i))
		if _, err := os.Stat(filepath); err != nil {
			continue
		}
		ret = append(ret, &oss.CompletedPart{
			PartNumber: int64(i),
			ETag:       "",
		})
	}

	return ret, nil
}

func getUploadTmpPath(uploadId string, partNumber int64) string {
	return "/tmp/" + uploadId + "/part_" + strconv.Itoa(int(partNumber)) + ".tmp"
}
