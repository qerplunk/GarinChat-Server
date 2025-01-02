# GarinChat server

> [!IMPORTANT]
> You will have to implement a third party user authentication service for your frontend that is able to generate JWT tokens.

## What it is

A [golang](https://go.dev/)-based WebSocket server (using [gorilla/websocket](https://github.com/gorilla/websocket)) for chat room messaging.

- [Server features](#server-resource-friendly-features)
- [How to run](#how-to-run)
- [Server API](#server-api-for-client-implementation)
- [Work in Progress features](#wip-features)

## Server resource-friendly features

> Helps avoid users from using scripts to spam your server.
>
> You can configure or remove these if desired.

- JWT authentication through WebSocket messages. Safer and faster than passing the token as a URL parameter.
- Closes connections if the API is not used correctly, such as if a user sends a message without authenticating or joining a room first.
- Rate limits the amount of messages users can send per set window of time (default is 1 message per 1 second and will close the connection if exceeded).
- Closes user connections if they have not authenticated or joined a chat room within a period of time after a connection is first established (default is 2 seconds to authenticate and another 2 seconds to join a room).
- Closes user connections if they have not sent a message within a period of time (WIP).

## How to run

### Required environment variables in `.env`:

- `ALLOWED_ORIGINS`: Comma-separated list of origins allowed to create connections (e.g., http://localhost:3000,https://www.google.com)
- `JWT_DECODE_SECRET`: JWT decode string
- `PORT`: Port to run

### Three ways to run

- **Docker compose**

  > Edit `docker-compose.yml` to change `.env` file or Docker port number

  `docker compose up --build`

- **Dockerfile**  
  `docker build -t garinchat-image .`  
  `docker run --env-file=.env -p 8080:8080 --name garinchat-container garinchat-image`

- **Build and run using the go cli**  
   `go mod download`  
   `go build .`  
   `./garin-chat`  
   <em>or</em>  
   `go mod download`  
   `go run .`

## Server API for client implementation

- **Establish a connection with your backend**

  ```javascript
  const ws = new WebSocket("wss://www.example.com");
  ```

- **Send JWT authentication token**

  The server will close the connection with the user if an `auth` message is not sent within 2 seconds of connecting.

  ```javascript
  ws.send(
    JSON.stringify({
      type: "auth",
      body: token,
    }),
  );
  ```

- **Join a room**

  The server will close the connection with the user if a `join` message is not sent within 2 seconds of authenticating.  
  By default, the username and room have to be strings length of 3 or greater.

  ```javascript
  ws.send(
    JSON.stringify({
      type: "join",
      username: username,
      room: room,
    }),
  );
  ```

- **Send a message to everyone in the same room**

  The server will not send the message back to the user that sent it, make sure to append the message to the user's local chat message list.  
  Messages of length 0 will remove the user from the room and cause the connection to terminate.

  ```javascript
  ws.send(
    JSON.stringify({
      type: "message",
      body: userInputMessage,
    }),
  );
  ```

- **Leave a room**

  The server will remove the user's connection from the room the user is currently in. If no users are left in that room the server will remove the room from the active rooms map.

  ```javascript
  ws.send(
    JSON.stringify({
      type: "userleave",
    }),
  );
  ```

# WIP features

- Better server settings for timeout lengths, rate limiting, etc.
- Time out users if they have not sent messages
- A ban list for users who repeatedly get kicked by the rate limiter, could use sqlite for this.
