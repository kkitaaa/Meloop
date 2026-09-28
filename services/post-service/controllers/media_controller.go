package controllers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func UploadFile(minioClient *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo procesar el archivo"})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		objectName := uuid.New().String() + ext
		bucketName := "meloop-media"

		_, err = minioClient.PutObject(context.Background(), bucketName, objectName, file, header.Size, minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error subiendo archivo a MinIO"})
			return
		}

		url := fmt.Sprintf("http://127.0.0.1:9000/%s/%s", bucketName, objectName)
		c.JSON(http.StatusOK, gin.H{"message": "Archivo subido exitosamente", "url": url})
	}
}