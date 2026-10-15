import urllib.request
import json
import random

base = "https://lostmedia.pro"
rand = random.randint(10000, 99999)
username = f"user_{rand}"
email = f"user_{rand}@lostmedia.pro"
password = "TestPassword123!"

print("=== TESTING COMPLETE AUTHENTICATION FLOW ON HTTPS://LOSTMEDIA.PRO ===")

headers = {
    "Content-Type": "application/json",
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
}

# 1. Register
reg_data = json.dumps({
    "username": username,
    "display_name": f"User {rand}",
    "email": email,
    "password": password
}).encode('utf-8')

req1 = urllib.request.Request(f"{base}/api/v1/auth/register", data=reg_data, headers=headers)
with urllib.request.urlopen(req1) as resp:
    reg_res = json.loads(resp.read().decode('utf-8'))
print(f"1. Registration OK: user={reg_res['user']['username']} id={reg_res['user']['id']}")

# 2. Login via /api/v1/auth/login
login_data = json.dumps({
    "identifier": username,
    "password": password
}).encode('utf-8')

req2 = urllib.request.Request(f"{base}/api/v1/auth/login", data=login_data, headers=headers)
with urllib.request.urlopen(req2) as resp:
    login_res = json.loads(resp.read().decode('utf-8'))
token = login_res["session_token"]
print(f"2. Login /api/v1/auth/login OK: token={token[:16]}...")

# 3. Login alias via /api/auth/login
req3 = urllib.request.Request(f"{base}/api/auth/login", data=login_data, headers=headers)
with urllib.request.urlopen(req3) as resp:
    login_alias = json.loads(resp.read().decode('utf-8'))
print(f"3. Login /api/auth/login Alias OK: user={login_alias['user']['username']}")

# 4. Authenticated /api/v1/me
auth_headers = {
    "Authorization": f"Bearer {token}",
    "User-Agent": headers["User-Agent"]
}
req4 = urllib.request.Request(f"{base}/api/v1/me", headers=auth_headers)
with urllib.request.urlopen(req4) as resp:
    me = json.loads(resp.read().decode('utf-8'))
print(f"4. Authenticated /api/v1/me OK: username={me['username']} email={me['email']}")

# 5. Authenticated /api/me alias
req5 = urllib.request.Request(f"{base}/api/me", headers=auth_headers)
with urllib.request.urlopen(req5) as resp:
    me_alias = json.loads(resp.read().decode('utf-8'))
print(f"5. Authenticated /api/me Alias OK: username={me_alias['username']}")

print("\n>>> ALL 5 AUTH & LOGIN ENDPOINTS PASSED WITH 100% SUCCESS! <<<")
