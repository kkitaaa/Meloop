from typing import Literal

from pydantic import BaseModel, Field


class Interaction(BaseModel):
    type: Literal["like", "friend_added", "comment", "post_interaction"]
    target_id: int = Field(..., ge=1)


class UserProfile(BaseModel):
    genres: list[str] = Field(default_factory=list)
    artists: list[str] = Field(default_factory=list)
    songs: list[str] = Field(default_factory=list)


class CandidateProfile(BaseModel):
    user_id: int = Field(..., ge=1)
    profile: UserProfile = Field(default_factory=UserProfile)


class RecommendationRequest(BaseModel):
    user_id: int = Field(..., ge=1, description="ID del usuario para generar recomendaciones.")
    limit: int = Field(default=10, ge=1, le=50, description="Cantidad máxima de sugerencias.")
    type: Literal["all", "music", "friends"] = "all"
    preferences: list[str] = Field(
        default_factory=list, description="Gustos o categorías del usuario."
    )
    profile: UserProfile = Field(default_factory=UserProfile)
    candidate_profiles: list[CandidateProfile] = Field(
        default_factory=list,
        description="Perfiles candidatos que el modelo debe ordenar.",
    )
    interactions: list[Interaction] = Field(
        default_factory=list,
        description="Interacciones recientes que pueden cambiar las sugerencias.",
    )


class RecommendationItem(BaseModel):
    item_id: int
    score: float
    reason: str


class RecommendationResponse(BaseModel):
    user_id: int
    recommendations: list[RecommendationItem]
    model: str
    interaction_count: int
