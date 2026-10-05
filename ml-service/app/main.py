import time

from fastapi import FastAPI, Request, Response
from fastapi.responses import JSONResponse

from app.logging_config import configure_logging
from app.schemas.recommendation_schema import (
    HealthResponse,
    MusicRecommendationRequest,
    RecommendationRequest,
    RecommendationResponse,
    ReadinessResponse,
)
from app.services.recommendation_service import RecommendationService

logger = configure_logging()
app = FastAPI(
    title="Meloop ML Service",
    description="Servicio de machine learning para recomendaciones y análisis de contenido.",
    version="0.1.0",
)

recommendation_service = RecommendationService()


@app.middleware("http")
async def request_logging(request: Request, call_next):
    started = time.perf_counter()
    try:
        response = await call_next(request)
    except Exception:
        logger.exception("request_failed method=%s path=%s", request.method, request.url.path)
        return JSONResponse(status_code=500, content={"detail": "Internal server error"})

    duration_ms = round((time.perf_counter() - started) * 1000, 2)
    logger.info(
        "http_request method=%s path=%s status=%s duration_ms=%s",
        request.method,
        request.url.path,
        response.status_code,
        duration_ms,
    )
    return response


@app.on_event("startup")
async def startup_event():
    logger.info("service_started")


@app.on_event("shutdown")
async def shutdown_event():
    logger.info("service_stopped")


@app.get("/health", response_model=HealthResponse)
def health_check():
    return {
        "status": "ok",
        "service": "ml-service",
    }


@app.get("/ready", response_model=ReadinessResponse)
def readiness_check(response: Response):
    model_loaded = recommendation_service.model is not None
    data_pipeline_loaded = recommendation_service.preferences_pipeline is not None
    ready = recommendation_service.is_ready
    response.status_code = 200 if ready else 503
    return {
        "status": "ready" if ready else "not_ready",
        "service": "ml-service",
        "model_loaded": model_loaded,
        "data_pipeline_loaded": data_pipeline_loaded,
        "model_version": recommendation_service.model.version if model_loaded else None,
    }


@app.get("/")
def root():
    return {
        "message": "Bienvenido al servicio de ML de Meloop",
        "docs": "/docs",
    }


@app.post("/predict", response_model=RecommendationResponse)
def predict_demo(payload: RecommendationRequest):
    logger.info(
        "recommendation_requested user_id=%s limit=%s interactions=%s",
        payload.user_id,
        payload.limit,
        len(payload.interactions),
    )
    return recommendation_service.generate(payload)


@app.post("/recommendations/friends", response_model=RecommendationResponse)
def recommend_friends(payload: RecommendationRequest):
    """Return the best compatible friend candidates for a user profile."""
    logger.info(
        "friend_recommendation_requested user_id=%s limit=%s candidates=%s",
        payload.user_id,
        payload.limit,
        len(payload.candidate_profiles),
    )
    return recommendation_service.generate_friend_recommendations(payload)


@app.post("/recommendations/music", response_model=RecommendationResponse)
def recommend_music(payload: MusicRecommendationRequest):
    logger.info(
        "music_recommendation_requested user_id=%s limit=%s catalog=%s",
        payload.user_id,
        payload.limit,
        len(payload.catalog),
    )
    return recommendation_service.generate_music_recommendations(payload)
