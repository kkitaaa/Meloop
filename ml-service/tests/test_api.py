from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_health_endpoint_returns_ok():
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json()["status"] == "ok"
    assert response.json()["service"] == "ml-service"


def test_predict_returns_recommendations():
    response = client.post(
        "/predict",
        json={"user_id": 42, "limit": 2, "preferences": ["rock"]},
    )

    assert response.status_code == 200
    body = response.json()
    assert body["user_id"] == 42
    assert len(body["recommendations"]) == 2
    assert body["model"] == "baseline_recommender"


def test_predict_rejects_invalid_limit():
    response = client.post("/predict", json={"user_id": 42, "limit": 0})

    assert response.status_code == 422
