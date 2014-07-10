# Reddan Simple GO-lang Redis client
I wrote this (initial version of) redis client during World Cup 2014 Match between Costarica and Greece (that boring it was).
Currently it support RESP only, no push/receive protocol yet. Still need to implement rest of commands.
SORT is not implemented because i need to refactor how i handle responses with many results of different types.

# Version
0.1

# Changelog
0.1 - Initial build supporting just basic RESP commands

# Example of usage

You will need to `go get` it and import it :)
```
import("github.com/JakubOboza/reddan/redis")

```
essential example:

```
  client, _ :=  redis.Dial("localhost:6379")

  val, _ := client.Get("keyName")

  fmt.Printf("Key value: '%s'\n", val)

```

full working example:

```
package main

import (
  "fmt"
  "github.com/JakubOboza/reddan/redis"
)

func main() {

  client, err :=  redis.Dial("localhost:6379")

  if err != nil {
    fmt.Printf("Error connecting to Redis: %s\n", err)
    return
  }

  client.Set("name", "kuba")

  res, err := client.Get("name")

  if err != nil {
    fmt.Printf("Error with GET: %s\n", err)
    return
  }

  fmt.Printf("My name is: '%s'\n", res)
}
```

# Supported commands

Currently it supports only this commands:
```
Close
Get
Set
Ping
Del
Exists
Expire
ExpireAt
Ttl
Keys
Move
Persist
Pexpire
PexpireAt
Pttl
RandomKey
Rename
RenameNx
Type
Append
Strlen
Incr
Decr
Lpush
LpushX
Rpush
RpushX
Lpop
Rpop
BlPop
BrPop
Lrange
Llen
Lindex
Lrem
Lset
Ltrim
Sadd
Smembers
Scard
Sdiff
SdiffStore
Sinter
SinterStore
Sismember
Smove
Spop
SrandMember
SrandMemberX
Srem
Sunion
SunionStore
RunCommand
RunArrayCommnad
```

# Cheers
Jakub Oboza