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
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type PostController struct {
	repo        *repositories.PostgresPostRepository
	minioClient *minio.Client
}

func NewPostController(repo *repositories.PostgresPostRepository, minioClient *minio.Client) *PostController {
	return &PostController{
		repo:        repo,
		minioClient: minioClient,
	}
}

func (pc *PostController) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	// Usamos ShouldBind para soportar tanto JSON como form-data (necesario para archivos)
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos de formulario inválidos"})
		return
	}

	post := models.Post{
		AuthorID: req.AuthorID,
		Content:  req.Content,
		MusicID:  req.MusicID,
	}

	// Integración con MinIO si la petición incluye un archivo multimedia (RF-16)
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		objectName := uuid.New().String() + ext
		bucketName := "meloop-media"

		_, err = pc.minioClient.PutObject(context.Background(), bucketName, objectName, file, header.Size, minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error subiendo archivo a MinIO"})
			return
		}
		post.MediaURL = fmt.Sprintf("http://127.0.0.1:9000/%s/%s", bucketName, objectName)
	}

	// Inserción en la base de datos de Supabase
	createdPost, err := pc.repo.Create(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error guardando publicación"})
		return
	}

	c.JSON(http.StatusCreated, createdPost)
}