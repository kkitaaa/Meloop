from __future__ import annotations

import json
from collections.abc import Mapping
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit
from urllib.request import Request, urlopen


def fetch_preferences(
    endpoint: str,
    user_id: int,
    *,
    timeout: float = 10.0,
) -> Mapping[str, Any]:
    """Fetch one user's raw preference payload from a Go HTTP endpoint."""
    parts = urlsplit(endpoint)
    query = dict(parse_qsl(parts.query))
    query["user_id"] = str(user_id)
    request_url = urlunsplit(
        (parts.scheme, parts.netloc, parts.path, urlencode(query), parts.fragment)
    )
    request = Request(request_url, headers={"Accept": "application/json"})

    try:
        with urlopen(request, timeout=timeout) as response:
            payload = json.load(response)
    except (HTTPError, URLError, TimeoutError) as error:
        raise RuntimeError(f"No se pudieron extraer preferencias desde {endpoint}") from error
    except json.JSONDecodeError as error:
        raise ValueError("El endpoint de preferencias no devolvió JSON válido") from error

    if not isinstance(payload, Mapping):
        raise ValueError("El endpoint de preferencias debe devolver un objeto JSON")
    return payload
