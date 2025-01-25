import socket
import threading
import time
import random
import string

HOST = "127.0.0.1"
PORT = 11212

NUM_THREADS = 4
OPS_PER_THREAD = 5000
KEY_PREFIX = "benchKey"
VALUE_SIZE = 32

SET_RATIO = 0.8
GET_RATIO = 0.2

global_lock = threading.Lock()
total_ops_done = 0
start_time = 0
end_time = 0

def random_value(size):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=size)).encode()

def do_benchmark(thread_id):
    global total_ops_done

    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect((HOST, PORT))
    sock_file = s.makefile('rwb', buffering=0)

    def send_line(cmd):
        sock_file.write((cmd + "\r\n").encode())
        sock_file.flush()

    def read_line():
        line = sock_file.readline()
        return line.strip().decode() if line else ""

    random.seed(thread_id)

    local_count = 0
    for i in range(OPS_PER_THREAD):
        op_type = "set" if random.random() < SET_RATIO else "get"

        key = f"{KEY_PREFIX}_{thread_id}_{random.randint(0, 999999)}"
        if op_type == "set":
            val_bytes = random_value(VALUE_SIZE)
            cmd = f"set {key} 0 {len(val_bytes)}"
            send_line(cmd)
            sock_file.write(val_bytes + b"\r\n")
            sock_file.flush()

            resp = read_line()

        else:
            cmd = f"get {key}"
            send_line(cmd)
            lines = []
            while True:
                r = read_line()
                if not r:
                    break
                lines.append(r)
                if r == "END":
                    break

        local_count += 1

    send_line("quit")
    read_line()

    sock_file.close()
    s.close()

    with global_lock:
        total_ops_done += local_count

def main():
    global start_time, end_time, total_ops_done

    print(f"[INFO] Starting benchmark with {NUM_THREADS} threads, each doing {OPS_PER_THREAD} ops.")
    print(f"[INFO] Using {SET_RATIO*100:.0f}% SET and {GET_RATIO*100:.0f}% GET ratio.")

    threads = []
    start_time = time.time()

    for t_id in range(NUM_THREADS):
        th = threading.Thread(target=do_benchmark, args=(t_id,))
        threads.append(th)
        th.start()

    for th in threads:
        th.join()

    end_time = time.time()
    elapsed = end_time - start_time

    print("\n=== Benchmark result ===")
    print(f"Total ops: {total_ops_done}")
    print(f"Elapsed time: {elapsed:.2f} sec")
    if elapsed > 0:
        qps = total_ops_done / elapsed
        print(f"Throughput: {qps:.2f} ops/sec")

if __name__ == "__main__":
    main()