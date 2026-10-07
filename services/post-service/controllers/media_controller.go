package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/meloop/post-service/services"
)

// UploadFile es una función que retorna el handler de Gin, inyectando el cliente de MinIO
func UploadFile(minioClient *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extraer el archivo enviado bajo la llave "file"
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No se recibió ningún archivo. Usa form-data con la llave 'file'"})
			return
		}

		// 2. Procesar y subir usando nuestro media_service
		url, err := services.UploadMedia(minioClient, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 3. Retornar la URL pública
		c.JSON(http.StatusOK, gin.H{
			"message": "Archivo subido exitosamente",
			"url":     url,
		})
	}
}