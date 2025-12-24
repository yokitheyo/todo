# Todo HTTP Server in Go

A simple HTTP server in Go for managing tasks (Todos).  

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/todos` | Create a new task |
| `GET` | `/todos` | Get all tasks |
| `GET` | `/todos/{id}` | Get a task by ID |
| `PUT` | `/todos/{id}` | Update a task by ID |
| `DELETE` | `/todos/{id}` | Delete a task by ID |

## Features

In-memory storage (no external database) 

Validation: title cannot be empty (returns 400 Bad Request)

Error handling: returns 404 Not Found if task is not found

Optional: request logging and context-based timeouts

## Running the server 

``` bash 
git clone https://github.com/yokitheyo/todo.git

cd todo

go run cmd/main.go 
```

## Testing

You can test the API via todo.rest or using curl.

```bash
make all
```

