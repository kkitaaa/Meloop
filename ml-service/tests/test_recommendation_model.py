import pytest
from scipy.sparse import csr_matrix

from app.models.recommendation_model import RecommendationModel


@pytest.fixture
def model():
    return RecommendationModel()


def test_compare_profiles_returns_full_compatibility_and_exact_reason(model):
    result = model.compare_profiles(
        [1, 1, 0],
        [1, 1, 0],
        ["artist:arctic monkeys", "genre:rock", "artist:daft punk"],
    )

    assert result.percentage == 100.0
    assert result.matching_features == ("artist:arctic monkeys", "genre:rock")
    assert result.reason == "Ambos escuchan a Arctic Monkeys, Rock"


def test_compare_profiles_converts_cosine_similarity_to_percentage(model):
    result = model.compare_profiles(
        [1, 1, 0, 0],
        [1, 0, 1, 0],
        ["artist:a", "artist:b", "genre:c", "friend:4"],
    )

    assert result.percentage == 50.0
    assert result.matching_features == ("artist:a",)
    assert result.reason == "Ambos escuchan a A"


def test_compare_profiles_rejects_different_vector_dimensions(model):
    with pytest.raises(ValueError, match="misma dimensión"):
        model.compare_profiles([1, 0], [1])


def test_compare_profiles_accepts_sparse_vectors(model):
    result = model.compare_profiles(
        csr_matrix([[1, 1, 0]]),
        csr_matrix([[1, 0, 1]]),
        ("artist:arctic monkeys", "genre:rock", "genre:jazz"),
    )

    assert result.percentage == 50.0
    assert result.matching_features == ("artist:arctic monkeys",)


def test_recommend_catalog_ranks_music_by_content_similarity(model):
    recommendations = model.recommend_catalog(
        csr_matrix([[1, 1, 0]]),
        csr_matrix([[1, 1, 0], [1, 0, 1], [0, 0, 1]]),
        [101, 102, 103],
        ["artist:a", "genre:rock", "genre:jazz"],
        limit=2,
    )

    assert [(item.item_id, item.score) for item in recommendations] == [(101, 1.0), (102, 0.5)]
    assert recommendations[0].matching_features == ("artist:a", "genre:rock")


def test_recommend_catalog_allows_injecting_collaborative_scores():
    model = RecommendationModel(collaborative_scorer=lambda _user_vector, _item_ids: [0.0, 1.0])

    recommendations = model.recommend_catalog([1, 0], [[1, 0], [1, 0]], [10, 20])

    assert [item.item_id for item in recommendations] == [20, 10]
    assert [item.score for item in recommendations] == [1.0, 0.5]
