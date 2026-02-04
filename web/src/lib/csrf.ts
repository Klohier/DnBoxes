export function getCsrfToken(): string | null {
  const match = document.cookie.match(/(^|;\s*)_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[2]) : null;
}
