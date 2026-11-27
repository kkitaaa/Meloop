from collections import Counter

from scipy.sparse import csr_matrix, vstack

from app.models.recommendation_model import RecommendationModel
from app.schemas.recommendation_schema import (
    CandidateProfile,
    MusicCatalogItem,
    MusicRecommendationRequest,
    RecommendationItem,
    RecommendationRequest,
    RecommendationResponse,
)
from data_processing import UserPreferencesPipeline

LIKE_FEATURE_WEIGHT = 0.5
MAX_LIKE_FEATURE_WEIGHT = 2.0


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

    @property
    def is_ready(self) -> bool:
        return self.model is not None and self.preferences_pipeline is not None

    def generate_music_recommendations(
        self, request: MusicRecommendationRequest
    ) -> RecommendationResponse:
        liked_items = {
            interaction.target_id
            for interaction in request.interactions
            if interaction.type == "like"
        }
        catalog = [item for item in request.catalog if item.item_id not in liked_items]
        if not catalog:
            return RecommendationResponse(
                user_id=request.user_id,
                recommendations=[],
                model="music_content_similarity",
                model_version=self.model.version,
                interaction_count=len(request.interactions),
            )

        user_data = self._profile_data(
            request.profile.genres, request.profile.artists, request.profile.songs
        )
        user_data["user_id"] = request.user_id
        base_user = self.preferences_pipeline.transform(user_data)
        explicit_features = set(base_user.preference_features)
        explicit_features.update(
            value.strip().casefold() for value in request.preferences if value.strip()
        )
        inferred_feature_counts: Counter[str] = Counter()
        seen_liked_items: set[str] = set()
        for item in request.catalog:
            if item.item_id not in liked_items or item.item_id in seen_liked_items:
                continue
            seen_liked_items.add(item.item_id)
            liked_profile = self.preferences_pipeline.transform(
                self._profile_data(item.profile.genres, item.profile.artists, item.profile.songs)
            ).normalized_data.iloc[0]
            user_data["genres"].extend(liked_profile["genres"])
            user_data["artists"].extend(liked_profile["artists"])
            inferred_feature_counts.update(f"genre:{genre}" for genre in liked_profile["genres"])
            inferred_feature_counts.update(
                f"artist:{artist}" for artist in liked_profile["artists"]
            )

        processed_user = self.preferences_pipeline.transform(user_data)
        preference_features = set(processed_user.preference_features) | explicit_features
        if not preference_features:
            return self._catalog_fallback(request, catalog)

        feature_weights = {feature: 1.0 for feature in base_user.preference_features}
        for feature, count in inferred_feature_counts.items():
            feature_weights[feature] = (1.0 if feature in explicit_features else 0.0) + min(
                count * LIKE_FEATURE_WEIGHT, MAX_LIKE_FEATURE_WEIGHT
            )
        processed_user = self.preferences_pipeline.transform(
            user_data, preference_weights=feature_weights
        )
        processed_catalog = [
            self.preferences_pipeline.transform(
                self._profile_data(item.profile.genres, item.profile.artists, item.profile.songs)
            )
            for item in catalog
        ]
        feature_names = tuple(
            sorted(
                preference_features
                | {feature for item in processed_catalog for feature in item.preference_features}
            )
        )
        user_vector = self._align_vector(
            processed_user.preference_matrix, processed_user.preference_features, feature_names
        )
        for feature in preference_features - set(processed_user.preference_features):
            user_vector[0, feature_names.index(feature)] = feature_weights.get(feature, 1.0)
        catalog_vectors = vstack(
            [
                self._align_vector(item.preference_matrix, item.preference_features, feature_names)
                for item in processed_catalog
            ]
        )
        ranked = self.model.recommend_catalog(
            user_vector,
            catalog_vectors,
            [item.item_id for item in catalog],
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
            model_version=self.model.version,
            interaction_count=len(request.interactions),
        )

    def _catalog_fallback(
        self, request: MusicRecommendationRequest, catalog: list[MusicCatalogItem]
    ) -> RecommendationResponse:
        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=[
                RecommendationItem(
                    item_id=item.item_id,
                    score=0.0,
                    reason="Sugerencia de descubrimiento mientras se recopilan preferencias musicales",
                )
                for item in catalog[: request.limit]
            ],
            model="music_catalog_fallback",
            model_version=self.model.version,
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
            item_id = f"generated:{request.user_id}:{recommendation_offset + index}"
            while item_id in liked_items:
                recommendation_offset += 1
                item_id = f"generated:{request.user_id}:{recommendation_offset + index}"
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
            model_version=self.model.version,
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
            model_version=self.model.version,
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

    def _empty_response(
        self, request: RecommendationRequest, reason: str
    ) -> RecommendationResponse:
        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=[],
            model="profile_compatibility",
            model_version=self.model.version,
            interaction_count=len(request.interactions),
        )
