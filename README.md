# kudos

Kudos is similar to a like. Just as you can add a like button at the bottom of a blog post, you can add a kudos button at the bottom of a blog post.

I really liked  http://kudosplease.com/ but their backend was really slow making it unusable, so I wrote my own.

I also host this little piece of software and you can use it free of charge: [blog post](https://www.niels-ole.com/kudos/blog/2016/11/04/now-with-kudos.html).

## Usage

Kudos requires a redis database that stores the state.

```
  -port string
        Listening port (default ":8080")
  -redis-address string
        redis host and port (default "localhost:6379")
  -redis-db int
        redis db to use
  -redis-password string
        redis password
```

### Docker

```
registry.gitlab.com/nielsole/kudos/kudos
```

### From source

```
go build main.go
```

