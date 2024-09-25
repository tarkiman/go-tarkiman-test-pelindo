package imgproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/tarkiman/go/configs"
)

var proxy *imgProxy
var config *configs.Config

func Init() {
	config = configs.Get()
}

const (
	defaultResizeType = "fit"
)

type imgProxy struct {
	keyBin  []byte
	saltBin []byte
}

type ImageCompressRequest struct {
	URL        string
	ResizeType string
	Width      int
	Height     int
	Quality    int
}

func Compress(req ImageCompressRequest) string {
	path := fmt.Sprintf("/rs:%s:%d:%d/q:%d/plain/%s", defaultResizeType, req.Width, req.Height, req.Quality, req.URL)
	return sign(path)
}

func sign(path string) string {
	mac := hmac.New(sha256.New, proxy.keyBin)
	mac.Write(proxy.saltBin)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s%s", signature, path)
}

func Resize(url string, width, height int) string {
	path := fmt.Sprintf("/rs:%s:%d:%d/plain/%s", defaultResizeType, width, height, url)
	return sign(path)
}

func NewImageCompressRequest(url string, width, height, quality int) ImageCompressRequest {
	return ImageCompressRequest{
		URL:        url,
		ResizeType: defaultResizeType,
		Width:      width,
		Height:     height,
		Quality:    quality,
	}
}

func GetURL(path string, cdn bool) string {
	return fmt.Sprintf("%s/%s", config.Upload.Imgproxy.HostURL, path)
}

func ReduceQuality(url string, quality int) string {
	path := fmt.Sprintf("/q:%d/plain/%s", quality, url)
	return sign(path)
}
