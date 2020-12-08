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
* `Counter.Widget` bonus method!


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
    logger, _ = conf.NewLogger("hello.log")
    logger.Level = logrc.LevelInfo

    // Logger usage:
    logger.Info("Hello, Logr!")
    logger.Debug("Wonderful!")
    logger.Notice("Nice!")

    // Counter usage:
    logger.WatchSystem()  // watch load average, cpu, memory, disk
    logger.WatchProcess() // watch heap size, goroutines num
    logger.Avg("random", rand.float64())
    logger.Inc("greeting", 1)

    // Widget usage:
    logger.Info("It's widget:", logr.Widget("avg", "random", 30))

    // Disable console output
    logger.Console = false
    logger.Info("this message will not be printed to the console")
}
```
