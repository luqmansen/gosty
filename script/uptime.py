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
        x = requests.get(url_test)
        ret.append({"time": str(datetime.now()), "status": x.status_code})
        time.sleep(delay)

        file.seek(0)
        file.truncate(0)
        json.dump(ret, file)

        cnt += delay
        if int(cnt) == 1:
            print("*", end="", flush=True)
            cnt = 0

    return ret


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--url', help="URL to check", type=str, required=True)
    parser.add_argument('--delay', help="delay between each request in second", type=float, required=True)
    parser.add_argument('--min', help="how many minute to run", type=float, required=True)

    args = parser.parse_args()
    with open(f'{datetime.now()}.json', 'w') as fout:
        uptime_check(args.url, args.delay, args.min, fout)



if __name__ == '__main__':
    main()
