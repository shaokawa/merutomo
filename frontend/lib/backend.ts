const DEFAULT_BACKEND_URL = "http://localhost:8080";

function getBackendURL() {
  const rawURL = process.env.BACKEND_URL || DEFAULT_BACKEND_URL;
  return rawURL.endsWith("/") ? rawURL.slice(0, -1) : rawURL;
}

export async function proxyToBackend(path: string, init?: RequestInit) {
  const response = await fetch(`${getBackendURL()}${path}`, {
    ...init,
    cache: "no-store",
  });

  const payload = await response.text();
  const contentType = response.headers.get("content-type");

  return new Response(payload, {
    status: response.status,
    headers: contentType
      ? {
          "content-type": contentType,
        }
      : undefined,
  });
}
