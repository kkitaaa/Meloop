from typing import List, Literal

from pydantic import BaseModel, Field


class Interaction(BaseModel):
    type: Literal["like", "friend_added"]
    target_id: int


class RecommendationRequest(BaseModel):
    user_id: int = Field(..., description="ID del usuario para generar recomendaciones.")
    limit: int = Field(default=10, ge=1, le=50, description="Cantidad máxima de sugerencias.")
    preferences: List[str] = Field(default_factory=list, description="Gustos o categorías del usuario.")
    interactions: List[Interaction] = Field(
        default_factory=list,
        description="Interacciones recientes que pueden cambiar las sugerencias.",
    )


class RecommendationItem(BaseModel):
    item_id: int
    score: float
    reason: str


class RecommendationResponse(BaseModel):
    user_id: int
    recommendations: List[RecommendationItem]
    model: str
    interaction_count: int
