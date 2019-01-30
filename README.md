# Golang REST Layer SQL Storage Handler

This [REST Layer](https://github.com/rs/rest-layer) resource storage backend stores data in a SQL Database using [database/sql](https://golang.org/pkg/database/sql/).

## Usage

```go
import "github.com/apuigsech/rest-layer-sql"
```

Create a resource storage handler with a given SQL driver, source and table:

```go
h := sqlStorage.NewHandler(DB_DRIVER, DB_SOURCE, DB_TABLE)
```

Bind this resource storage handled to a resource:

```go
index.Bind("resourceName", resourceSchema, h, resource.DefaultConf)
```

## Supported SQL Drivers

All supported SQL Drivers are listed [here](https://github.com/golang/go/wiki/SQLDrivers)


## Examples

Run the example:

```bash
$ go run example/example.go
2018/12/07 18:29:53 Serving API on http://localhost:8080
```

You can perform requests using [HTTPie](https://httpie.org/).

```bash
$ http GET :8080/units
HTTP/1.1 200 OK
Content-Length: 2
Content-Type: application/json
Date: Fri, 07 Dec 2018 17:31:23 GMT
Etag: W/"d41d8cd98f00b204e9800998ecf8427e"
X-Total: 0

[]

$ http POST :8080/units str="foo" int:=0
HTTP/1.1 201 Created
Content-Length: 139
Content-Location: /units/bg5at5bmvban389193a0
Content-Type: application/json
Date: Fri, 07 Dec 2018 17:32:05 GMT
Etag: W/"80b0d96e674f761e87950ccb19bdc279"
Last-Modified: Fri, 07 Dec 2018 17:32:05 GMT

{
    "created": "2018-12-07T18:32:05.290925+01:00",
    "id": "bg5at5bmvban389193a0",
    "int": 0,
    "str": "foo",
    "updated": "2018-12-07T18:32:05.290925+01:00"
}

$ http POST :8080/units str="bar" int:=1
HTTP/1.1 201 Created
Content-Length: 137
Content-Location: /units/bg5at73mvban389193ag
Content-Type: application/json
Date: Fri, 07 Dec 2018 17:32:12 GMT
Etag: W/"8bb72222f5979bc9f9259c0262807667"
Last-Modified: Fri, 07 Dec 2018 17:32:12 GMT

{
    "created": "2018-12-07T18:32:12.08373+01:00",
    "id": "bg5at73mvban389193ag",
    "int": 1,
    "str": "bar",
    "updated": "2018-12-07T18:32:12.08373+01:00"
}

$ http GET :8080/units
HTTP/1.1 200 OK
Content-Length: 365
Content-Type: application/json
Date: Fri, 07 Dec 2018 17:32:44 GMT
Etag: W/"0a320d03fee8e1fb268d3cc48fac391d"
X-Total: 0

[
    {
        "_etag": "80b0d96e674f761e87950ccb19bdc279",
        "created": "2018-12-07T18:32:05.290925+01:00",
        "id": "bg5at5bmvban389193a0",
        "int": 0,
        "str": "foo",
        "updated": "2018-12-07T18:32:05.290925+01:00"
    },
    {
        "_etag": "8bb72222f5979bc9f9259c0262807667",
        "created": "2018-12-07T18:32:12.08373+01:00",
        "id": "bg5at73mvban389193ag",
        "int": 1,
        "str": "bar",
        "updated": "2018-12-07T18:32:12.08373+01:00"
    }
]
```
