import pandas as pd
import pytest

from data_processing.synthetic_demo import generate_synthetic_data, score_user_pairs


def test_synthetic_data_is_reproducible_and_includes_liked_songs():
    users, interactions = generate_synthetic_data(seed=12, users_per_group=3)
    repeated_users, repeated_interactions = generate_synthetic_data(seed=12, users_per_group=3)

    pd.testing.assert_frame_equal(users, repeated_users)
    pd.testing.assert_frame_equal(interactions, repeated_interactions)
    assert len(users) == 6
    assert len(interactions) == 12
    assert all(
        interaction.song in users.loc[users.user_id == interaction.user_id, "songs"].iloc[0]
        for interaction in interactions.itertuples()
    )


def test_model_scores_same_group_users_higher_than_cross_group_users():
    users, _ = generate_synthetic_data()
    pair_scores = score_user_pairs(users)
    same_group_mean = pair_scores.loc[pair_scores.same_group, "score"].mean()
    cross_group_mean = pair_scores.loc[~pair_scores.same_group, "score"].mean()

    assert same_group_mean > cross_group_mean
    assert cross_group_mean == 0


def test_data_generation_requires_two_users_per_group():
    with pytest.raises(ValueError, match="al menos dos usuarios"):
        generate_synthetic_data(users_per_group=1)
