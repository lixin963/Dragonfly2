/*
 *     Copyright 2022 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package objectstorage

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-http-utils/headers"
)

type oss struct {
	// OSS client.
	client *aliyunoss.Client
}

// New oss instance.
func newOSS(region, endpoint, accessKey, secretKey string) (ObjectStorage, error) {
	client, err := aliyunoss.New(endpoint, accessKey, secretKey, aliyunoss.Region(region))
	if err != nil {
		return nil, fmt.Errorf("new oss client failed: %s", err)
	}

	return &oss{
		client: client,
	}, nil
}

// GetBucketMetadata returns metadata of bucket.
func (o *oss) GetBucketMetadata(ctx context.Context, bucketName string) (*BucketMetadata, error) {
	resp, err := o.client.GetBucketInfo(bucketName)
	if err != nil {
		return nil, err
	}

	return &BucketMetadata{
		Name:     resp.BucketInfo.Name,
		CreateAt: resp.BucketInfo.CreationDate,
	}, nil
}

// GetBucketMetadata returns metadata of bucket.
func (o *oss) CreateBucket(ctx context.Context, bucketName string) error {
	return o.client.CreateBucket(bucketName)
}

// DeleteBucket deletes bucket of object storage.
func (o *oss) DeleteBucket(ctx context.Context, bucketName string) error {
	return o.client.DeleteBucket(bucketName)
}

// DeleteBucket deletes bucket of object storage.
func (o *oss) ListBucketMetadatas(ctx context.Context) ([]*BucketMetadata, error) {
	resp, err := o.client.ListBuckets()
	if err != nil {
		return nil, err
	}

	var metadatas []*BucketMetadata
	for _, bucket := range resp.Buckets {
		metadatas = append(metadatas, &BucketMetadata{
			Name:     bucket.Name,
			CreateAt: bucket.CreationDate,
		})
	}

	return metadatas, nil
}

// GetObjectMetadata returns metadata of object.
func (o *oss) GetObjectMetadata(ctx context.Context, bucketName, objectKey string) (*ObjectMetadata, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	header, err := bucket.GetObjectMeta(objectKey)
	if err != nil {
		return nil, err
	}

	contentLength, err := strconv.ParseInt(header.Get(headers.ContentLength), 10, 64)
	if err != nil {
		return nil, err
	}

	return &ObjectMetadata{
		Key:                objectKey,
		ContentDisposition: header.Get(headers.ContentDisposition),
		ContentEncoding:    header.Get(headers.ContentEncoding),
		ContentLanguage:    header.Get(headers.ContentLanguage),
		ContentLength:      contentLength,
		ContentType:        header.Get(headers.ContentType),
		Etag:               header.Get(headers.ETag),
		Digest:             header.Get(aliyunoss.HTTPHeaderOssMetaPrefix + MetaDigest),
	}, nil
}

// GetOject returns data of object.
func (o *oss) GetOject(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket.GetObject(bucketName)
}

// CreateObject creates data of object.
func (o *oss) CreateObject(ctx context.Context, bucketName, objectKey, digest string, reader io.Reader) error {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return err
	}

	meta := aliyunoss.Meta(MetaDigest, digest)
	return bucket.PutObject(objectKey, reader, meta)
}

// DeleteObject deletes data of object.
func (o *oss) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return err
	}

	return bucket.DeleteObject(objectKey)
}

// ListObjectMetadatas returns metadata of objects.
func (o *oss) ListObjectMetadatas(ctx context.Context, bucketName, prefix, marker string, limit int64) ([]*ObjectMetadata, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	resp, err := bucket.ListObjects(aliyunoss.Prefix(prefix), aliyunoss.Marker(marker), aliyunoss.MaxKeys(int(limit)))
	if err != nil {
		return nil, err
	}

	var metadatas []*ObjectMetadata
	for _, object := range resp.Objects {
		metadatas = append(metadatas, &ObjectMetadata{
			Key:  object.Key,
			Etag: object.ETag,
		})
	}

	return metadatas, nil
}

// GetSignURL returns sign url of object.
func (o *oss) GetSignURL(ctx context.Context, bucketName, objectKey string, method Method, expire time.Duration) (string, error) {
	var ossHTTPMethod aliyunoss.HTTPMethod
	switch method {
	case MethodGet:
		ossHTTPMethod = aliyunoss.HTTPGet
	case MethodPut:
		ossHTTPMethod = aliyunoss.HTTPPut
	case MethodHead:
		ossHTTPMethod = aliyunoss.HTTPHead
	case MethodPost:
		ossHTTPMethod = aliyunoss.HTTPPost
	case MethodDelete:
		ossHTTPMethod = aliyunoss.HTTPDelete
	case MethodList:
		ossHTTPMethod = aliyunoss.HTTPGet
	default:
		return "", fmt.Errorf("not support method %s", method)
	}

	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return "", err
	}

	return bucket.SignURL(objectKey, ossHTTPMethod, int64(expire.Seconds()))
}
