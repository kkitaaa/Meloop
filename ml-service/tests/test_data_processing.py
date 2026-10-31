import numpy as np

from data_processing import UserPreferencesPipeline, fetch_preferences


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


def test_transform_includes_songs_in_the_model_matrix():
    processed = UserPreferencesPipeline().transform(
        {"user_id": 10, "tracks": [" Song A ", None, "song a"]}
    )

    assert processed.normalized_data.iloc[0]["songs"] == ["song a"]
    assert processed.preference_features == ("song:song a",)
    assert np.array_equal(processed.preference_matrix.toarray(), [[1]])


def test_fetch_preferences_adds_user_id_and_decodes_json(monkeypatch):
    class FakeResponse:
        def __enter__(self):
            return self

        def __exit__(self, *_args):
            return False

        def read(self):
            return b'{"genres": ["Rock"]}'

    captured = {}

    def fake_urlopen(request, timeout):
        captured["url"] = request.full_url
        captured["timeout"] = timeout
        return FakeResponse()

    monkeypatch.setattr("data_processing.source.urlopen", fake_urlopen)

    assert fetch_preferences("http://go-service/preferences?format=raw", 42) == {"genres": ["Rock"]}
    assert captured == {
        "url": "http://go-service/preferences?format=raw&user_id=42",
        "timeout": 10.0,
    }
