onramp
======

High-level, easy-to-use listeners and clients for I2P and onion URL's from Go.
Provides only the most widely-used functions in a basic way. It expects nothing
from the users, an otherwise empty instance of the structs will listen and dial
I2P Streaming and Tor TCP sessions successfully.

I2P(Garlic) Usage:
------------------

```Go
package main

import(
    "log"

    "github.com/eyedeekay/onramp"
)

func main() {
    garlic := &onramp.Garlic{}
	defer garlic.Close()
	listener, err := garlic.Listen()
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
}
```

Tor(Onion) Usage:
-----------------

```Go
package main

import(
    "log"

    "github.com/eyedeekay/onramp"
)

func main() {
    onion := &onramp.Onion{}
	defer garlic.Close()
	listener, err := onion.Listen()
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
}
```