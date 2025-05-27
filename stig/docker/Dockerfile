# Create CUDA GPU image.
# FROM pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime

# Create CPU-only image.
FROM python:3.13-slim

# Avoid apt prompts & speed up builds
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        git gcc build-essential libgl1 && \
    rm -rf /var/lib/apt/lists/*

# Setup working directory.
WORKDIR /app

# Source code is included as a volume.

# Install dependencies.
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
