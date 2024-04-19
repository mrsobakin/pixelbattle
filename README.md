[![Go Report Card](https://goreportcard.com/badge/mrsobakin/pixelbattle)](https://goreportcard.com/report/mrsobakin/pixelbattle) [![License](https://img.shields.io/badge/license-GPLv3-blue)](https://github.com/mrsobakin/pixelbattle/blob/master/LICENSE)

# 🎨 Pixelbattle - Backend ⚙️

Blazingly fast ⚡ and reliable 🦾 backend for the pixelbattle game.

This repository is a part of `pixelbattle` series - a full stack journey into creating a simple, production grade service.

## 📝 Protocol

1. The client connects to the websocket endpoint. Based on its `session` cookie, auth server provides a user id (or says that cookie is incorrect).
2. Server sends image of a canvas in png format to the client, as a binary message.
3. Client and server exchange symmetric canvas update messages in JSON format: `{"pos": [x, y], "color": [r, g, b]}`

## 🤔 Interesting facts

For this project I had to write my own [mpmc channel](internal/mpmc/mpmc.go) for broadcasting canvas changes. Maybe go already had something like it, or I could kludge something up with the go channels, but I've decided that it would be better just to write it from scratch.

Here's some info about this channel:

- It's a multi-producer multi-consumer channel (obviously).
- It's based on a ring buffer.
- The buffer is shared between all consumers and producers.
- Sending messages never fails and never blocks (unless you do simultaneous reads/writes).
- If the consumer is lagging behind by more than the buffer size, current read failes, and consumer queue is reset to the top message in channel.
- No messages are lost (unless consumer had lagged behind).
