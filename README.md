# whip-broadcast-go

whip-broadcast-go is a simple WHIP (https://datatracker.ietf.org/doc/draft-ietf-wish-whip/) client implementation in go using the WebRTC [Pion libraries](https://github.com/pion).
It includes a WHIPClient class and a simple command line client supporting screensharing to a WHIP ingestion endpoint.

It has been tested with [janus-gateway](https://github.com/meetecho/janus-gateway) with [whip-simple-server](https://github.com/lminiero/simple-whip-server) and also with mediasoup, millicast and cloudflare implementations.

## Installation

```
go build
```

## Running

```bash
./whip-broadcast-go -vc VIDEO_CODEC -t TOKEN WHIP_ENDPOINT_URL
```

To screenshare instead of using webcam:

```bash
./whip-broadcast-go -s -vc VIDEO_CODEC -t TOKEN WHIP_ENDPOINT_URL
```

The supported video codecs are `vp8` and `h264`.

For more information and additional configuration run:

```bash
./whip-broadcast-go -h
```
