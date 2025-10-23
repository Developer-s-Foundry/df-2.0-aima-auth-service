## 🔒 AIMA Auth Service

The **AIMA Auth Service** is a Go-based microservice designed to handle user registration and login functionality, integrating with **PostgreSQL** for database operations and **RabbitMQ** for asynchronous tasks, such as email notifications.

To access a deployed version of this app on Render, visit: [here](https://df-2-0-aima-auth-service.onrender.com)
To access the app through the API Gateway of the aimas project, use: [here](https://unlikely-cathrin-ubermensch-c0536783.koyeb.app/auth)

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
| `POST` | `/register` | Handles new user registration. |
| `POST` | `/login` | Handles user authentication and login. |

-----

## ⚙️ Key Features

  * **Database Connection:** Uses `postgres.ConnectPostgres` to establish a connection to **PostgreSQL**.
  * **Message Queuing:** Integrates with **RabbitMQ** for asynchronous communication, utilizing a dedicated consumer for `notification_email_queue` messages.
  * **Graceful Shutdown:** Implements robust graceful shutdown logic for the HTTP server and **RabbitMQ** connection upon receiving `SIGINT` or `SIGTERM` signals, ensuring clean exit for all goroutines.
  * **Configuration:** Loads configuration from environment variables via `godotenv`.

## API Example Responses

These examples illustrate the expected JSON structures returned by the AIMA Auth Service for successful and failure cases on the /auth/register and /auth/login endpoints.
1. **POST /register Request Body**

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

2. **POST /login Request Body**

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

# RabbitMQ Message Publishing Documentation

This documentation describes how the authentication service publishes messages to RabbitMQ, detailing which exchanges and queues are used and what other services can consume these messages.

---

## Overview

The system uses RabbitMQ as a communication layer between microservices.  
Messages are published to specific exchanges with defined routing keys so that subscribed services can consume them based on their responsibilities.

There are currently *two main exchanges*:

- **notification_exchange** — handles all notification-related events such as sending emails or alerts.
- **user_exchange** — handles events related to user data synchronization and management.

---

## Published Messages

### 1. Notification Exchange

Messages sent to this exchange are consumed by the *Notification Service*.  
They are routed through the **notification.queue**.

- *Exchange Name:* notification_exchange  
- *Routing Key / Queue:* notification.queue  
- *Purpose:* To deliver notification-related messages such as welcome emails.

Example message type published to this exchange:

- **auth_welcome_mail** — triggers a welcome email to new users after registration.

---

### 2. User Management Exchange

Messages sent to this exchange are consumed by the *User Management Service*.  
They are routed through the **user.queue**.

- *Exchange Name:* user_exchange  
- *Routing Key / Queue:* user.queue  
- *Purpose:* To synchronize or update user information across services.

Example message type published to this exchange:

- **auth_user_info** — used to send user details for synchronization between services.

---

## Message Structure

Each published message contains data in a key-value format, allowing consumers to easily identify and process events.

Typical message payload structure:

| Field | Description |
|--------|-------------|
| type | The event type that describes the purpose of the message (e.g., auth_welcome_mail). |
| email | The user’s email address associated with the event. |
| Id | The unique identifier of the user in the system. |

Example:
text
type: auth_welcome_mail  
email: user@example.com  
Id: 12345


---

## Event Constants

| Constant | Value | Description |
|-----------|--------|-------------|
| NotifyUserSuccessfulSignUp | auth_welcome_mail | Sent when a user successfully signs up. Triggers a welcome email notification. |
| AuthUser | auth_user_info | Used to share or update user information between services. |
| WelcomeEmailQueue | queue | Represents the bound queue name for welcome emails. |

---

## Consumption Guidelines

- The *Notification Service* should listen on the notification.queue queue bound to the notification_exchange exchange.
- The *User Management Service* should listen on the user.queue queue bound to the user_exchange exchange.
- Each consumer should inspect the type field of the message to determine the action to perform.
- Messages are JSON-encoded for consistent parsing across services.

---

## Summary Table

| Event Type | Exchange | Queue | Consuming Service | Purpose |
|-------------|-----------|--------|------------------|----------|
| auth_welcome_mail | notification_exchange | notification.queue | Notification Service | Sends a welcome email to new users. |
| auth_user_info | user_exchange | user.queue | User Management Service | Synchronizes user data across services. |
