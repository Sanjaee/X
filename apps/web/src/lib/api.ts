export function getApiBase(): string {
  if (typeof window !== "undefined") {
    const origin = window.location.origin;
    if (origin.includes("localhost") && !origin.includes(":3000")) {
      return `${origin}/api/v1`;
    }
  }
  let base = process.env.NEXT_PUBLIC_API_URL;
  if (!base) {
    if (typeof window !== "undefined") {
      base = `${window.location.origin}/api/v1`;
    } else {
      base = "http://api:8080/api/v1";
    }
  }
  base = base.replace(/\/+$/, "");
  if (base.endsWith("/api")) {
    base = `${base}/v1`;
  } else if (!base.includes("/api/v1")) {
    base = `${base}/api/v1`;
  }
  return base;
}

export function getStoredToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("x_session_token");
  }
  return null;
}

export function getStoredUser(): any | null {
  if (typeof window !== "undefined") {
    try {
      const data = localStorage.getItem("x_session_user");
      return data ? JSON.parse(data) : null;
    } catch {
      return null;
    }
  }
  return null;
}

export function setStoredUser(user: any | null) {
  if (typeof window !== "undefined") {
    if (user) {
      localStorage.setItem("x_session_user", JSON.stringify(user));
    } else {
      localStorage.removeItem("x_session_user");
    }
  }
}

export function setStoredToken(token: string | null) {
  if (typeof window !== "undefined") {
    if (token) {
      localStorage.setItem("x_session_token", token);
      document.cookie = `x_session_token=${token}; path=/; max-age=2592000; SameSite=Lax`;
    } else {
      localStorage.removeItem("x_session_token");
      localStorage.removeItem("x_session_user");
      document.cookie = `x_session_token=; path=/; max-age=0; SameSite=Lax`;
    }
  }
}

export async function apiRequest<T = any>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getStoredToken();
  const headers: Record<string, string> = {
    Accept: "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (!(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const base = getApiBase();
  const normalizedEndpoint = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
  const url = normalizedEndpoint.startsWith("/api/v1")
    ? (base.endsWith("/api/v1") ? `${base.slice(0, -7)}${normalizedEndpoint}` : `${base}${normalizedEndpoint}`)
    : `${base}${normalizedEndpoint}`;

  const res = await fetch(url, {
    ...options,
    headers,
    credentials: "include",
  });

  if (res.status === 204) {
    return {} as T;
  }

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 401 && typeof window !== "undefined" && token) {
      setStoredToken(null);
      window.dispatchEvent(new Event("x_auth_state_change"));
    }
    throw new Error(data.error || `HTTP ${res.status}`);
  }

  return data as T;
}
