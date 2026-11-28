package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
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

type PostRepository interface {
	CreatePost(context.Context, string, string, *string) (*repositories.Post, error)
	UpdatePost(context.Context, string, string, string) error
	DeletePost(context.Context, string, string) error
	GetPostByIDExcluding(context.Context, string, string, []string) (*repositories.Post, error)
	GetPostsByAuthorExcluding(context.Context, string, string, int, int, []string) ([]repositories.Post, error)
	SearchPostsExcluding(context.Context, string, string, int, int, []string) ([]repositories.Post, error)
	AddLike(context.Context, string, string) error
	GetPostAuthorID(context.Context, string) (string, error)
	SavePost(context.Context, string, string) error
	UnsavePost(context.Context, string, string) error
	GetSavedPosts(context.Context, string, []string, int, int) ([]repositories.Post, error)
	RemoveLike(context.Context, string, string) error
}

type PostService struct {
	repo                PostRepository
	blockedUsers        BlockedUsersProvider
	friendshipValidator FriendshipValidator
	postShareSender     PostShareSender
}

func NewPostService(repo PostRepository, blockedUsers ...BlockedUsersProvider) *PostService {
	service := &PostService{repo: repo}
	if len(blockedUsers) > 0 {
		service.blockedUsers = blockedUsers[0]
	}
	return service
}

func NewPostServiceWithSharing(
	repo PostRepository,
	blockedUsers BlockedUsersProvider,
	friendshipValidator FriendshipValidator,
	postShareSender PostShareSender,
) *PostService {
	service := NewPostService(repo, blockedUsers)
	service.friendshipValidator = friendshipValidator
	service.postShareSender = postShareSender
	return service
}

// CreatePost valida y envía la nueva publicación al repositorio
func (s *PostService) CreatePost(ctx context.Context, authorID string, content string, mediaURLs []string) (*repositories.Post, error) {
	if content == "" {
		return nil, errors.New("el contenido de la publicación no puede estar vacío")
	}

	var url *string
	if len(mediaURLs) > 0 && mediaURLs[0] != "" {
		url = &mediaURLs[0]
	}

	return s.repo.CreatePost(ctx, authorID, content, url)
}

// EditPost valida los datos antes de enviarlos a la base de datos (RF-17)
func (s *PostService) EditPost(ctx context.Context, postID string, authorID string, newContent string) error {
	if newContent == "" {
		return errors.New("el contenido de la publicación no puede estar vacío")
	}

	return s.repo.UpdatePost(ctx, postID, authorID, newContent)
}

// DeletePost procesa la solicitud de eliminación (RF-18)
func (s *PostService) DeletePost(ctx context.Context, postID string, authorID string) error {
	return s.repo.DeletePost(ctx, postID, authorID)
}

func (s *PostService) GetPost(ctx context.Context, postID string, requesterID string) (*repositories.Post, error) {
	blocked, err := s.blockedWith(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPostByIDExcluding(ctx, postID, requesterID, blocked)
}

func (s *PostService) GetAuthorPosts(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 10 // Máximo por defecto seguro
	}
	if offset < 0 {
		offset = 0
	}
	blocked, err := s.blockedWith(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPostsByAuthorExcluding(ctx, authorID, requesterID, limit, offset, blocked)
}

func (s *PostService) SearchPosts(ctx context.Context, query string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	blocked, err := s.blockedWith(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	return s.repo.SearchPostsExcluding(ctx, query, requesterID, limit, offset, blocked)
}

// RF-20: Agregar Me gusta
func (s *PostService) AddLike(ctx context.Context, postID string, userID string) error {
	if err := s.EnsureCanInteract(ctx, userID, postID); err != nil {
		return err
	}
	return s.repo.AddLike(ctx, postID, userID)
}

var ErrBlocked = errors.New("BLOCKED")
var ErrNotFriend = errors.New("recipient is not a friend")
var ErrSharingUnavailable = errors.New("sharing dependency is unavailable")

type FriendshipValidator interface {
	ValidateFriendship(context.Context, string, string, string) error
}

type PostShareSender interface {
	SendPostShare(context.Context, string, string, string) (string, error)
}

func (s *PostService) EnsureCanInteract(ctx context.Context, userID, postID string) error {
	authorID, err := s.repo.GetPostAuthorID(ctx, postID)
	if err != nil {
		return err
	}
	return s.EnsureCanInteractWithUser(ctx, userID, authorID)
}

func (s *PostService) EnsureCanInteractWithUser(ctx context.Context, userID, targetUserID string) error {
	blocked, err := s.blockedWithFresh(ctx, userID)
	if err != nil {
		return err
	}
	for _, blockedID := range blocked {
		if blockedID == targetUserID {
			return ErrBlocked
		}
	}
	return nil
}

func (s *PostService) SavePost(ctx context.Context, userID, postID string) error {
	if err := s.EnsureCanInteract(ctx, userID, postID); err != nil {
		return err
	}
	return s.repo.SavePost(ctx, userID, postID)
}

func (s *PostService) UnsavePost(ctx context.Context, userID, postID string) error {
	return s.repo.UnsavePost(ctx, userID, postID)
}

func (s *PostService) GetSavedPosts(ctx context.Context, userID string, limit, offset int) ([]repositories.Post, bool, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	blocked, err := s.blockedWith(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	posts, err := s.repo.GetSavedPosts(ctx, userID, blocked, limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}
	return posts, hasMore, nil
}

func (s *PostService) SharePost(ctx context.Context, userID, postID, friendID, authorization string) (string, error) {
	userID = strings.TrimSpace(userID)
	friendID = strings.TrimSpace(friendID)
	if userID == "" || friendID == "" || friendID == userID {
		return "", ErrNotFriend
	}
	if err := s.EnsureCanInteract(ctx, userID, postID); err != nil {
		return "", err
	}
	if err := s.EnsureCanInteractWithUser(ctx, userID, friendID); err != nil {
		return "", err
	}
	if s.friendshipValidator == nil || s.postShareSender == nil {
		return "", ErrSharingUnavailable
	}
	if err := s.friendshipValidator.ValidateFriendship(ctx, userID, friendID, authorization); err != nil {
		return "", err
	}
	messageID, err := s.postShareSender.SendPostShare(ctx, userID, friendID, postID)
	if err != nil {
		return "", err
	}
	return messageID, nil
}

func (s *PostService) blockedWith(ctx context.Context, userID string) ([]string, error) {
	if s.blockedUsers == nil {
		return nil, errors.New("blocked users provider is not configured")
	}
	return s.blockedUsers.BlockedWith(ctx, userID)
}

func (s *PostService) blockedWithFresh(ctx context.Context, userID string) ([]string, error) {
	if s.blockedUsers == nil {
		return nil, errors.New("blocked users provider is not configured")
	}
	if provider, ok := s.blockedUsers.(FreshBlockedUsersProvider); ok {
		return provider.BlockedWithFresh(ctx, userID)
	}
	return s.blockedUsers.BlockedWith(ctx, userID)
}

// RF-20: Quitar Me gusta
func (s *PostService) RemoveLike(ctx context.Context, postID string, userID string) error {
	return s.repo.RemoveLike(ctx, postID, userID)
}
