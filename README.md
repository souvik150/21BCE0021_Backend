
# File Sharing & Management System (Backend)

**Deployed Link**: [https://trademarkia.souvik150.com](https://trademarkia.souvik150.com)

**API Documentation**: [https://zbimfk3mre.apidog.io](https://zbimfk3mre.apidog.io)

## Overview

This project is a backend system for a file-sharing and management platform. It provides functionalities like user registration, file upload, file sharing, file deletion, and real-time notifications through WebSocket. The system is deployed on AWS EC2 and integrates services like Redis, PostgreSQL, and AWS S3 for file storage.

---

## Table of Contents

- [File Sharing \& Management System (Backend)](#file-sharing--management-system-backend)
  - [Overview](#overview)
  - [Table of Contents](#table-of-contents)
  - [Deployment](#deployment)
  - [API Endpoints](#api-endpoints)
    - [Authentication](#authentication)
    - [File Management](#file-management)
    - [File Sharing](#file-sharing)
    - [WebSocket](#websocket)
  - [Local Setup](#local-setup)
    - [Prerequisites](#prerequisites)
    - [Steps](#steps)

---

## Deployment

The system is deployed on AWS EC2 with the following setup:

- **CI/CD Pipeline**: Integrated with GitHub Actions for continuous deployment. Each code push to the `main` branch triggers a CI pipeline, which builds the Docker image and pushes it to DockerHub. The CD pipeline runs on a self-hosted runner, pulling the latest image and deploying it to the server.
- **DockerHub Repository**: `souvik150/trademarkia`

---

## API Endpoints

### Authentication

1. **Register a new user**

   - **POST** `/register`
   - **Body**:
     ```json
     {
       "email": "example@example.com",
       "password": "password123"
     }
     ```

2. **Login user**

   - **POST** `/login`
   - **Body**:
     ```json
     {
       "email": "example@example.com",
       "password": "password123"
     }
     ```

3. **Get current user details**

   - **GET** `/me`
   - **Authorization**: Bearer token required.

---

### File Management

1. **Upload multiple files**

   - **POST** `/upload`
   - **Form Data**:
     - `files[]`: multiple files to be uploaded.

2. **Delete a file**

   - **DELETE** `/delete/:id`
   - **Path Parameter**:
     - `id`: The ID of the file to delete.

3. **Get all user files**

   - **GET** `/my-files`
   - **Query Parameters (optional)**:
     - `name`: filter by file name.
     - `type`: filter by file type.
     - `uploadDate`: filter by upload date in `YYYY-MM-DD` format.

4. **Get deleted files**

   - **GET** `/deleted-files`
   - **Authorization**: Bearer token required.

5. **Rename a file**

   - **PATCH** `/update`
   - **Query Parameters**:
     - `fileId`: The ID of the file to rename.
     - `newFileName`: New name for the file.

---

### File Sharing

1. **Generate a shareable link for a file**

   - **GET** `/generate/:id`
   - **Path Parameter**:
     - `id`: The ID of the file to generate a shareable link.

2. **Access a shared file**

   - **GET** `/share/:share_token`
   - **Path Parameter**:
     - `share_token`: The unique token for accessing the shared file.

---

### WebSocket

1. **Connect to WebSocket**

   - **GET** `/ws`
   - **Query Parameter**:
     - `token`: JWT token for the user.

---

## Local Setup

### Prerequisites

- Docker
- AWS CLI (for AWS services)
- PostgreSQL
- Redis

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/souvik150/file-sharing-app.git
   cd file-sharing-app
   ```

2. Create an `.env` file with the following variables:
   ```bash
   POSTGRES_URI=your_postgres_uri
   REDIS_URI=your_redis_uri
   AWS_ACCESS_KEY_ID=your_aws_access_key
   AWS_SECRET_ACCESS_KEY=your_aws_secret_key
   AWS_REGION=your_aws_region
   AWS_BUCKET_NAME=your_bucket_name
   ENCRYPTION_KEY=your_encryption_key
   BACKEND_URL=your_backend_url
   ```

3. Build and run using Docker:
   ```bash
   docker-compose up --build
   ```

4. The application will be running at `http://localhost:8080`.
