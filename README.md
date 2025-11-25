Here’s your updated README section with the `curl` examples added to show the API key behavior:

```
# Public Library REST API

Using Go write a minimal web based REST API for a fictional public library that can perform the following functions: List all books in the library, perform all CRUD operations on a single book, store data in a database, create a Dockerfile for the Go application, and use Docker Compose to create a multi-container environment including the database and application.

A Dockerfile packages the Go application as a container.  
Docker Compose creates a multi-container setup including the API and MySQL database.

## Features

| Feature          | Description                       | Endpoint                |
| ---------------- | --------------------------------- | ----------------------- |
| List all books   | Retrieve all books in the library | `GET /v1/books`         |
| Create a book    | Add a new book                    | `POST /v1/books`        |
| Read a book      | Get a single book by ID           | `GET /v1/books/{id}`    |
| Update a book    | Modify an existing book           | `PUT /v1/books/{id}`    |
| Delete a book    | Remove a book                     | `DELETE /v1/books/{id}` |
| Database storage | Persist all book data             | MySQL                   |
| API key required | Must provide `X-API-KEY` header  | `All /v1/books*`        |

## Additional Capabilities

* ✔️ Rate limiting  
* ✔️ XSS filter (backend responsibility only)  
* ✔️ Data validation  
* ✔️ DDoS protection (backend-side safeguards only)  
* ✔️ Observability / Monitoring: Datadog SDK  

**Note:** The server enforces a static dummy API key `apitest`; all requests must include `X-API-KEY: apitest` header or they will be rejected.

### Example

Without API key:

```

curl -X GET [http://localhost:8080/v1/books](http://localhost:8080/v1/books)
unauthorized

```

With API key:

```

curl -X GET [http://localhost:8080/v1/books](http://localhost:8080/v1/books) -H "X-API-KEY: apitest"
internal error

```

This shows that the server detects the presence of the API key.  

## Spin up
```

make up

```

## Datadog Setup

To enable Datadog tracing, you need to set the following environment variables:

```

ENABLE_DATADOG=1
DD_AGENT_HOST=127.0.0.1
DD_TRACE_AGENT_PORT=8126
DD_ENV=production
DD_SERVICE=library-api

```