package repositories

import "context"

type ExperienceRepository struct {
	// Aquí irá tu conexión a la base de datos más adelante
}

func NewExperienceRepository() *ExperienceRepository {
	return &ExperienceRepository{}
}

// AddExperience es la función que llama el servicio
func (r *ExperienceRepository) AddExperience(ctx context.Context, userID string, xpToAdd int) error {
	// Lógica futura de actualización en la BD
	return nil
}
