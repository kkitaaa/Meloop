from app.schemas.recommendation_schema import (
    RecommendationItem,
    RecommendationRequest,
    RecommendationResponse,
)


class RecommendationService:
    def __init__(self) -> None:
        self.model_name = "baseline_recommender"

    def generate(self, request: RecommendationRequest) -> RecommendationResponse:
        recommendations: list[RecommendationItem] = []

        for index in range(1, min(request.limit, 5) + 1):
            recommendations.append(
                RecommendationItem(
                    item_id=index + request.user_id,
                    score=round(1.0 - (index * 0.1), 2),
                    reason="Sugerencia de prueba basada en preferencias del usuario",
                )
            )

        return RecommendationResponse(
            user_id=request.user_id,
            recommendations=recommendations,
            model=self.model_name,
        )
