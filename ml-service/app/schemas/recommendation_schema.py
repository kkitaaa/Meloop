from pydantic import BaseModel, Field


class RecommendationRequest(BaseModel):
    user_id: int = Field(..., description="ID del usuario para generar recomendaciones.")
    limit: int = Field(default=10, ge=1, le=50, description="Cantidad máxima de sugerencias.")
    preferences: list[str] = Field(
        default_factory=list, description="Gustos o categorías del usuario."
    )


class RecommendationItem(BaseModel):
    item_id: int
    score: float
    reason: str


class RecommendationResponse(BaseModel):
    user_id: int
    recommendations: list[RecommendationItem]
    model: str
