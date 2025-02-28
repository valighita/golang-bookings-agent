FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates

COPY .env /app/.env
COPY bookings-ai-chat /app/bookings-ai-chat
COPY frontend/index.html /app/frontend/index.html

CMD ["/app/bookings-ai-chat"]
