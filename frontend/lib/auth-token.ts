const TOKEN_STORAGE_KEY = "merutomo.access_token";

export function saveAuthToken(token: string) {
  window.localStorage.setItem(TOKEN_STORAGE_KEY, token);
}

export function getAuthToken() {
  return window.localStorage.getItem(TOKEN_STORAGE_KEY);
}

export function clearAuthToken() {
  window.localStorage.removeItem(TOKEN_STORAGE_KEY);
}
