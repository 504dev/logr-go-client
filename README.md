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
    logr "github.com/504dev/logr-go-client"
)

func main() {
    conf := logr.Config{
        Udp:        ":7776",
        PublicKey:  "MCAwDQYJKoZIhvcNAQEBBQADDwAwDAIFAMg7IrMCAwEAAQ",
        PrivateKey: "MC0CAQACBQDIOyKzAgMBAAECBQCHaZwRAgMA0nkCAwDziwIDAL+xAgJMKwICGq0=",
    }
    logger, _ = conf.NewLogger("hello.log")

    // Logger usage:
    logger.Info("Hello, Logr!")
    logger.Debug("Wonderful!")

    // Counter usage:
    logger.Inc("greeting", 1)

    // Widget usage:
    logger.Info("It's widget:", logr.Widget("inc", "greeting", 30))
}
```
