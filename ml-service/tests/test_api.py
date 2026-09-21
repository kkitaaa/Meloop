from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


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
    assert updated.json()["interaction_count"] == 1
    assert updated.json()["recommendations"] != initial.json()["recommendations"]


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
    assert response.json()["recommendations"] == [
        {"item_id": 7, "score": 0.7071, "reason": "Ambos escuchan a Rock"},
        {"item_id": 8, "score": 0.0, "reason": "No se encontraron preferencias musicales o sociales en común"},
    ]


def test_friend_recommendations_return_empty_for_new_user():
    response = client.post(
        "/recommendations/friends",
        json={"user_id": 42, "candidate_profiles": [{"user_id": 7}]},
    )

    assert response.status_code == 200
    assert response.json()["recommendations"] == []
