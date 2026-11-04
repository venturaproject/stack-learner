const KEY = 'session_expired'

/**
 * Call the moment a session is discovered to have died — a request 401'd and
 * the refresh attempt also failed, or a background-tab rehydrate found the
 * server no longer honors the cookies. sessionStorage (not React/module
 * state) survives axios.ts's onUnauthorized() hard `window.location.href`
 * redirect, so the login page can still read it after the reload.
 */
export function markSessionExpired(): void {
  try {
    sessionStorage.setItem(KEY, '1')
  } catch {
    // Blocked/unavailable storage (private mode, some embedded contexts) — the
    // banner is a nice-to-have, never worth failing the redirect over.
  }
}

/** Reads and clears the flag — call once on the login page's mount, so the banner only shows once, not on every re-render or subsequent visit. */
export function consumeSessionExpired(): boolean {
  try {
    const expired = sessionStorage.getItem(KEY) === '1'
    sessionStorage.removeItem(KEY)

    return expired
  } catch {
    return false
  }
}
