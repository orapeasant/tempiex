# Tempiex Python SDK — base image for worker deployments
FROM python:3.13-slim

WORKDIR /app

# Install SDK
COPY sdk-python/pyproject.toml sdk-python/uv.lock* ./
COPY sdk-python/tempiex/ ./tempiex/

RUN pip install --no-cache-dir grpcio>=1.72.0 protobuf>=5.29.0

# Default: run the sample app (override CMD in derived images)
COPY sdk-python/samples/ ./samples/

CMD ["python", "samples/order_workflow.py"]
