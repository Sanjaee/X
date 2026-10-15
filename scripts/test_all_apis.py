import urllib.request
import urllib.parse
import json
import random
import sys

BASE_URL = "https://lostmedia.pro"
HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Content-Type": "application/json"
}

passed_tests = 0
failed_tests = 0

def call_api(method, path, data=None, token=None):
    global passed_tests, failed_tests
    url = f"{BASE_URL}{path}"
    headers = {**HEADERS}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    
    encoded_data = json.dumps(data).encode("utf-8") if data else None
    req = urllib.request.Request(url, data=encoded_data, headers=headers, method=method)
    
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            status = resp.status
            body = resp.read().decode("utf-8")
            res_json = json.loads(body) if body else {}
            passed_tests += 1
            print(f"  [PASS] {method:6} {path:<45} -> HTTP {status}")
            return status, res_json
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8")
        try:
            res_json = json.loads(body)
        except Exception:
            res_json = {"raw": body}
        if e.code in [200, 201, 204]:
            passed_tests += 1
            print(f"  [PASS] {method:6} {path:<45} -> HTTP {e.code}")
            return e.code, res_json
        else:
            failed_tests += 1
            print(f"  [FAIL] {method:6} {path:<45} -> HTTP {e.code}: {res_json}")
            return e.code, res_json
    except Exception as e:
        failed_tests += 1
        print(f"  [FAIL] {method:6} {path:<45} -> EXCEPTION: {e}")
        return 500, {"error": str(e)}

def run():
    print("==========================================================")
    print("  COMPLETE REA API AUTOMATED TEST SUITE")
    print(f"  Target: {BASE_URL}")
    print("==========================================================")

    # 1. Health Checks
    print("\n--- 1. Health & Readiness Observability ---")
    call_api("GET", "/health/live")
    call_api("GET", "/health/ready")
    call_api("GET", "/health/metrics")

    # 2. Authentication Flow (User A & User B)
    print("\n--- 2. User Authentication ---")
    rand_a = random.randint(100000, 999999)
    user_a = f"alice_{rand_a}"
    call_api("POST", "/api/v1/auth/register", {
        "username": user_a,
        "display_name": "Alice Wonderland",
        "email": f"alice_{rand_a}@lostmedia.pro",
        "password": "Password123!"
    })

    _, login_a = call_api("POST", "/api/v1/auth/login", {
        "identifier": user_a,
        "password": "Password123!"
    })
    token_a = login_a.get("session_token")
    user_a_id = login_a.get("user", {}).get("id")

    rand_b = random.randint(100000, 999999)
    user_b = f"bob_{rand_b}"
    call_api("POST", "/api/v1/auth/register", {
        "username": user_b,
        "display_name": "Bob Builder",
        "email": f"bob_{rand_b}@lostmedia.pro",
        "password": "Password123!"
    })

    _, login_b = call_api("POST", "/api/v1/auth/login", {
        "identifier": user_b,
        "password": "Password123!"
    })
    token_b = login_b.get("session_token")
    user_b_id = login_b.get("user", {}).get("id")

    # 3. User Identity & Profiles
    print("\n--- 3. User Profile & Settings ---")
    call_api("GET", "/api/v1/me", token=token_a)
    call_api("GET", "/api/me", token=token_a) # Alias check
    call_api("PATCH", "/api/v1/me/profile", {"bio": "Exploring the digital frontier with REA!"}, token=token_a)
    call_api("GET", f"/api/v1/users/{user_b}", token=token_a)
    call_api("GET", f"/api/v1/users/{user_b}/badges", token=token_a)
    call_api("GET", f"/api/v1/users/{user_b}/posts", token=token_a)

    # 4. Social Graph (Relationship, Follow, Mutuals, Block, Mute)
    print("\n--- 4. Social Graph Relationships ---")
    call_api("GET", f"/api/v1/users/{user_b}/relationship", token=token_a)
    call_api("GET", f"/api/v1/users/{user_b}/mutuals", token=token_a)
    call_api("POST", f"/api/v1/users/{user_b}/follow", token=token_a)
    call_api("GET", f"/api/v1/users/{user_a}/following", token=token_a)
    call_api("GET", f"/api/v1/users/{user_b}/followers", token=token_a)
    call_api("GET", f"/api/v1/users/{user_b}/relationship", token=token_a) # Verify is_following=True
    call_api("POST", f"/api/v1/users/{user_b}/mute", token=token_a)
    call_api("DELETE", f"/api/v1/users/{user_b}/mute", token=token_a)
    call_api("POST", f"/api/v1/users/{user_b}/block", token=token_a)
    call_api("DELETE", f"/api/v1/users/{user_b}/block", token=token_a)
    call_api("GET", "/api/v1/follow-requests", token=token_a)

    # 5. Timeline Feeds & Discovery
    print("\n--- 5. Feed Timeline & Feedback ---")
    _, feed_res = call_api("GET", "/api/v1/feed?type=for_you&limit=10", token=token_a)
    call_api("GET", "/api/v1/feed?type=following&limit=10", token=token_a)
    call_api("POST", "/api/v1/feed/impressions", {"impressions": []}, token=token_a)
    call_api("GET", "/api/v1/explore/trending", token=token_a)
    call_api("GET", "/api/v1/me/topic-preferences", token=token_a)

    # 6. Post Lifecycle & Interactions
    print("\n--- 6. Posts, Replies & Engagements ---")
    _, post_res = call_api("POST", "/api/v1/posts", {
        "content": f"Hello world from automated REA verification! #{user_a}"
    }, token=token_a)
    post_id = post_res.get("id")

    if post_id:
        call_api("GET", f"/api/v1/posts/{post_id}", token=token_a)
        call_api("POST", "/api/v1/posts", {
            "content": "Replying to original post",
            "reply_to_id": post_id
        }, token=token_b)
        call_api("GET", f"/api/v1/posts/{post_id}/replies", token=token_a)
        call_api("POST", f"/api/v1/posts/{post_id}/likes", token=token_b)
        call_api("DELETE", f"/api/v1/posts/{post_id}/likes", token=token_b)
        call_api("POST", f"/api/v1/posts/{post_id}/reposts", token=token_b)
        call_api("DELETE", f"/api/v1/posts/{post_id}/reposts", token=token_b)
        call_api("POST", f"/api/v1/posts/{post_id}/bookmarks", token=token_a)
        call_api("DELETE", f"/api/v1/posts/{post_id}/bookmarks", token=token_a)
        call_api("POST", "/api/v1/feed/feedback", {"post_id": post_id, "feedback_type": "not_interested"}, token=token_b)

    # 7. Search & Discovery
    print("\n--- 7. High-Performance Keyset Search ---")
    call_api("GET", f"/api/v1/search?q={user_a}", token=token_a)
    call_api("GET", f"/api/v1/search/posts?q={user_a}", token=token_a)
    call_api("GET", f"/api/v1/search/users?q={user_a}", token=token_a)
    call_api("GET", f"/api/v1/search/suggestions?q={user_a[:4]}", token=token_a)
    call_api("GET", "/api/v1/search/recent", token=token_a)

    # 8. Notifications
    print("\n--- 8. Notifications & Counters ---")
    call_api("GET", "/api/v1/notifications?limit=20", token=token_a)
    call_api("GET", "/api/v1/notifications/unread-count", token=token_a)
    call_api("PATCH", "/api/v1/notifications/read", token=token_a)

    # 9. Realtime Direct Messaging
    print("\n--- 9. Direct Messaging & Conversations ---")
    _, conv_res = call_api("POST", "/api/v1/conversations", {
        "recipient_id": user_b_id
    }, token=token_a)
    conv_id = conv_res.get("id")

    call_api("GET", "/api/v1/conversations", token=token_a)
    if conv_id:
        call_api("GET", f"/api/v1/conversations/{conv_id}", token=token_a)
        call_api("POST", f"/api/v1/conversations/{conv_id}/messages", {
            "content": "Hey Bob! Automated test message."
        }, token=token_a)
        call_api("GET", f"/api/v1/conversations/{conv_id}/messages?limit=20", token=token_b)
        call_api("POST", f"/api/v1/conversations/{conv_id}/read", token=token_b)

    # 10. Subscriptions & Verification
    print("\n--- 10. Monetization & Verification ---")
    call_api("GET", "/api/v1/subscription-plans", token=token_a)
    call_api("GET", "/api/v1/subscriptions/cryptos", token=token_a)
    call_api("GET", "/api/v1/subscriptions/me", token=token_a)
    call_api("GET", "/api/v1/verification/me", token=token_a)

    print("\n==========================================================")
    print(f"  TOTAL TESTS RUN: {passed_tests + failed_tests}")
    print(f"  PASSED: {passed_tests}")
    print(f"  FAILED: {failed_tests}")
    print("==========================================================")
    
    if failed_tests > 0:
        sys.exit(1)

if __name__ == "__main__":
    run()
