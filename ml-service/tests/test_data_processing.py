import numpy as np

from data_processing import UserPreferencesPipeline


def test_transform_normalizes_deduplicates_and_vectorizes_preferences():
    processed = UserPreferencesPipeline().transform(
        {
            "user_id": 7,
            "genres": [" Rock ", "rock", "MÚSICA  electrónica"],
            "favorite_artists": ["Beyonce", " beyonce ", "Daft Punk"],
            "friends": [12, "12", " 18 "],
        }
    )

    row = processed.normalized_data.iloc[0]
    assert row["genres"] == ["música electrónica", "rock"]
    assert row["artists"] == ["beyonce", "daft punk"]
    assert row["friends"] == ["12", "18"]
    assert processed.preference_features == (
        "artist:beyonce",
        "artist:daft punk",
        "genre:música electrónica",
        "genre:rock",
    )
    assert processed.preference_matrix.shape == (1, 4)
    assert np.array_equal(processed.preference_matrix.toarray(), [[1, 1, 1, 1]])
    assert processed.friend_features == ("friend:12", "friend:18")
    assert np.array_equal(processed.friend_matrix.toarray(), [[1, 1]])


def test_transform_accepts_aliases_and_empty_values():
    processed = UserPreferencesPipeline().transform(
        {"user_id": 9, "music_genres": None, "artists": "Jazz", "friend_ids": []}
    )

    assert processed.normalized_data.iloc[0]["genres"] == []
    assert processed.preference_features == ("artist:jazz",)
    assert processed.preference_matrix.shape == (1, 1)
    assert processed.friend_matrix.shape == (1, 0)
