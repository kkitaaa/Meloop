from scipy.sparse import csr_matrix, vstack

from app.models.recommendation_model import RecommendationModel
from app.schemas.recommendation_schema import (
    CandidateProfile,
    MusicRecommendationRequest,
    RecommendationItem,
    RecommendationRequest,
    RecommendationResponse,
)
from data_processing import UserPreferencesPipeline


class CompatibilityReason:
    @staticmethod
    def from_features(features: tuple[str, ...]) -> str:
        if not features:
            return "No se encontraron preferencias musicales en común"
        labels = ", ".join(feature.partition(":")[2].title() for feature in features)
        return f"Coincide con tus preferencias: {labels}"


class RecommendationService:
    def __init__(self) -> None:
        self.model = RecommendationModel()
        self.model_name = self.model.name
        self.preferences_pipeline = UserPreferencesPipeline()

    def generate_music_recommendations(
        self, request: MusicRecommendationRequest
    ) -> RecommendationResponse:
        user_data = self._profile_data(request.profile.genres, request.profile.artists, request.profile.songs)
        user_data["user_id"] = request.user_id
        processed_user = self.preferences_pipeline.transform(user_data)
        if not processed_user.preference_features:
            return self._empty_response(request, "El usuario no tiene historial musical suficiente")

        processed_catalog = [
            self.preferences_pipeline.transform(
                self._profile_data(item.profile.genres, item.profile.artists, item.profile.songs)
            )
            for item in request.catalog
        ]
        feature_names = tuple(
            sorted(
                set(processed_user.preference_features)
                | {
                    feature
                    for item in processed_catalog
                    for feature in item.preference_features
                }
            )
        )
        user_vector = self._align_vector(
            processed_user.preference_matrix, processed_user.preference_features, feature_names
        )
        catalog_vectors = vstack(
            [
                self._align_vector(item.preference_matrix, item.preference_features, feature_names)
                for item in processed_catalog
            ]
        )
        ranked = self.model.recommend_catalog(
            user_vector,
            catalog_vectors,
            [item.item_id for item in request.catalog],
            feature_names,
            request.limit,
        )
        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=[
                RecommendationItem(
                    item_id=item.item_id,
                    score=item.score,
                    reason=CompatibilityReason.from_features(item.matching_features),
                )
                for item in ranked
            ],
            model="music_content_similarity",
            interaction_count=len(request.interactions),
        )

    def generate(self, request: RecommendationRequest) -> RecommendationResponse:
        if request.candidate_profiles:
            return self.generate_friend_recommendations(request)

        recommendations: list[RecommendationItem] = []
        interaction_count = len(request.interactions)
        liked_items = {
            interaction.target_id
            for interaction in request.interactions
            if interaction.type == "like"
        }
        recommendation_offset = interaction_count * 10

        for index in range(1, min(request.limit, 5) + 1):
            item_id = request.user_id + recommendation_offset + index
            if item_id in liked_items:
                item_id += 1000
            recommendations.append(
                RecommendationItem(
                    item_id=item_id,
                    score=round(max(0.1, 1.0 - (index * 0.1) + interaction_count * 0.01), 2),
                    reason=(
                        "Sugerencia adaptada a la actividad reciente del usuario"
                        if interaction_count
                        else "Sugerencia de prueba basada en preferencias del usuario"
                    ),
                )
            )

        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=recommendations,
            model=self.model_name,
            interaction_count=interaction_count,
        )

    def generate_friend_recommendations(
        self, request: RecommendationRequest
    ) -> RecommendationResponse:
        user_features = self._profile_features(request)
        if not user_features:
            return self._empty_response(request, "El usuario no tiene historial musical suficiente")

        candidates = [
            candidate
            for candidate in request.candidate_profiles
            if candidate.user_id != request.user_id
        ]
        feature_names = sorted(
            user_features
            | {
                feature
                for candidate in candidates
                for feature in self._candidate_features(candidate)
            }
        )
        user_vector = [float(feature in user_features) for feature in feature_names]
        ranked: list[RecommendationItem] = []
        for candidate in candidates:
            candidate_features = self._candidate_features(candidate)
            result = self.model.compare_profiles(
                user_vector,
                [float(feature in candidate_features) for feature in feature_names],
                feature_names,
            )
            ranked.append(
                RecommendationItem(
                    item_id=candidate.user_id,
                    score=result.percentage / 100,
                    reason=result.reason,
                )
            )

        ranked.sort(key=lambda item: (-item.score, item.item_id))
        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=ranked[: request.limit],
            model=self.model_name,
            interaction_count=len(request.interactions),
        )

    @staticmethod
    def _profile_features(request: RecommendationRequest) -> set[str]:
        return RecommendationService._features(
            request.profile.genres,
            request.profile.artists,
            request.profile.songs,
        ) | set(request.preferences)

    @staticmethod
    def _profile_data(genres: list[str], artists: list[str], songs: list[str]) -> dict:
        return {"genres": genres, "artists": artists, "songs": songs}

    @staticmethod
    def _align_vector(
        matrix: csr_matrix, names: tuple[str, ...], feature_names: tuple[str, ...]
    ) -> csr_matrix:
        positions = {name: index for index, name in enumerate(feature_names)}
        indices = [positions[name] for name in names]
        return csr_matrix(
            (matrix.data, ([0] * len(indices), indices)),
            shape=(1, len(feature_names)),
        )

    @staticmethod
    def _candidate_features(candidate: CandidateProfile) -> set[str]:
        return RecommendationService._features(
            candidate.profile.genres,
            candidate.profile.artists,
            candidate.profile.songs,
        )

    @staticmethod
    def _features(genres: list[str], artists: list[str], songs: list[str]) -> set[str]:
        return {
            *{f"genre:{value.casefold().strip()}" for value in genres if value.strip()},
            *{f"artist:{value.casefold().strip()}" for value in artists if value.strip()},
            *{f"song:{value.casefold().strip()}" for value in songs if value.strip()},
        }

    @staticmethod
    def _empty_response(request: RecommendationRequest, reason: str) -> RecommendationResponse:
        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=[],
            model="profile_compatibility",
            interaction_count=len(request.interactions),
        )
