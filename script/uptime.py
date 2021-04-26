import time
import argparse
import requests
import json

def uptime_check(url_test: str, delay: int, run_min: int):
    ret = []
    t_end = time.time() + 60 * run_min
    cnt = 0
    while time.time() < t_end:
        x = requests.get(url_test)
        ret.append({"time": time.time(), "status": x.status_code})
        time.sleep(delay)
        cnt += 1
        if cnt == 100:
            print("*", end="", flush=True)
            cnt = 0

    return ret


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--url', help="URL to check", type=str, required=True)
    parser.add_argument('--delay', help="delay between each request in second", type=float, required=True)
    parser.add_argument('--min', help="how many minute to run", type=float, required=True)

    args = parser.parse_args()
    ret = uptime_check(args.url, args.delay, args.min)
    with open('result.json', 'w') as fout:
        json.dump(ret, fout)


if __name__ == '__main__':
    main()
