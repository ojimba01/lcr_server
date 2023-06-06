# Build backend
FROM golang:1.20.4-bullseye AS backend

# Define a build argument for the Firebase credentials
ARG POSTGRES_PASSWORD
# Set the environment variable
ENV POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
# COPY backend/lcr_webapp.json .

RUN go build -o backend
EXPOSE 3000

CMD ["./backend"]

