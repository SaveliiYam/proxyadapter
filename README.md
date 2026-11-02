# proxyadapter

Библиотека для **proxy adapter** сервисов (`*_pa`): публичный HTTP (gin) без Basic на входе,
проброс на upstream API с service-auth, декларативные маршруты из YAML (`yamlroutes`).

Стек: [`ginserver`](https://github.com/SaveliiYam/ginserver), [`yamlroutes`](https://github.com/SaveliiYam/yamlroutes).

## Установка

```bash
go get github.com/SaveliiYam/proxyadapter@v0.1.0
```

## routes.yaml

```yaml
routes:
  - method: POST
    path: /api/v1/sign_in/password
    do: auth_client_proxy

  - method: POST
    path: /api/v1/user/info
    do: GetUserInfo
    auth: [jwt]
```

- `do: auth_client_proxy` — регистрируется автоматически, если передан `upstream.Client`.
- `auth: [jwt]` — Bearer access JWT через `jwt.AccessParser`.
- `auth: [basic]` на ingress по умолчанию **запрещён** (proxy-адаптер для клиентов).

## Пример (gin)

```go
up, _ := upstream.New(upstreamCfg)

parser := myJWTAccessParser{} // jwt.AccessParser

handlers := map[string]gin.HandlerFunc{
	"GetUserInfo": myGetUserInfo(up),
}

err := ginroutes.MountHandlersFile(r, "routes.yaml", handlers, ginroutes.MountConfig{
	Upstream:         up,
	AccessParser:     parser,
	InvalidTokenCode: constants.ErrInvalidToken,
	InternalErrCode:  constants.ErrInternal,
})
```

## Пакеты

| Пакет | Назначение |
|-------|------------|
| `upstream` | HTTP-клиент, `Forward` на upstream с Basic/Bearer |
| `jwt` | Интерфейс `AccessParser` |
| `ginroutes` | Proxy handler, JWT middleware, `MountHandlersFile` |

Go ≥ 1.27.
