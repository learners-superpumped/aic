package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// BucketDTO is a storage bucket.
type BucketDTO struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// ObjectDTO is a stored object's metadata.
type ObjectDTO struct {
	Key         string `json:"key"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
	ETag        string `json:"etag"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func storageBasePath(teamID, projectID string) string {
	return fmt.Sprintf("/v1/teams/%s/projects/%s/storage",
		url.PathEscape(teamID), url.PathEscape(projectID))
}

// escapeKey percent-escapes each path segment of an object key while preserving
// the slashes that separate segments.
func escapeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

func (c *Client) CreateBucket(ctx context.Context, team, project, name string) (BucketDTO, error) {
	var out BucketDTO
	return out, c.do(ctx, "POST", storageBasePath(team, project)+"/buckets", map[string]string{"name": name}, &out)
}

func (c *Client) ListBuckets(ctx context.Context, team, project string, limit int, cursor string) (Page[BucketDTO], error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := storageBasePath(team, project) + "/buckets"
	if e := q.Encode(); e != "" {
		path += "?" + e
	}
	var out Page[BucketDTO]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) DeleteBucket(ctx context.Context, team, project, bucket string) error {
	return c.do(ctx, "DELETE", storageBasePath(team, project)+"/buckets/"+url.PathEscape(bucket), nil, nil)
}

func (c *Client) PutObject(ctx context.Context, team, project, bucket, key string, body []byte, contentType string) (ObjectDTO, error) {
	var out ObjectDTO
	path := storageBasePath(team, project) + "/buckets/" + url.PathEscape(bucket) + "/objects/" + escapeKey(key)
	return out, c.doBytes(ctx, "PUT", path, body, contentType, &out)
}

// GetObject downloads an object's raw bytes and its Content-Type.
func (c *Client) GetObject(ctx context.Context, team, project, bucket, key string) ([]byte, string, error) {
	path := storageBasePath(team, project) + "/buckets/" + url.PathEscape(bucket) + "/objects/" + escapeKey(key)
	data, hdr, err := c.doRaw(ctx, "GET", path)
	if err != nil {
		return nil, "", err
	}
	return data, hdr.Get("Content-Type"), nil
}

func (c *Client) ListObjects(ctx context.Context, team, project, bucket, prefix string, limit int, cursor string) (Page[ObjectDTO], error) {
	q := url.Values{}
	if prefix != "" {
		q.Set("prefix", prefix)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := storageBasePath(team, project) + "/buckets/" + url.PathEscape(bucket) + "/objects"
	if e := q.Encode(); e != "" {
		path += "?" + e
	}
	var out Page[ObjectDTO]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) DeleteObject(ctx context.Context, team, project, bucket, key string) error {
	path := storageBasePath(team, project) + "/buckets/" + url.PathEscape(bucket) + "/objects/" + escapeKey(key)
	return c.do(ctx, "DELETE", path, nil, nil)
}

// SignURL returns a time-limited URL for the given object; op is "download" or
// "upload". expires is an optional Go duration string (e.g. "15m").
func (c *Client) SignURL(ctx context.Context, team, project, bucket, key, op, expires string) (string, error) {
	body := struct {
		Bucket  string `json:"bucket"`
		Key     string `json:"key"`
		Expires string `json:"expires,omitempty"`
	}{Bucket: bucket, Key: key, Expires: expires}
	path := storageBasePath(team, project) + "/signed-download"
	if op == "upload" {
		path = storageBasePath(team, project) + "/signed-upload"
	}
	var out struct {
		URL string `json:"url"`
	}
	return out.URL, c.do(ctx, "POST", path, body, &out)
}
