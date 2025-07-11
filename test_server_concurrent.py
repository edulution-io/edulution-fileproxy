import requests
import time
import tempfile
import os
import shutil
from concurrent.futures import ThreadPoolExecutor, as_completed
from collections import defaultdict

import urllib3
urllib3.disable_warnings()

# 🔧 Konfiguration
BASE_URL = "https://10.0.0.3:8443/webdav/default-school/students/niclass"
USERS = {f"netzint{i}": "Muster!" for i in range(1, 6)}
VERIFY_SSL = False  # bei self-signed certs auf False setzen

MAX_WORKERS=len(USERS)

def timed_request(label, method, url, auth, **kwargs):
    start = time.perf_counter()
    r = requests.request(method, url, auth=auth, verify=VERIFY_SSL, **kwargs)
    duration = (time.perf_counter() - start) * 1000  # ms
    if not r.status_code // 100 == 2:
        #raise Exception(f"{label} fehlgeschlagen ({r.status_code})\n -> {url}")
        print(f"{label} fehlgeschlagen ({r.status_code})\n -> {url}")
    return duration

def run_user_test(username, password):
    url = f"{BASE_URL}/{username}"
    auth = (username, password)
    temp_dir = tempfile.mkdtemp()
    local_file = os.path.join(temp_dir, "testfile.txt")
    with open(local_file, "w") as f:
        f.write(f"Inhalt {time.time()}")

    local_copy = os.path.join(temp_dir, "copy.txt")
    dav_file = "testfile.txt"
    renamed_file = "renamed.txt"
    subfolder = "subfolder"

    times = {}

    try:
        #times["cleanup_dir"] = timed_request("Ordner löschen", "DELETE", f"{url}/{subfolder}", auth)

        times["mkcol"] = timed_request("Ordner erstellen", "MKCOL", f"{url}/{subfolder}", auth)
        times["put"] = timed_request("Datei hochladen", "PUT", f"{url}/{dav_file}", auth, data=open(local_file, "rb"))
        
        start = time.perf_counter()
        r = requests.get(f"{url}/{dav_file}", auth=auth, verify=VERIFY_SSL)
        duration = (time.perf_counter() - start) * 1000
        if not r.ok:
            #raise Exception("Download fehlgeschlagen")
            print("Download fehlgeschlagen")
        with open(local_copy, "wb") as f:
            f.write(r.content)
        times["get"] = duration

        # vergleichen
        with open(local_file, "rb") as f1, open(local_copy, "rb") as f2:
            if f1.read() != f2.read():
                #raise Exception("Dateiinhalt stimmt nicht überein")
                print("Dateiinhalt stimmt nicht überein")

        times["move"] = timed_request("Datei umbenennen", "MOVE", f"{url}/{dav_file}", auth, headers={
            "Destination": f"{url}/{renamed_file}"
        })

        times["delete_file"] = timed_request("Datei löschen", "DELETE", f"{url}/{renamed_file}", auth)
        times["delete_dir"] = timed_request("Ordner löschen", "DELETE", f"{url}/{subfolder}", auth)
    finally:
        shutil.rmtree(temp_dir)

    return username, times

def main():
    results = []
    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        futures = [executor.submit(run_user_test, u, p) for u, p in USERS.items()]
        for future in as_completed(futures):
            try:
                username, times = future.result()
                print(f"✅ {username} erfolgreich getestet")
                for op, dur in times.items():
                    print(f"   ⏱ {op}: {dur:.1f} ms")
                results.append(times)
            except Exception as e:
                print(f"❌ Fehler bei Benutzer: {e}")

    # 📊 Statistik
    print("\n📊 Durchschnittszeiten (alle Benutzer):")
    op_durations = defaultdict(list)
    for result in results:
        for op, dur in result.items():
            op_durations[op].append(dur)

    for op, durations in sorted(op_durations.items()):
        avg = sum(durations) / len(durations)
        print(f" - {op:<12}: {avg:.1f} ms")

if __name__ == "__main__":
    main()
