# DynamiDB - In-Memory Key-Value Store

DynamiDB is an efficient in-memory key-value store written in Go, inspired by Memcached. It includes features such as LRU caching, TTL management, and memory slab allocation to optimize storage and retrieval performance. This project is suitable for lightweight caching solutions and supports simple TCP-based client interactions.

---

## Features

1. **LRU Cache**
   - Implements a Least Recently Used (LRU) cache for efficient memory management.
   - Automatically removes the least recently used items when the cache reaches its capacity.

2. **TTL Management**
   - Supports time-to-live (TTL) expiration for cache entries.
   - Periodically cleans up expired items to free memory.

3. **Memory Slab Allocation**
   - Uses slab allocation to optimize memory management for variable-sized entries.
   - Prevents fragmentation and supports efficient memory reuse.

4. **TCP-Based Protocol**
   - Handles client requests with simple commands like `SET`, `GET`, `DELETE`, and `QUIT`.

5. **Concurrency**
   - Utilizes goroutines for handling multiple client connections simultaneously.
   - Ensures thread safety with mutexes and channels.

---

## Commands

### 1. `SET`
Stores a key-value pair in the cache with an optional expiration time.

**Usage:**
```
SET <key> <expire_seconds> <value_size>
<value>

```
- `<key>`: Key for the value.
- `<expire_seconds>`: Time in seconds before the key expires. Use `0` for no expiration.
- `<value_size>`: Size of the value in bytes.
- `<value>`: Actual data to store.

**Example:**
```
SET mykey 60 5
hello

```

### 2. `GET`
Retrieves the value associated with a key.

**Usage:**
```
GET <key>

```
- `<key>`: Key to retrieve the value for.

**Example:**
```
GET mykey

```

**Response:**
```
VALUE <key> <value_size>
<value>
END

```

### 3. `DELETE`
Deletes a key-value pair from the cache.

**Usage:**
```
DELETE <key>

```
- `<key>`: Key to delete.

**Example:**
```
DELETE mykey

```

### 4. `QUIT`
Closes the client connection.

**Usage:**
```
QUIT

```

---

## Project Structure

- **`main.go`**: Entry point for the application, initializes components, and starts the TCP server.
- **`ttl.go`**: Manages TTL expiration for cache entries.
- **`lru.go`**: Implements the LRU caching mechanism.
- **`server.go`**: Handles client requests and executes commands.
- **`slab.go`**: Implements memory slab allocation and defragmentation.

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Feralthedogg/DynamiDB.git
   cd DynamiDB
   ```

2. Build the project:
   ```bash
   go build -o dynamidb
   ```

3. Run the server:
   ```bash
   ./dynamidb
   ```

---

## Configuration

- Default port: `11212`
- LRU Cache capacity: `1000` items
- Slab sizes: `{64, 128, 256, 1024, 4096}` bytes

These settings can be customized in the source code.

---

## Example Client

You can use tools like `netcat` or `telnet` to interact with DynamiDB:

```bash
echo -e "SET key1 30 5\r\nhello\r\n" | nc localhost 11212
echo -e "GET key1\r\n" | nc localhost 11212
echo -e "DELETE key1\r\n" | nc localhost 11212
```

---