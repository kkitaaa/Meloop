from fastapi.testclient import TestClient

from app import main as main_module
from app.main import app

client = TestClient(app)


def precision_at_k(recommended_ids, relevant_ids, k):
    selected = recommended_ids[:k]
    return sum(item_id in relevant_ids for item_id in selected) / k


def recall_at_k(recommended_ids, relevant_ids, k):
    if not relevant_ids:
        return 0.0
    selected = recommended_ids[:k]
    return sum(item_id in relevant_ids for item_id in selected) / len(relevant_ids)


def test_health_reports_live_process():
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok", "service": "ml-service"}


def test_ready_reports_loaded_model_and_data_pipeline():
    response = client.get("/ready")

    assert response.status_code == 200
    assert response.json() == {
        "status": "ready",
        "service": "ml-service",
        "model_loaded": True,
        "data_pipeline_loaded": True,
        "model_version": "1.0.0",
    }


def test_ready_returns_service_unavailable_when_model_is_missing(monkeypatch):
    class UnavailableService:
        model = None
        preferences_pipeline = object()
        is_ready = False

    monkeypatch.setattr(main_module, "recommendation_service", UnavailableService())

    response = client.get("/ready")

    assert response.status_code == 503
    assert response.json()["status"] == "not_ready"
    assert response.json()["model_loaded"] is False


def test_predict_changes_after_new_like():
    initial = client.post("/predict", json={"user_id": "auth-user-42", "limit": 1})
    updated = client.post(
        "/predict",
        json={
            "user_id": "auth-user-42",
            "limit": 1,
            "interactions": [{"type": "like", "target_id": "spotify-track-7"}],
        },
    )

    assert initial.status_code == 200
    assert updated.status_code == 200
    assert updated.json()["interaction_count"] == 1
    assert updated.json()["recommendations"] != initial.json()["recommendations"]


def test_predict_rejects_unknown_interaction_type():
    response = client.post(
        "/predict",
        json={
            "user_id": "auth-user-42",
            "interactions": [{"type": "share", "target_id": "spotify-track-7"}],
        },
    )

    assert response.status_code == 422


def test_friend_recommendations_rank_candidates_and_explain_matches():
    response = client.post(
        "/recommendations/friends",
        json={
            "user_id": "auth-user-42",
            "limit": 2,
            "profile": {"genres": ["rock"], "artists": ["Arctic Monkeys"]},
            "candidate_profiles": [
                {"user_id": "auth-user-7", "profile": {"genres": ["rock"]}},
                {"user_id": "auth-user-8", "profile": {"genres": ["jazz"]}},
                {"user_id": "auth-user-42", "profile": {"genres": ["rock"]}},
            ],
        },
    )

    assert response.status_code == 200
    assert response.json()["recommendations"] == [
        {"item_id": "auth-user-7", "score": 0.7071, "reason": "Ambos escuchan a Rock"},
        {
            "item_id": "auth-user-8",
            "score": 0.0,
            "reason": "No se encontraron preferencias musicales o sociales en común",
        },
    ]


def test_friend_recommendations_return_empty_for_new_user():
    response = client.post(
        "/recommendations/friends",
        json={"user_id": "auth-user-42", "candidate_profiles": [{"user_id": "auth-user-7"}]},
    )

    assert response.status_code == 200
    assert response.json()["recommendations"] == []


def test_music_recommendations_rank_catalog_with_pipeline_and_model():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-42",
            "limit": 2,
            "profile": {"genres": ["rock"], "artists": ["Arctic Monkeys"]},
            "catalog": [
                {
                    "item_id": "spotify-track-101",
                    "profile": {"genres": ["rock"], "artists": ["Arctic Monkeys"]},
                },
                {"item_id": "spotify-track-102", "profile": {"genres": ["rock"]}},
                {"item_id": "spotify-track-103", "profile": {"genres": ["jazz"]}},
            ],
        },
    )

    assert response.status_code == 200
    body = response.json()
    assert body["user_id"] == "auth-user-42"
    assert body["model"] == "music_content_similarity"
    assert body["model_version"] == "1.0.0"
    assert [item["item_id"] for item in body["recommendations"]] == [
        "spotify-track-101",
        "spotify-track-102",
    ]
    assert body["recommendations"][0]["score"] == 1.0


def test_music_recommendations_require_catalog():
    response = client.post(
        "/recommendations/music",
        json={"user_id": "auth-user-42", "profile": {"genres": ["rock"]}},
    )

    assert response.status_code == 422


def test_music_recommendations_exclude_the_liked_song_identifier():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-abc123",
            "profile": {"genres": ["rock"], "songs": ["spotify-track-loved"]},
            "interactions": [{"type": "like", "target_id": "spotify-track-loved"}],
            "catalog": [
                {"item_id": "spotify-track-loved", "profile": {"songs": ["spotify-track-loved"]}},
                {"item_id": "spotify-track-new", "profile": {"genres": ["rock"]}},
            ],
        },
    )

    assert response.status_code == 200
    assert [item["item_id"] for item in response.json()["recommendations"]] == ["spotify-track-new"]


def test_music_recommendations_filter_likes_before_applying_limit():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-abc123",
            "limit": 2,
            "profile": {"genres": ["rock"]},
            "interactions": [
                {"type": "like", "target_id": "spotify-track-loved-1"},
                {"type": "like", "target_id": "spotify-track-loved-2"},
            ],
            "catalog": [
                {
                    "item_id": "spotify-track-loved-1",
                    "profile": {"genres": ["rock"]},
                },
                {
                    "item_id": "spotify-track-loved-2",
                    "profile": {"genres": ["rock"]},
                },
                {"item_id": "spotify-track-new-1", "profile": {"genres": ["rock"]}},
                {"item_id": "spotify-track-new-2", "profile": {"genres": ["rock"]}},
            ],
        },
    )

    assert response.status_code == 200
    assert [item["item_id"] for item in response.json()["recommendations"]] == [
        "spotify-track-new-1",
        "spotify-track-new-2",
    ]


def test_music_recommendations_infer_genre_from_likes_and_measures_quality_at_k():
    liked_tracks = [
        {
            "item_id": f"deftones-liked-{index}",
            "profile": {"genres": ["metal"], "artists": ["Deftones"]},
        }
        for index in range(5)
    ]
    catalog = [
        *liked_tracks,
        {"item_id": "deftones-unheard", "profile": {"genres": ["metal"], "artists": ["Deftones"]}},
        {"item_id": "gojira-unheard", "profile": {"genres": ["metal"], "artists": ["Gojira"]}},
        {"item_id": "pop-unheard", "profile": {"genres": ["pop"], "artists": ["Dua Lipa"]}},
    ]
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-metal",
            "limit": 2,
            "interactions": [
                {"type": "like", "target_id": track["item_id"]} for track in liked_tracks
            ],
            "catalog": catalog,
        },
    )

    assert response.status_code == 200
    body = response.json()
    recommended_ids = [item["item_id"] for item in body["recommendations"]]
    relevant_ids = {"deftones-unheard", "gojira-unheard"}
    assert recommended_ids == ["deftones-unheard", "gojira-unheard"]
    assert precision_at_k(recommended_ids, relevant_ids, 2) == 1.0
    assert recall_at_k(recommended_ids, relevant_ids, 2) == 1.0


def test_music_recommendations_learn_more_than_one_artist_from_likes():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-varied",
            "limit": 2,
            "interactions": [
                {"type": "like", "target_id": "deftones-liked"},
                {"type": "like", "target_id": "radiohead-liked"},
            ],
            "catalog": [
                {
                    "item_id": "deftones-liked",
                    "profile": {"genres": ["metal"], "artists": ["Deftones"]},
                },
                {
                    "item_id": "radiohead-liked",
                    "profile": {"genres": ["alternative"], "artists": ["Radiohead"]},
                },
                {
                    "item_id": "deftones-next",
                    "profile": {"genres": ["metal"], "artists": ["Deftones"]},
                },
                {
                    "item_id": "radiohead-next",
                    "profile": {"genres": ["alternative"], "artists": ["Radiohead"]},
                },
                {
                    "item_id": "unrelated-pop",
                    "profile": {"genres": ["pop"], "artists": ["Dua Lipa"]},
                },
            ],
        },
    )

    assert response.status_code == 200
    assert {item["item_id"] for item in response.json()["recommendations"]} == {
        "deftones-next",
        "radiohead-next",
    }


def test_music_recommendations_return_catalog_fallback_without_preferences():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-new",
            "limit": 2,
            "catalog": [
                {"item_id": "catalog-first"},
                {"item_id": "catalog-second"},
                {"item_id": "catalog-third"},
            ],
        },
    )

    assert response.status_code == 200
    assert response.json()["model"] == "music_catalog_fallback"
    assert [item["item_id"] for item in response.json()["recommendations"]] == [
        "catalog-first",
        "catalog-second",
    ]


def test_music_recommendations_return_empty_when_all_catalog_songs_are_liked():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-all-liked",
            "limit": 3,
            "interactions": [{"type": "like", "target_id": "already-liked"}],
            "catalog": [{"item_id": "already-liked", "profile": {"genres": ["metal"]}}],
        },
    )

    assert response.status_code == 200
    assert response.json()["recommendations"] == []


def test_music_recommendations_handle_liked_songs_without_genres():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": "auth-user-no-genre",
            "limit": 1,
            "interactions": [{"type": "like", "target_id": "liked-no-genre"}],
            "catalog": [
                {"item_id": "liked-no-genre", "profile": {"artists": ["Deftones"]}},
                {"item_id": "same-artist", "profile": {"artists": ["Deftones"]}},
                {"item_id": "unrelated", "profile": {"artists": ["Dua Lipa"]}},
            ],
        },
    )

    assert response.status_code == 200
    assert [item["item_id"] for item in response.json()["recommendations"]] == ["same-artist"]
