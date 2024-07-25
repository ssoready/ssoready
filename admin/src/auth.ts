const sessionTokenKey = "s";

export function setSessionToken(s: string) {
  localStorage.setItem(sessionTokenKey, s);
}

export function getSessionToken(): string | null {
  return localStorage.getItem(sessionTokenKey);
}
