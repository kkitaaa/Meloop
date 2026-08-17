from fastapi import FastAPI

from app.schemas.recommendation_schema import RecommendationRequest, RecommendationResponse
from app.services.recommendation_service import RecommendationService

app = FastAPI(
    title="Meloop ML Service",
    description="Servicio de machine learning para recomendaciones y análisis de contenido.",
    version="0.1.0",
)

recommendation_service = RecommendationService()


@app.get("/health")
def health_check():
    return {
        "status": "ok",
        "service": "ml-service",
        "message": "ML service is running",
    }


@app.get("/")
def root():
    return {
        "message": "Bienvenido al servicio de ML de Meloop",
        "docs": "/docs",
    }


@app.post("/predict", response_model=RecommendationResponse)
def predict_demo(payload: RecommendationRequest):
    return recommendation_service.generate(payload)
