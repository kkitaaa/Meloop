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
