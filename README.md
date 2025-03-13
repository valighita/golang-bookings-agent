
# Bookings AI Chat

This project demonstrates the implementation of an AI appointments agent for a dental clinic.
It showcases how to build interactive agents in Golang using `langchaingo`.

The application can be executed as an HTTP server, which serves a single-page frontend containing a chat box and connects to a web socket for the agent conversation. Alternatively, it can run in CLI mode to interact in the terminal.

The application currently stores the data only in memory, so the appointments are not persistent accross restarts.
It can be easily extended to store the data in a database, or even use a separate backend API.

## Running the Project

### Setup env vars

Set the following vars in `.env` or directly as env variables.

```
OPENAI_API_KEY=<token>
DEBUG_MODE=false
HTTP_SERVER_PORT=5001
HTTP_SERVER_USERNAME=user
HTTP_SERVER_PASSWORD=password
```

`HTTP_SERVER_USERNAME` and `HTTP_SERVER_PASSWORD` are optional. If specified, the http server asks for authentication when accessed.

### HTTP Server Mode

To run the project as an HTTP server:

```sh
make run
```

The server will start and serve a single-page frontend containing a chat box.
The chat box connects to a web socket for the agent conversation.

### CLI Mode

To run the project in CLI mode:

```sh
make run-cli
```

You can interact with the appointment agent directly in the terminal.

## Employees and Services

Employees and services available for the appointments are defined in `main.go`.
You can modify this file for custom data.
