package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/meloop/social-service/models"
)

// GetFriendSuggestions obtiene la lista de sugerencias de amistad elegibles y ordenadas por afinidad para el usuario.
func (s *friendshipService) GetFriendSuggestions(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidUser
	}

	// 1. Obtener candidatos elegibles desde el repositorio
	// Excluye estrictamente: propio usuario, amigos aceptados, bloqueos bidireccionales y cuentas inactivas/moderadas
	candidates, err := s.repo.GetEligibleCandidates(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	if len(candidates) == 0 {
		return []models.FriendSuggestion{}, nil
	}

	// Filtrar usuarios activos no nulos
	validCandidates := make([]models.CandidateUser, 0, len(candidates))
	candidateIDs := make([]string, 0, len(candidates))
	for _, cand := range candidates {
		candID := strings.TrimSpace(cand.IDUsuario)
		if candID == "" || candID == userID || !cand.IsActive {
			continue
		}
		validCandidates = append(validCandidates, cand)
		candidateIDs = append(candidateIDs, candID)
	}

	if len(validCandidates) == 0 {
		return []models.FriendSuggestion{}, nil
	}

	// 2. Obtener conteo de amigos en común en una sola consulta
	mutualFriendsMap, err := s.repo.GetMutualFriendsCount(ctx, userID, candidateIDs)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	if mutualFriendsMap == nil {
		mutualFriendsMap = make(map[string]int)
	}

	// 3. Obtener perfiles musicales en lote y calcular compatibilidad mediante motor RF-15
	allUserIDs := append([]string{userID}, candidateIDs...)
	scoresMap := make(map[string]*models.CompatibilityScore, len(validCandidates))

	if s.compatRepo != nil {
		profilesMap, err := s.compatRepo.GetBatchMusicalProfiles(ctx, allUserIDs)
		if err != nil {
			return nil, err
		}
		userProfile := profilesMap[userID]
		for _, cand := range validCandidates {
			candProfile := profilesMap[cand.IDUsuario]
			// Reutilizar exactamente la función de RF-15 sin duplicar ni inventar nuevas fórmulas
			score := CalculateCompatibilityScore(userProfile, candProfile)
			scoresMap[cand.IDUsuario] = score
		}
	} else {
		userProfile := &models.UserMusicalData{UserID: userID}
		for _, cand := range validCandidates {
			candProfile := &models.UserMusicalData{UserID: cand.IDUsuario}
			score := CalculateCompatibilityScore(userProfile, candProfile)
			scoresMap[cand.IDUsuario] = score
		}
	}

	// 4. Ranking determinista combinando señales de afinidad y cálculo de motivo
	suggestions := RankSuggestions(validCandidates, scoresMap, mutualFriendsMap)

	return suggestions, nil
}

// DetermineSuggestionReason determina la principal coincidencia real entre el usuario y el candidato
// y construye un texto descriptivo y específico sin inventar coincidencias inexistentes.
func DetermineSuggestionReason(score *models.CompatibilityScore, mutualFriends int) string {
	if score != nil {
		// Prioridad 1: Coincidencia de Artista compartido
		if len(score.CommonArtists) > 0 {
			artist := strings.TrimSpace(score.CommonArtists[0])
			if artist != "" {
				return fmt.Sprintf("Ambos escuchan a %s", artist)
			}
		}

		// Prioridad 2: Coincidencia de Canción compartida
		if len(score.CommonTracks) > 0 {
			track := strings.TrimSpace(score.CommonTracks[0])
			if track != "" {
				return fmt.Sprintf("Ambos tienen interés en %s", track)
			}
		}
	}

	// Prioridad 3: Amigos en común
	if mutualFriends > 0 {
		if mutualFriends == 1 {
			return "Tienen 1 amigo en común"
		}
		return fmt.Sprintf("Tienen %d amigos en común", mutualFriends)
	}

	if score != nil {
		// Prioridad 4: Actividad/Interacciones musicales compartidas
		if len(score.CommonInteractedTracks) > 0 {
			song := strings.TrimSpace(score.CommonInteractedTracks[0])
			if song != "" {
				return fmt.Sprintf("Ambos interactuaron con la canción %s", song)
			}
		}

		// Prioridad 5: Género musical compartido
		if len(score.CommonGenres) > 0 {
			genre := strings.TrimSpace(score.CommonGenres[0])
			if genre != "" {
				return fmt.Sprintf("Ambos disfrutan del género %s", genre)
			}
		}
	}

	// Fallback para candidatos sin señales directas compartidas
	return "Sugerencia de la comunidad"
}

// CalculateAffinityScore calcula la puntuación combinada de afinidad para el ranking.
func CalculateAffinityScore(score *models.CompatibilityScore, mutualFriends int) float64 {
	matchScore := 0.0
	if score != nil {
		matchScore = float64(score.MatchScore)
	}
	return matchScore + (float64(mutualFriends) * 10.0)
}

// RankSuggestions construye el listado de sugerencias de amistad ordenado de mayor a menor afinidad de manera determinista.
func RankSuggestions(candidates []models.CandidateUser, scores map[string]*models.CompatibilityScore, mutualFriends map[string]int) []models.FriendSuggestion {
	suggestions := make([]models.FriendSuggestion, 0, len(candidates))

	for _, cand := range candidates {
		score := scores[cand.IDUsuario]
		mutual := mutualFriends[cand.IDUsuario]

		matchScore := 0
		var commonGenres []string
		var commonArtists []string
		var commonTracks []string
		var commonInteractions []string

		if score != nil {
			matchScore = score.MatchScore
			commonGenres = score.CommonGenres
			commonArtists = score.CommonArtists
			commonTracks = score.CommonTracks
			commonInteractions = score.CommonInteractedTracks
		}

		reason := DetermineSuggestionReason(score, mutual)

		sugg := models.FriendSuggestion{
			IDUsuario:              cand.IDUsuario,
			Username:               cand.Username,
			Correo:                 cand.Correo,
			IDNivel:                cand.IDNivel,
			Experiencia:            cand.Experiencia,
			MatchScore:             matchScore,
			MutualFriends:          mutual,
			CommonGenres:           commonGenres,
			CommonArtists:          commonArtists,
			CommonTracks:           commonTracks,
			CommonInteractedTracks: commonInteractions,
			Motivo:                 reason,
		}
		suggestions = append(suggestions, sugg)
	}

	sort.SliceStable(suggestions, func(i, j int) bool {
		scoreI := scores[suggestions[i].IDUsuario]
		scoreJ := scores[suggestions[j].IDUsuario]

		// 1. Mayor puntuación de afinidad
		affI := CalculateAffinityScore(scoreI, suggestions[i].MutualFriends)
		affJ := CalculateAffinityScore(scoreJ, suggestions[j].MutualFriends)
		if affI != affJ {
			return affI > affJ
		}

		// 2. Mayor compatibilidad musical RF-15
		if suggestions[i].MatchScore != suggestions[j].MatchScore {
			return suggestions[i].MatchScore > suggestions[j].MatchScore
		}

		// 3. Mayor cantidad de amigos en común
		if suggestions[i].MutualFriends != suggestions[j].MutualFriends {
			return suggestions[i].MutualFriends > suggestions[j].MutualFriends
		}

		var artistCountI, artistCountJ int
		var trackCountI, trackCountJ int
		var interactedCountI, interactedCountJ int
		var genreCountI, genreCountJ int

		if scoreI != nil {
			artistCountI = len(scoreI.CommonArtists)
			trackCountI = len(scoreI.CommonTracks)
			interactedCountI = len(scoreI.CommonInteractedTracks)
			genreCountI = len(scoreI.CommonGenres)
		}
		if scoreJ != nil {
			artistCountJ = len(scoreJ.CommonArtists)
			trackCountJ = len(scoreJ.CommonTracks)
			interactedCountJ = len(scoreJ.CommonInteractedTracks)
			genreCountJ = len(scoreJ.CommonGenres)
		}

		// 4. Mayor cantidad de artistas en común
		if artistCountI != artistCountJ {
			return artistCountI > artistCountJ
		}

		// 5. Mayor cantidad de canciones en común
		if trackCountI != trackCountJ {
			return trackCountI > trackCountJ
		}

		// 6. Mayor cantidad de interacciones compartidas (historial de actividad)
		if interactedCountI != interactedCountJ {
			return interactedCountI > interactedCountJ
		}

		// 7. Mayor cantidad de géneros en común
		if genreCountI != genreCountJ {
			return genreCountI > genreCountJ
		}

		// 8. Desempate determinista por ID de usuario
		return suggestions[i].IDUsuario < suggestions[j].IDUsuario
	})

	return suggestions
}
