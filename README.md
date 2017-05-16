# Go kit: stringsvc2 with JWT

[Go kit](https://gokit.io)  
[stringsvc2](https://gokit.io/examples/stringsvc.html#middlewares)  

As part of learning Go Kit, now I'm trying to add JWT to stringsvc2.

- Use custom claims to save client ID in JWT
- Add ```/auth``` endpoint for getting a token
- Protect ```/uppercase``` and ```/count``` endpoints with JWT
- Use Context to pass client ID
- Log client ID and record client ID in metrics
- Protect ```/metrics``` endpoint with basic auth

```
go build -o server
./server
```

```
curl -v -XPOST -d '{"clientId": "mobile", "clientSecret": "m_secret"}' http://localhost:8080/auth
```

```
curl -v -XPOST -d '{"s": "hello world"}' -H "Authorization: Bearer eyJhbGci..." http://localhost:8080/uppercase
```

```
curl -v --user prometheus:password http://localhost:8080/metrics
```