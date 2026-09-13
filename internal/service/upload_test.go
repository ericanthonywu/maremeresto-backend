package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
)

func createMultipartFileHeader(filename string, content []byte) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="`+filename+`"`)
	h.Set("Content-Type", "image/png")

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, err
	}
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		return nil, err
	}
	defer form.RemoveAll()

	files := form.File["image"]
	if len(files) == 0 {
		return nil, os.ErrNotExist
	}
	return files[0], nil
}

func TestUploadAndCompressImage_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "upload_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	svc := &Service{
		cfg: &config.Config{
			UploadDir: tempDir,
		},
	}

	// Create valid PNG image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		t.Fatalf("failed to encode sample image: %v", err)
	}

	fileHeader, err := createMultipartFileHeader("test.png", imgBuf.Bytes())
	if err != nil {
		t.Fatalf("failed to create multipart header: %v", err)
	}

	url, err := svc.UploadAndCompressImage(context.Background(), fileHeader)
	if err != nil {
		t.Fatalf("UploadAndCompressImage failed: %v", err)
	}

	if url == "" {
		t.Errorf("expected non-empty url, got empty")
	}

	// Verify the file was saved on disk in tempDir
	filename := filepath.Base(url)
	savedFilePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(savedFilePath); os.IsNotExist(err) {
		t.Errorf("expected uploaded file to exist at %s", savedFilePath)
	}
}

func TestUploadAndCompressImage_InvalidImageRejection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "upload_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	svc := &Service{
		cfg: &config.Config{
			UploadDir: tempDir,
		},
	}

	// Invalid non-image content (plain text or script)
	invalidContent := []byte("this is not an image file")

	fileHeader, err := createMultipartFileHeader("hack.txt", invalidContent)
	if err != nil {
		t.Fatalf("failed to create multipart header: %v", err)
	}

	url, err := svc.UploadAndCompressImage(context.Background(), fileHeader)
	if err == nil {
		t.Fatalf("expected error for invalid image, got url: %s", url)
	}

	// Verify no files were created in tempDir
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read tempDir: %v", err)
	}
	if len(entries) > 0 {
		t.Errorf("expected no files to be saved in upload directory, found %d files", len(entries))
	}
}
