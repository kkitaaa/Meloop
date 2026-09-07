import time

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from app.logging_config import configure_logging
from app.schemas.recommendation_schema import RecommendationRequest, RecommendationResponse
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
    logger.info(
        "recommendation_requested user_id=%s limit=%s interactions=%s",
        payload.user_id,
        payload.limit,
        len(payload.interactions),
    )
    return recommendation_service.generate(payload)
