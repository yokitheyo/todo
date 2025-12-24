HOST := http://localhost:8080

.PHONY: all health create_todo create_todo_empty create_todo_unknown get_todos get_todo_1 get_todo_999 update_todo update_todo_empty delete_todo

all: health create_todo create_todo_empty create_todo_unknown get_todos get_todo_1 get_todo_999 update_todo update_todo_empty delete_todo

health:
	curl -s -X GET $(HOST)/health -H "Accept: application/json" | jq

create_todo:
	curl -s -X POST $(HOST)/todos \
	-H "Content-Type: application/json" \
	-d '{"title": "Buy groceries", "description": "Milk, eggs, bread", "completed": false}' | jq

create_todo_empty:
	curl -s -X POST $(HOST)/todos \
	-H "Content-Type: application/json" \
	-d '{"title": "", "description": "No title", "completed": false}' | jq

create_todo_unknown:
	curl -s -X POST $(HOST)/todos \
	-H "Content-Type: application/json" \
	-d '{"title": "Test", "unknown": "field", "completed": false}' | jq

get_todos:
	curl -s -X GET $(HOST)/todos | jq

get_todo_1:
	curl -s -X GET $(HOST)/todos/1 | jq

get_todo_999:
	curl -s -X GET $(HOST)/todos/999 | jq

update_todo:
	curl -s -X PUT $(HOST)/todos/1 \
	-H "Content-Type: application/json" \
	-d '{"description": "Milk, eggs, bread, butter"}' | jq

update_todo_empty:
	curl -s -X PUT $(HOST)/todos/1 \
	-H "Content-Type: application/json" \
	-d '{"title": ""}' | jq

delete_todo:
	curl -s -X DELETE $(HOST)/todos/1 | jq
