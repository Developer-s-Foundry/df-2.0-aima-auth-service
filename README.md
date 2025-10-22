## 🔒 AIMA Auth Service

The **AIMA Auth Service** is a Go-based microservice designed to handle user registration and login functionality, integrating with **PostgreSQL** for database operations and **RabbitMQ** for asynchronous tasks, such as email notifications.

-----

## 🚀 Getting Started

### Prerequisites

  * Go (version 1.18 or higher)
  * PostgreSQL database instance
  * RabbitMQ instance

### Environment Variables

The service relies on several environment variables, typically loaded from a `.env` file using `godotenv`.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `PORT` | The port the HTTP server will listen on. | `8080` |
| `RABBITMQ_URL` | Connection string for RabbitMQ. | `amqp://user:pass@host:5672/` |
| `DB_URL` | PostgreSQL connection URL (can be partially overridden by other `DB_*` vars). | `postgres://host:port` |
| `DB_USER` | PostgreSQL user. | `myuser` |
| `DB_HOST` | PostgreSQL host. | `localhost` |
| `DB_PASSWORD` | PostgreSQL password. | `mypassword` |
| `DB_PORT` | PostgreSQL port. | `5432` |
| `DB_NAME` | PostgreSQL database name. | `auth_db` |
| `DB_SSL` | PostgreSQL SSL mode (e.g., `disable`, `require`). | `disable` |

### Installation and Run

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd <repository-name>
    ```
2.  **Install dependencies:**
    ```bash
    go mod download
    ```
3.  **Run the service:**
    ```bash
    go run main.go
    ```

-----

## 💻 API Endpoints

The service uses `julienschmidt/httprouter` and provides the following routes:

| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/auth/register` | Handles new user registration. |
| `POST` | `/auth/login` | Handles user authentication and login. |

-----

## ⚙️ Key Features

  * **Database Connection:** Uses `postgres.ConnectPostgres` to establish a connection to **PostgreSQL**.
  * **Message Queuing:** Integrates with **RabbitMQ** for asynchronous communication, utilizing a dedicated consumer for `notification_email_queue` messages.
  * **Graceful Shutdown:** Implements robust graceful shutdown logic for the HTTP server and **RabbitMQ** connection upon receiving `SIGINT` or `SIGTERM` signals, ensuring clean exit for all goroutines.
  * **Configuration:** Loads configuration from environment variables via `godotenv`.

## API Example Responses

These examples illustrate the expected JSON structures returned by the AIMA Auth Service for successful and failure cases on the /auth/register and /auth/login endpoints.
1. **POST /auth/register Request Body**

```json
{
  "email": "user@example.com",
  "password": "strongpassword123"
}
```

Success Response (Status: 201 Created)
A new user is created in the PostgreSQL database, and a JWT session token is returned. RabbitMQ notifications are published asynchronously.
```json
{
  "email": "user@example.com",
  "message": "User created successfully",
  "status_code": 201,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYjA5YzVhNGUtNjI2NC00YTc1LTgxZTctMWY3YzRjZjM4Y2FkIiwiZXhwIjoxNjUwMDAwMDAwMH0.S-i-n4oEwX_Jt1z0X6j0sZ9KzQ8Bw4R2lF7V4GgVw",
  "userId": "b09c5a4e-6264-4a75-81e7-1f7c4cf38cad"
}
```

Failure Response: User Exists (Status: 409 Conflict)
"User already exists"

2. **POST /auth/login Request Body**

```json
{
  "email": "user@example.com",
  "password": "strongpassword123"
}
```

Success Response (Status: 200 OK)
The user is authenticated, and a new JWT session token is returned.

```json
{
  "id": "b09c5a4e-6264-4a75-81e7-1f7c4cf38cad",
  "email": "user@example.com",
  "message": "Login successful! Welcome to AIMAS",
  "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYjA5YzVhNGUtNjI2NC00YTc1LTgxZTctMWY3YzRjZjM4Y2FkIiwiZXhwIjoxNjUwMDAwMDAwMH0.S-i-n4oEwX_Jt1z0X6j0sZ9KzQ8Bw4R2lF7V4GgVw",
  "status_code": 200
}
```

Failure Response: Invalid Credentials (Status: 400 Bad Request)
If the email does not exist in the database.

```json
{
  "error": "email or password does not exists",
  "status": "Bad Request"
}
```

Failure Response: Unauthorized (Status: 401 Unauthorized)
If the email exists but the password is wrong.
Invalid login credentials

(Note: This is returned as a plain text string by http.Error in the handler.)
