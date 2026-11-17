import pytest
from fastapi.testclient import TestClient

from app import main as main_module
from app.main import app

client = TestClient(app)


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
    initial = client.post("/predict", json={"user_id": 42, "limit": 1})
    updated = client.post(
        "/predict",
        json={
            "user_id": 42,
            "limit": 1,
            "interactions": [{"type": "like", "target_id": 7}],
        },
    )

    assert initial.status_code == 200
    assert updated.status_code == 200
    body = updated.json()
    assert set(body) == {
        "user_id",
        "recommendations",
        "model",
        "model_version",
        "interaction_count",
    }
    assert body["interaction_count"] == 1
    assert body["recommendations"] != initial.json()["recommendations"]
    assert set(body["recommendations"][0]) == {"item_id", "score", "reason"}


def test_predict_rejects_unknown_interaction_type():
    response = client.post(
        "/predict",
        json={
            "user_id": 42,
            "interactions": [{"type": "share", "target_id": 7}],
        },
    )

    assert response.status_code == 422


def test_friend_recommendations_rank_candidates_and_explain_matches():
    response = client.post(
        "/recommendations/friends",
        json={
            "user_id": 42,
            "limit": 2,
            "profile": {"genres": ["rock"], "artists": ["Arctic Monkeys"]},
            "candidate_profiles": [
                {"user_id": 7, "profile": {"genres": ["rock"]}},
                {"user_id": 8, "profile": {"genres": ["jazz"]}},
                {"user_id": 42, "profile": {"genres": ["rock"]}},
            ],
        },
    )

    assert response.status_code == 200
    body = response.json()
    assert set(body) == {
        "user_id",
        "recommendations",
        "model",
        "model_version",
        "interaction_count",
    }
    assert body["user_id"] == 42
    assert body["model"] == "profile_compatibility"
    assert body["model_version"] == "1.0.0"
    assert body["interaction_count"] == 0
    assert body["recommendations"] == [
        {"item_id": 7, "score": 0.7071, "reason": "Ambos escuchan a Rock"},
        {
            "item_id": 8,
            "score": 0.0,
            "reason": "No se encontraron preferencias musicales o sociales en común",
        },
    ]


def test_friend_recommendations_return_empty_for_new_user():
    response = client.post(
        "/recommendations/friends",
        json={"user_id": 42, "candidate_profiles": [{"user_id": 7}]},
    )

    assert response.status_code == 200
    assert response.json()["recommendations"] == []


def test_music_recommendations_rank_catalog_with_pipeline_and_model():
    response = client.post(
        "/recommendations/music",
        json={
            "user_id": 42,
            "limit": 2,
            "profile": {"genres": ["rock"], "artists": ["Arctic Monkeys"]},
            "catalog": [
                {"item_id": 101, "profile": {"genres": ["rock"], "artists": ["Arctic Monkeys"]}},
                {"item_id": 102, "profile": {"genres": ["rock"]}},
                {"item_id": 103, "profile": {"genres": ["jazz"]}},
            ],
        },
    )

    assert response.status_code == 200
    body = response.json()
    assert set(body) == {
        "user_id",
        "recommendations",
        "model",
        "model_version",
        "interaction_count",
    }
    assert body["model"] == "music_content_similarity"
    assert body["model_version"] == "1.0.0"
    assert [item["item_id"] for item in body["recommendations"]] == [101, 102]
    assert body["recommendations"][0]["score"] == 1.0
    assert set(body["recommendations"][0]) == {"item_id", "score", "reason"}


def test_music_recommendations_require_catalog():
    response = client.post(
        "/recommendations/music",
        json={"user_id": 42, "profile": {"genres": ["rock"]}},
    )

    assert response.status_code == 422


@pytest.mark.parametrize(
    ("path", "payload"),
    [
        ("/predict", {"user_id": 0}),
        (
            "/recommendations/friends",
            {"user_id": 42, "candidate_profiles": [{"user_id": 0}]},
        ),
        (
            "/recommendations/music",
            {"user_id": 42, "catalog": [{"item_id": 0}]},
        ),
        ("/recommendations/music", {"user_id": 42, "catalog": []}),
    ],
)
def test_recommendation_endpoints_reject_invalid_payloads(path, payload):
    response = client.post(path, json=payload)

    assert response.status_code == 422
    assert "detail" in response.json()
