from collections.abc import Callable, Sequence
from dataclasses import dataclass

import numpy as np
from scipy.sparse import issparse, spmatrix
from sklearn.metrics.pairwise import cosine_similarity


@dataclass(frozen=True)
class CompatibilityResult:
    percentage: float
    matching_features: tuple[str, ...]

    @property
    def reason(self) -> str:
        if not self.matching_features:
            return "No se encontraron preferencias musicales o sociales en común"

        labels = ", ".join(self._display_feature(feature) for feature in self.matching_features)
        return f"Ambos escuchan a {labels}"

    @staticmethod
    def _display_feature(feature: str) -> str:
        _, separator, value = feature.partition(":")
        return value.title() if separator else feature


@dataclass(frozen=True)
class CatalogRecommendation:
    item_id: int
    score: float
    matching_features: tuple[str, ...] = ()


CollaborativeScorer = Callable[[Sequence[float] | np.ndarray | spmatrix, Sequence[int]], Sequence[float]]


@dataclass
class RecommendationModel:
    name: str = "profile_compatibility"
    version: str = "1.0.0"
    collaborative_scorer: CollaborativeScorer | None = None

    def compare_profiles(
        self,
        first_vector: Sequence[float] | np.ndarray | spmatrix,
        second_vector: Sequence[float] | np.ndarray | spmatrix,
        feature_names: Sequence[str] = (),
    ) -> CompatibilityResult:
        """Compare two generated profile vectors and explain their overlap."""
        first = self._as_row_vector(first_vector)
        second = self._as_row_vector(second_vector)

        if first.shape != second.shape:
            raise ValueError("Los vectores de perfil deben tener la misma dimensión")

        if first.shape[1] == 0:
            return CompatibilityResult(percentage=0.0, matching_features=())

        similarity = float(cosine_similarity(first, second)[0, 0])
        percentage = round(float(np.clip(similarity, 0.0, 1.0)) * 100, 2)
        first_values = first.toarray().ravel() if issparse(first) else first[0]
        second_values = second.toarray().ravel() if issparse(second) else second[0]
        matching_features = self._matching_features(first_values, second_values, feature_names)

        return CompatibilityResult(
            percentage=percentage,
            matching_features=matching_features,
        )

    def recommend_catalog(
        self,
        user_vector: Sequence[float] | np.ndarray | spmatrix,
        catalog_vectors: Sequence[Sequence[float]] | np.ndarray | spmatrix,
        item_ids: Sequence[int],
        feature_names: Sequence[str] = (),
        limit: int = 10,
    ) -> tuple[CatalogRecommendation, ...]:
        """Rank catalog items by content similarity with the user's vector.

        ``catalog_vectors`` must use the same feature columns as ``user_vector``.
        A collaborative scorer can be injected later; when present, its scores are
        averaged with the content scores without changing this method's contract.
        """
        if limit < 1:
            raise ValueError("El límite debe ser mayor que cero")
        if len(item_ids) != self._row_count(catalog_vectors):
            raise ValueError("Debe existir un ID por cada elemento del catálogo")

        user = self._as_row_vector(user_vector)
        catalog = self._as_catalog_matrix(catalog_vectors)
        if user.shape[1] != catalog.shape[1]:
            raise ValueError("El usuario y el catálogo deben tener la misma dimensión")
        if feature_names and len(feature_names) != user.shape[1]:
            raise ValueError("Debe existir un nombre por cada componente del vector")
        if catalog.shape[0] == 0:
            return ()

        content_scores = cosine_similarity(user, catalog)[0]
        scores = np.asarray(content_scores, dtype=float)
        if self.collaborative_scorer is not None:
            collaborative_scores = np.asarray(
                self.collaborative_scorer(user_vector, item_ids), dtype=float
            )
            if collaborative_scores.shape != scores.shape:
                raise ValueError("El filtro colaborativo debe devolver un puntaje por elemento")
            scores = (scores + collaborative_scores) / 2

        user_values = user.toarray().ravel() if issparse(user) else user[0]
        ranked_indices = sorted(range(len(item_ids)), key=lambda index: (-scores[index], index))
        return tuple(
            CatalogRecommendation(
                item_id=item_ids[index],
                score=round(float(np.clip(scores[index], 0.0, 1.0)), 4),
                matching_features=self._matching_features(
                    user_values,
                    catalog.getrow(index).toarray().ravel()
                    if issparse(catalog)
                    else catalog[index],
                    feature_names,
                ),
            )
            for index in ranked_indices[:limit]
        )

    @staticmethod
    def _as_row_vector(
        vector: Sequence[float] | np.ndarray | spmatrix,
    ) -> np.ndarray | spmatrix:
        values = vector.astype(float) if issparse(vector) else np.asarray(vector, dtype=float)
        if values.ndim == 1:
            values = values.reshape(1, -1)
        if values.ndim != 2 or values.shape[0] != 1:
            raise ValueError("Cada perfil debe ser un vector unidimensional")
        return values

    @staticmethod
    def _as_catalog_matrix(
        vectors: Sequence[Sequence[float]] | np.ndarray | spmatrix,
    ) -> np.ndarray | spmatrix:
        values = vectors.astype(float) if issparse(vectors) else np.asarray(vectors, dtype=float)
        if values.ndim != 2:
            raise ValueError("El catálogo debe ser una matriz bidimensional")
        return values

    @staticmethod
    def _row_count(vectors: Sequence[Sequence[float]] | np.ndarray | spmatrix) -> int:
        if not hasattr(vectors, "shape"):
            return len(vectors)
        return vectors.shape[0]

    @staticmethod
    def _matching_features(
        first: np.ndarray,
        second: np.ndarray,
        feature_names: Sequence[str],
    ) -> tuple[str, ...]:
        if len(feature_names) != len(first):
            if feature_names:
                raise ValueError("Debe existir un nombre por cada componente del vector")
            return ()

        return tuple(
            name
            for name, first_value, second_value in zip(feature_names, first, second, strict=True)
            if first_value > 0 and second_value > 0
        )
