In Memory Cache
===

What is InMemoryCache?

InMemoryCache is a lightweight, high-performance in-memory key-value store built in Go, designed as a minimalist Redis alternative. It supports core commands like GET, SET, and EXPIRE, and enables real-time data flow with Redis Streams and a basic pub/sub system. The datastore features RDB-style persistence using compact binary serialization for reliable point-in-time snapshots. Additionally, it implements master-slave replication with offset-based synchronization and command propagation, making it a solid foundation for exploring distributed caching and reactive data systems.

===
Here’s your customized version in the same style, adapted for InMemoryCache:

⸻

Getting Started with InMemoryCache

The easiest way to get started with InMemoryCache is by cloning the repository and running the executable script with the desired configuration:

$ git clone https://github.com/your-username/in-memory-cache.git
$ cd in-memory-cache
$ ./execute.sh --port 6379

The above command will start the InMemoryCache server locally on port 6379.

You can interact with it using the standard Redis CLI or any Redis-compatible client:

$ redis-cli -p 6379



