from dataclasses import dataclass


@dataclass
class RecommendationModel:
    name: str = "baseline_recommender"
    version: str = "0.1.0"
