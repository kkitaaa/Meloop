from __future__ import annotations

import re
import unicodedata
from collections.abc import Mapping
from dataclasses import dataclass
from typing import Any

import numpy as np
import pandas as pd
from scipy.sparse import csr_matrix
from sklearn.preprocessing import MultiLabelBinarizer


@dataclass(frozen=True)
class ProcessedUserPreferences:
    """Numeric representation of one user's musical preferences."""

    normalized_data: pd.DataFrame
    preference_matrix: csr_matrix
    preference_features: tuple[str, ...]
    friend_matrix: csr_matrix
    friend_features: tuple[str, ...]


class UserPreferencesPipeline:
    """Normalize raw preference JSON and convert it into model-ready matrices."""

    _FIELD_ALIASES = {
        "genres": "genres",
        "music_genres": "genres",
        "artists": "artists",
        "favorite_artists": "artists",
        "friends": "friends",
        "friend_ids": "friends",
    }

    def transform(self, raw_data: Mapping[str, Any]) -> ProcessedUserPreferences:
        if not isinstance(raw_data, Mapping):
            raise TypeError("raw_data debe ser un objeto JSON con preferencias del usuario")

        user_id = raw_data.get("user_id")
        genres = self._normalize_text_values(self._get_values(raw_data, "genres"))
        artists = self._normalize_text_values(self._get_values(raw_data, "artists"))
        friends = self._normalize_friend_values(self._get_values(raw_data, "friends"))

        normalized_data = pd.DataFrame(
            [
                {
                    "user_id": user_id,
                    "genres": genres,
                    "artists": artists,
                    "friends": friends,
                }
            ]
        )

        preference_labels = [
            [f"genre:{genre}" for genre in genres] + [f"artist:{artist}" for artist in artists]
        ]
        preference_matrix, preference_features = self._encode(preference_labels)

        friend_labels = [[f"friend:{friend}" for friend in friends]]
        friend_matrix, friend_features = self._encode(friend_labels)

        return ProcessedUserPreferences(
            normalized_data=normalized_data,
            preference_matrix=preference_matrix,
            preference_features=preference_features,
            friend_matrix=friend_matrix,
            friend_features=friend_features,
        )

    def _get_values(self, raw_data: Mapping[str, Any], field: str) -> list[Any]:
        for key, canonical_key in self._FIELD_ALIASES.items():
            if canonical_key == field and key in raw_data:
                values = raw_data[key]
                if values is None:
                    return []
                if isinstance(values, (str, bytes)) or not isinstance(values, list):
                    return [values]
                return values
        return []

    @staticmethod
    def _normalize_text(value: Any) -> str | None:
        if not isinstance(value, str):
            return None
        normalized = unicodedata.normalize("NFKC", value).strip()
        normalized = re.sub(r"\s+", " ", normalized)
        return normalized.casefold() or None

    def _normalize_text_values(self, values: list[Any]) -> list[str]:
        return self._unique_sorted(
            normalized
            for value in values
            if (normalized := self._normalize_text(value)) is not None
        )

    def _normalize_friend_values(self, values: list[Any]) -> list[str]:
        normalized_values = (
            str(value).strip() for value in values if value is not None and str(value).strip()
        )
        return self._unique_sorted(normalized_values)

    @staticmethod
    def _unique_sorted(values: Any) -> list[str]:
        return sorted(set(values))

    @staticmethod
    def _encode(labels: list[list[str]]) -> tuple[csr_matrix, tuple[str, ...]]:
        encoder = MultiLabelBinarizer(sparse_output=True)
        matrix = encoder.fit_transform(labels).tocsr().astype(np.float32)
        return matrix, tuple(str(label) for label in encoder.classes_)
