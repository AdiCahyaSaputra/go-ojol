package uploadthing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sqids/sqids-go"
)

const (
	defaultIngestHost = "ingest.uploadthing.com"
	signaturePrefix   = "hmac-sha256="
	defaultAlphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Client interface {
	Upload(ctx context.Context, filename, contentType string, size int64, body io.Reader) (string, error)
}

type client struct {
	apiKey    string
	appID     string
	ingestURL string
	http      *http.Client
}

type tokenPayload struct {
	APIKey     string   `json:"apiKey"`
	AppID      string   `json:"appId"`
	Regions    []string `json:"regions"`
	IngestHost string   `json:"ingestHost"`
}

type uploadResponse struct {
	URL    string `json:"url"`
	UfsURL string `json:"ufsUrl"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}

func NewClientFromEnv() (Client, error) {
	token := os.Getenv("UPLOADTHING_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("UPLOADTHING_TOKEN is not set")
	}
	return NewClient(token)
}

func NewClient(token string) (Client, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("decode UPLOADTHING_TOKEN: %w", err)
	}

	var payload tokenPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("parse UPLOADTHING_TOKEN: %w", err)
	}
	if payload.APIKey == "" || payload.AppID == "" || len(payload.Regions) == 0 {
		return nil, fmt.Errorf("UPLOADTHING_TOKEN missing apiKey, appId, or regions")
	}

	ingestHost := payload.IngestHost
	if ingestHost == "" {
		ingestHost = defaultIngestHost
	}

	return &client{
		apiKey:    payload.APIKey,
		appID:     payload.AppID,
		ingestURL: fmt.Sprintf("https://%s.%s", payload.Regions[0], ingestHost),
		http:      &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (c *client) Upload(ctx context.Context, filename, contentType string, size int64, body io.Reader) (string, error) {
	fileSeed := base64.RawURLEncoding.EncodeToString([]byte(uuid.NewString()))
	fileKey := generateKey(fileSeed, c.appID)

	signedURL, err := c.signedUploadURL(fileKey, filename, contentType, size)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, body); err != nil {
		return "", fmt.Errorf("copy file body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, &buf)
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read upload response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var uploaded uploadResponse
	if err := json.Unmarshal(respBody, &uploaded); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}

	if uploaded.UfsURL != "" {
		return uploaded.UfsURL, nil
	}
	if uploaded.URL != "" {
		return uploaded.URL, nil
	}
	return "", fmt.Errorf("upload response missing url")
}

func (c *client) signedUploadURL(fileKey, filename, contentType string, size int64) (string, error) {
	base := fmt.Sprintf("%s/%s", c.ingestURL, fileKey)
	values := url.Values{}
	values.Set("expires", strconv.FormatInt(time.Now().Add(time.Hour).UnixMilli(), 10))
	values.Set("x-ut-identifier", c.appID)
	values.Set("x-ut-file-name", filename)
	values.Set("x-ut-file-size", strconv.FormatInt(size, 10))
	values.Set("x-ut-file-type", contentType)
	values.Set("x-ut-content-disposition", "inline")
	values.Set("x-ut-acl", "public-read")

	unsigned := base + "?" + values.Encode()
	mac := hmac.New(sha256.New, []byte(c.apiKey))
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		return "", fmt.Errorf("sign upload url: %w", err)
	}
	signature := signaturePrefix + hex.EncodeToString(mac.Sum(nil))
	return unsigned + "&" + url.Values{"signature": {signature}}.Encode(), nil
}

// File-key generation follows UploadThing's documented Go reference:
// https://docs.uploadthing.com/uploading-files#generating-presigned-urls
func djb2(s string) int32 {
	h := int64(5381)
	for i := len(s) - 1; i >= 0; i-- {
		h = (h * 33) ^ int64(s[i])
		h &= 0xFFFFFFFF
	}
	h = (h & 0xBFFFFFFF) | ((h >> 1) & 0x40000000)
	if h >= 0x80000000 {
		h -= 0x100000000
	}
	return int32(h)
}

func shuffle(input, seed string) string {
	chars := []rune(input)
	seedNum := djb2(seed)
	for i := range chars {
		j := (int(seedNum)%(i+1) + i) % len(chars)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func generateKey(fileSeed, appID string) string {
	alphabet := shuffle(defaultAlphabet, appID)
	s, err := sqids.New(sqids.Options{
		MinLength: 12,
		Alphabet:  alphabet,
	})
	if err != nil {
		return fileSeed
	}

	encodedAppID, err := s.Encode([]uint64{uint64(math.Abs(float64(djb2(appID))))})
	if err != nil {
		return fileSeed
	}
	return encodedAppID + fileSeed
}
