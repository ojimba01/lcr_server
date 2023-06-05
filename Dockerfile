# Build backend
FROM golang:1.20.4-bullseye AS backend

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
# COPY backend/lcr_webapp.json .

RUN go build -o backend
EXPOSE 3000

CMD ["./backend"]

