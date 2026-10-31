FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt ./
RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir -r requirements.txt \
    && useradd --system --uid 10001 --create-home app

COPY --chown=app:app app ./app
COPY --chown=app:app data_processing ./data_processing

EXPOSE 8001

USER app

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8001"]
