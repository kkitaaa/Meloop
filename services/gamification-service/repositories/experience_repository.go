package repositories
import "context"

type ExperienceRepository struct{}

func NewExperienceRepository() *ExperienceRepository {
	return &ExperienceRepository{}
}

func (r *ExperienceRepository) AddExperience(ctx context.Context, userID string, xpToAdd int) error {
	return nil
}