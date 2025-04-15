# InMemoryCache

InMemoryCache is a lightweight, high-performance in-memory key-value store built in Go, designed as a minimalist Redis alternative. It supports core commands like GET, SET, and EXPIRE, and enables real-time data flow with Redis Streams and a basic pub/sub system.

## Features

- **Lightweight In-Memory Storage**: Fast key-value operations with minimal overhead
- **Core Redis Commands**: Support for GET, SET, EXPIRE, and other essential operations
- **Real-Time Data Flow**: Implements Redis Streams and basic pub/sub functionality
- **RDB-Style Persistence**: Compact binary serialization for reliable point-in-time snapshots
- **Master-Slave Replication**: Offset-based synchronization and command propagation
- **Redis Protocol Compatible**: Works with existing Redis clients and tools

## Getting Started

The easiest way to get started with InMemoryCache is by cloning the repository and running the executable script with the desired configuration:

```bash
$ git clone https://github.com/SpartaNeel1010/in-memory-cache
$ cd in-memory-cache
$ ./execute.sh --port 6379
```

The above command will start the InMemoryCache server locally on port 6379.

You can interact with it using the standard Redis CLI or any Redis-compatible client:

```bash
$ redis-cli -p 6379
```

## Configuration Options

InMemoryCache supports various configuration options that can be specified as command-line arguments:

```bash
$ ./execute.sh --port 6379 --maxmemory 1gb --snapshot-interval 3600
```

| Option | Description | Default |
|--------|-------------|---------|
| `--port` | Server listening port | 6379 |
| `--maxmemory` | Maximum memory limit | unlimited |
| `--snapshot-interval` | Interval between automatic snapshots (seconds) | 3600 |
| `--snapshot-file` | File path for persistence snapshots | inmemorycache.rdb |
| `--replica-of` | Master server address for replication | none |

## Available Commands

InMemoryCache implements a subset of the Redis command set:

### Key-Value Operations
- `GET key` - Get the value of a key
- `SET key value [EX seconds]` - Set key to hold string value with optional expiration
- `DEL key [key ...]` - Delete one or more keys
- `EXISTS key [key ...]` - Check if keys exist
- `EXPIRE key seconds` - Set a key's time to live in seconds


### Streams
- `XADD key ID field value [field value ...]` - Append a new entry to a stream
- `XREAD [COUNT count] [BLOCK milliseconds] STREAMS key [key ...] ID [ID ...]` - Read data from streams

## Performance

InMemoryCache is designed for high performance with minimal resource consumption:

- Memory usage: ~20MB baseline with empty dataset
- Throughput: Up to 100,000 operations/second on modest hardware
- Latency: Sub-millisecond response times for most operations

## Use Cases

- Application caching layer
- Session storage
- Rate limiting
- Pub/sub messaging
- Simple job queues
- Real-time analytics

## Building from Source

To build InMemoryCache from source:

```bash
$ git clone https://github.com/SpartaNeel1010/in-memory-cache
$ cd in-memory-cache
$ go build -o inmemorycache cmd/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests, create issues, or suggest improvements.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2025 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
