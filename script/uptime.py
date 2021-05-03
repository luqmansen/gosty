import time
from datetime import datetime
import argparse
import requests
import json

def uptime_check(url_test: str, delay: int, run_min: int, file):
    ret = []
    t_end = time.time() + 60 * run_min
    cnt = 0
    while time.time() < t_end:
        start = time.perf_counter()
        x = requests.get(url_test)
        ret.append({
            "time": str(datetime.now()),
            "status": x.status_code,
            "elapsed_time": x.elapsed.total_seconds(),
            "actual_request_time": time.perf_counter() - start,
        })

        time.sleep(delay)

        file.seek(0)
        file.truncate(0)
        json.dump(ret, file)

        cnt += delay
        if int(cnt) == 1:
            print("*", end="", flush=True)
            cnt = 0

    return ret


def avg_time(filename):
    f = open(filename)
    data = json.load(f)
    f.close()

    elapsed = sum(d["elapsed_time"] for d in data) / len(data)
    actual_resp = sum(d["actual_request_time"] for d in data) / len(data)
    return elapsed, actual_resp


def status_code(filename):
    f = open(filename)
    data = json.load(f)
    f.close()

    status_200 = len([d["status"] for d in data if d["status"] == 200])
    status_400 = len([d["status"] for d in data if 400 <= d["status"] < 500])
    status_500 = len([d["status"] for d in data if d["status"] >= 500])

    return status_200, status_400, status_500


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--url', help="URL to check", type=str, required=True)
    parser.add_argument('--delay', help="delay between each request in second", type=float, required=True)
    parser.add_argument('--min', help="how many minute to run", type=float, required=True)

    args = parser.parse_args()
    filename = f'{datetime.now()}.json'
    with open(filename, 'w') as fout:
        uptime_check(args.url, args.delay, args.min, fout)

    avgtime = avg_time(filename)
    print(f"\naverage elapsed_time:{avgtime[0]}")
    print(f"average actual response:{avgtime[1]}")
    status = status_code(filename)
    print(f"request cnt status 200: {status[0]}")
    print(f"request cnt status 400: {status[1]}")
    print(f"request cnt status 500: {status[2]}")


if __name__ == '__main__':
    main()
