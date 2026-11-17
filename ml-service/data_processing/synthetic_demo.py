from __future__ import annotations

import argparse
from collections.abc import Sequence

import numpy as np
import pandas as pd

from app.models.recommendation_model import RecommendationModel
from data_processing.pipeline import UserPreferencesPipeline

_GROUPS = {
    "rock": {
        "genre": "rock",
        "artists": ("Arctic Monkeys", "Foo Fighters", "Radiohead", "The Strokes"),
        "songs": ("Do I Wanna Know?", "Everlong", "Creep", "Reptilia"),
    },
    "reggaeton": {
        "genre": "reggaeton",
        "artists": ("Bad Bunny", "Karol G", "Rauw Alejandro", "J Balvin"),
        "songs": ("Tití Me Preguntó", "TQG", "Todo de Ti", "Mi Gente"),
    },
}


def generate_synthetic_data(
    users_per_group: int = 4,
    seed: int = 42,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """Create reproducible user profiles and likes for two distinct music groups."""
    if users_per_group < 2:
        raise ValueError("Se necesitan al menos dos usuarios por grupo para comparar perfiles")

    rng = np.random.default_rng(seed)
    users: list[dict[str, object]] = []
    interactions: list[dict[str, object]] = []

    for group_index, (group, profile) in enumerate(_GROUPS.items()):
        group_user_offset = group_index * users_per_group
        group_track_offset = group_index * 1000
        for user_index in range(users_per_group):
            user_id = group_user_offset + user_index + 1
            liked_indices = rng.choice(len(profile["songs"]), size=2, replace=False)
            liked_songs = [profile["songs"][int(index)] for index in liked_indices]

            users.append(
                {
                    "user_id": user_id,
                    "group": group,
                    "genres": [profile["genre"]],
                    "artists": [
                        profile["artists"][0],
                        profile["artists"][user_index % (len(profile["artists"]) - 1) + 1],
                    ],
                    "songs": liked_songs,
                }
            )
            interactions.extend(
                {
                    "user_id": user_id,
                    "type": "like",
                    "target_id": group_track_offset + int(song_index) + 1,
                    "song": profile["songs"][int(song_index)],
                }
                for song_index in liked_indices
            )

    return pd.DataFrame(users), pd.DataFrame(interactions)


def score_user_pairs(users: pd.DataFrame) -> pd.DataFrame:
    """Score each unique user pair with the production preference pipeline and model."""
    pipeline = UserPreferencesPipeline()
    model = RecommendationModel()
    processed_users = [
        pipeline.transform(
            {
                "user_id": row.user_id,
                "genres": row.genres,
                "artists": row.artists,
                "songs": row.songs,
            }
        )
        for row in users.itertuples(index=False)
    ]
    feature_names = sorted(
        {feature for processed in processed_users for feature in processed.preference_features}
    )
    pair_scores: list[dict[str, object]] = []

    for first_index in range(len(users)):
        first = users.iloc[first_index]
        first_features = set(processed_users[first_index].preference_features)
        first_vector = np.fromiter(
            (feature in first_features for feature in feature_names),
            dtype=np.float32,
            count=len(feature_names),
        )
        for second_index in range(first_index + 1, len(users)):
            second = users.iloc[second_index]
            second_features = set(processed_users[second_index].preference_features)
            second_vector = np.fromiter(
                (feature in second_features for feature in feature_names),
                dtype=np.float32,
                count=len(feature_names),
            )
            result = model.compare_profiles(first_vector, second_vector, feature_names)
            pair_scores.append(
                {
                    "user_id_a": first["user_id"],
                    "user_id_b": second["user_id"],
                    "group_a": first["group"],
                    "group_b": second["group"],
                    "score": result.percentage,
                    "same_group": first["group"] == second["group"],
                }
            )

    return pd.DataFrame(pair_scores)


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Genera perfiles e interacciones sintéticas y evalúa su compatibilidad."
    )
    parser.add_argument("--seed", type=int, default=42, help="Semilla aleatoria (default: 42)")
    parser.add_argument(
        "--users-per-group",
        type=int,
        default=4,
        help="Usuarios de cada perfil musical (mínimo: 2; default: 4)",
    )
    args = parser.parse_args(argv)
    if args.users_per_group < 2:
        parser.error("--users-per-group debe ser al menos 2")

    users, interactions = generate_synthetic_data(args.users_per_group, args.seed)
    scores = score_user_pairs(users)
    same_group_mean = float(scores.loc[scores["same_group"], "score"].mean())
    cross_group_mean = float(scores.loc[~scores["same_group"], "score"].mean())

    print(f"Datos sintéticos (seed={args.seed})")
    print(f"Usuarios: {len(users)}; interacciones like: {len(interactions)}")
    print("\nPerfiles:")
    print(users.to_string(index=False))
    print("\nInteracciones:")
    print(interactions.to_string(index=False))
    print(f"\nCompatibilidad media intragrupo: {same_group_mean:.2f}%")
    print(f"Compatibilidad media entre grupos: {cross_group_mean:.2f}%")

    if same_group_mean <= cross_group_mean:
        print("RESULTADO: FALLÓ; la compatibilidad intragrupo no supera la intergrupo.")
        return 1

    print("RESULTADO: OK; el modelo detecta mayor compatibilidad dentro de cada grupo.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
