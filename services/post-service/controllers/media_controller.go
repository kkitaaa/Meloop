package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/services"
	"github.com/minio/minio-go/v7"
)

// UploadFile sube un archivo multimedia vinculado al usuario autenticado
func UploadFile(minioClient *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetString("userID"))
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
			return
		}

		// 1. Extraer el archivo enviado bajo la llave "file"
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No se recibió ningún archivo. Usa form-data con la llave 'file'"})
			return
		}

		// 2. Procesar y subir bajo el espacio del usuario autenticado
		objectKey, url, err := services.UploadMediaForOwner(c.Request.Context(), minioClient, userID, file)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "formato no soportado") || strings.Contains(errMsg, "supera el tamaño") {
				c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}

		// 3. Retornar la URL pública y la clave registrada por el servidor
		c.JSON(http.StatusOK, gin.H{
			"message":    "Archivo subido exitosamente",
			"url":        url,
			"object_key": objectKey,
		})
	}
}
