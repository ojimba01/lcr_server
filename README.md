# LCR Backend

This repository contains the backend code for the LCR_Server project.

## Description

`lcr_server` is a backend project aimed at providing a RESTful API for interacting with the LCR (Left, Center, Right) game. The project is built using Golang, Firebase SDK, Postgres, and other frameworks/libraries, ensuring a fast and efficient backend service that can handle concurrent requests.

## Project Structure

The backend project consists of the following main directories and files:

- `controllers`: This directory contains the controllers for the server.
- `db`: This directory contains the database related files.
- `docs`: This directory contains the compiled swagger documentation for the backend code.
- `lcr`: This directory contains the core logic of the LCR game.
- `model`: This directory contains the data models.
- `responses`: This directory contains response formatting.
- `routes`: This directory contains route definitions.
- `static`: This directory contains static files that the server might need to serve.
- `util`: This directory contains utility functions and structures.
- `backend`: This is the executeable produced from compiling the project.
- `go.mod` & `go.sum`: These files are used by Go's dependency management system.
- `main.go`: This is the main entry point for the GoFiber server.


## Local Development

To run this project locally, follow the steps below:

1. Make sure you have [Go](https://golang.org/dl/) installed on your machine.

2. Clone the repository:

```bash
git clone https://github.com/ojimba01/lcr_server.git
```

3. Navigate into the project directory:

```bash
cd lcr_server/backend
```
4. Build the project:

```bash
go build
```
5. Run the fiber server:

```bash
./backend
```

## Local Development with Docker

1. Make sure you have [Go](https://golang.org/dl/) installed on your machine.

2. Clone the repository:

```bash
git clone https://github.com/ojimba01/lcr_server.git
```

3. Navigate into the project directory:

```bash
cd lcr_server
```
4. Build the Docker image:

```bash
docker build -t lcr_server .
```
5. Run the Docker container:

```bash
docker run -p 3000:3000 lcr_server
```
