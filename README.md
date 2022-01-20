# FServer

Not a very helpful name isn't it?

Basically this is a simple TCP socket server that can be used to transfer a file over the network.

### Protocol

Header: `0x21 0x12 0x01`

Body: `[name-length]{2}[name]{name-length} [content-length]{4}[content]{content-length}`

Both `name-length` and `content-length` are little-endian 16 bits and 32 bits integers respectively.
