# logr-go-client

[Logr] client library for Go.

[Logr]: https://github.com/504dev/logr

### Logger functions

* `Logger.Emerg`
* `Logger.Alert`
* `Logger.Crit`
* `Logger.Error`
* `Logger.Warn`
* `Logger.Notice`
* `Logger.Info`
* `Logger.Debug`

### Counter functions

* `Counter.Inc`
* `Counter.Avg`
* `Counter.Max`
* `Counter.Min`
* `Counter.Per`
* `Counter.Time`
* `Counter.Snippet`


Installing
----------

	go get github.com/504dev/logr-go-client
	
Usage
-----

``` golang
package main

import (
    logrc "github.com/504dev/logr-go-client"
    "rand"
)

func main() {
    conf := logrc.Config{
        Udp:        ":7776",
        PublicKey:  "MCAwDQYJKoZIhvcNAQEBBQADDwAwDAIFAMg7IrMCAwEAAQ",
        PrivateKey: "MC0CAQACBQDIOyKzAgMBAAECBQCHaZwRAgMA0nkCAwDziwIDAL+xAgJMKwICGq0=",
    }
    logr, _ := conf.NewLogger("hello.log")
    logr.Level = logrc.LevelInfo

    // Logger usage:
    logr.Info("Hello, Logr!")
    logr.Debug("Wonderful!")
    logr.Notice("Nice!")

    // Counter usage:
    logr.WatchSystem()  // watch load average, cpu, memory, disk
    logr.WatchProcess() // watch heap size, goroutines num
    logr.Avg("random", rand.float64())
    logr.Inc("greeting", 1)

    // Counter snippet usage:
    logr.Info("It's counter snippet:", logr.Snippet("avg", "random", 30))

    // Disable console output
    logr.Console = false
    logr.Info("this message will not be printed to the console")
}
```
