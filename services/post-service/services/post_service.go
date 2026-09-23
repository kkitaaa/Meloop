package services

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/meloop/post-service/models"
)

// spotifyIDRegex valida que el ID sea alfanumérico y de exactamente 22 caracteres
var spotifyIDRegex = regexp.MustCompile(`^[a-zA-Z0-9]{22}$`)

// ValidateMusicReferences encapsula la lógica de negocio para validar los metadatos musicales
func ValidateMusicReferences(refs []models.MusicReference) error {
	for _, ref := range refs {
		if ref.ReferenceType != models.TrackReference &&
			ref.ReferenceType != models.ArtistReference &&
			ref.ReferenceType != models.AlbumReference {
			return fmt.Errorf("tipo de referencia inválido: %s", ref.ReferenceType)
		}

		if ref.Provider != "spotify" && ref.Provider != "youtube" {
			return fmt.Errorf("proveedor no soportado: %s", ref.Provider)
		}

		if ref.ReferenceID == "" {
			return errors.New("el reference_id no puede estar vacío")
		}

		// Validación estricta del formato del ID según el proveedor
		if ref.Provider == "spotify" {
			if !spotifyIDRegex.MatchString(ref.ReferenceID) {
				return fmt.Errorf("formato de spotify_id inválido: %s", ref.ReferenceID)
			}
		}
	}
	return nil
}