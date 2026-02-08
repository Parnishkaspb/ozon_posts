# Ozon Posts

GraphQL + gRPC сервис для постов и иерархических комментариев.

## Что реализовано
- Посты: создание, чтение одного поста, чтение списка с cursor pagination.
- Комментарии: неограниченная вложенность, ограничение длины текста, pagination по `postId` и `parentId`.
- GraphQL Subscriptions: `commentAdded(postId: ID!)` для асинхронной доставки новых комментариев.
- Два backend-хранилища:
  - `postgres`
  - `memory`

## Архитектура
- `service/`: gRPC backend (бизнес-логика, репозитории, миграции).
- `graphql/`: gqlgen API gateway, dataloader, subscription hub.
- `proto/`: protobuf контракты.

## Конфигурация хранилища
Файл: `/ozon/service/config/config.yaml`

```yaml
storage:
  driver: "postgres" # или "memory"
```

## Локальный запуск
```bash
docker compose up --build
```

или через `Makefile` в корне:
```bash
make build
```

- GraphQL Playground: [http://localhost:8080](http://localhost:8080)
- gRPC service: `localhost:9090`

## Тесты
```bash
cd service && go test ./...
cd graphql && go test ./...
```

## Makefile
### Корневой `/ozon/Makefile`
- `make build`:
  - выполняет `docker compose build && docker-compose up -d`
- `make protoc_generate`:
  - генерирует gRPC/Go-код из `/ozon/proto/service/v1/service.proto`

### GraphQL `/ozon/graphql/Makefile`
- `make generate`:
  - запускает `gqlgen generate`

## Границы задания и доп. решения
- Намеренно не добавлялись отдельные сценарии, которые не требовались в задаче как обязательные:
  - публичная регистрация пользователя (`createUser` в GraphQL и полный user-flow);
  - расширенная бизнес-валидация JWT за пределами базовой проверки в middleware.
- При этом для демонстрации практических навыков и удобства использования API добавлены:
  - `login` и JWT middleware;
  - извлечение пользователя из токена в контекст, чтобы не передавать `authorId` вручную при каждом `createPost`/`createComment`.

## Subscription smoke-check
1. В Playground открыть подписку:
```graphql
subscription {
  commentAdded(postId: "<POST_ID>") { id text parentId createdAt }
}
```
2. Во второй вкладке создать комментарий:
```graphql
mutation {
  createComment(postId: "<POST_ID>", text: "hello") { id }
}
```
3. Событие должно сразу прийти в подписку.

## Быстрые GraphQL запросы
### 1) Логин и получение JWT
```graphql
mutation {
  login(login: "Ivan", password: "MoscowNeverSleep") {
    token
  }
}
```

### 2) Создать пост
```graphql
mutation {
  createPost(text: "Первый пост", withoutComment: false) {
    id
    text
    withoutComment
    createdAt
    author {
      id
      name
    }
  }
}
```

### 3) Список постов + корневые комментарии + replies
```graphql
query {
  posts(first: 10) {
    edges {
      node {
        id
        text
        createdAt
        author {
          id
          name
        }
        comments(first: 10) {
          edges {
            node {
              id
              text
              parentId
              replies(first: 10) {
                edges {
                  node {
                    id
                    text
                    parentId
                  }
                }
                pageInfo {
                  endCursor
                  hasNextPage
                }
              }
            }
          }
          pageInfo {
            endCursor
            hasNextPage
          }
        }
      }
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}
```

### 4) Создать комментарий к посту
```graphql
mutation {
  createComment(postId: "<POST_ID>", text: "Корневой комментарий") {
    id
    postId
    parentId
    text
    createdAt
  }
}
```

### 5) Создать ответ на комментарий
```graphql
mutation {
  createComment(
    postId: "<POST_ID>"
    parentId: "<PARENT_COMMENT_ID>"
    text: "Ответ на комментарий"
  ) {
    id
    postId
    parentId
    text
    createdAt
  }
}
```
