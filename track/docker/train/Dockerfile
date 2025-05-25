# CUDA GPU build (comment out if CPU-only is fine)
# FROM pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime
# CPU-only (smaller, safe everywhere):
FROM python:3.12-slim

# Avoid apt prompts & speed up builds
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        git gcc build-essential libgl1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy requirements first → leverage Docker layer cache
COPY track/requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

# Copy the actual trainer code last
COPY track/ .

# Entry point = python -m model.train so `docker run ... --epochs 5` works
ENTRYPOINT ["python", "-m", "model.train"]
