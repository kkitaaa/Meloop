from collections.abc import Sequence
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


@dataclass
class RecommendationModel:
    name: str = "profile_compatibility"
    version: str = "1.0.0"

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
            for name, first_value, second_value in zip(
                feature_names, first, second, strict=True
            )
            if first_value > 0 and second_value > 0
        )
