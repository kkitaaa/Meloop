from typing import List

from app.schemas.recommendation_schema import (
    RecommendationItem,
    RecommendationRequest,
    RecommendationResponse,
)


class RecommendationService:
    def __init__(self) -> None:
        self.model_name = "baseline_recommender"

    def generate(self, request: RecommendationRequest) -> RecommendationResponse:
        recommendations: List[RecommendationItem] = []
        interaction_count = len(request.interactions)
        liked_items = {
            interaction.target_id
            for interaction in request.interactions
            if interaction.type == "like"
        }
        recommendation_offset = interaction_count * 10

        for index in range(1, min(request.limit, 5) + 1):
            item_id = request.user_id + recommendation_offset + index
            if item_id in liked_items:
                item_id += 1000
            recommendations.append(
                RecommendationItem(
                    item_id=item_id,
                    score=round(max(0.1, 1.0 - (index * 0.1) + interaction_count * 0.01), 2),
                    reason=(
                        "Sugerencia adaptada a la actividad reciente del usuario"
                        if interaction_count
                        else "Sugerencia de prueba basada en preferencias del usuario"
                    ),
                )
            )

        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=recommendations,
            model=self.model_name,
            interaction_count=interaction_count,
        )
