import axios from 'axios'
import type { AuthUser } from './auth'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface SessionData {
  user: AuthUser
  permissions: string[]
  roles: string[]
}

export interface AuthAdapter {
  /** Called before protected requests to attach extra headers. */
  getRequestHeaders(): Record<string, string>
  /** Perform login; returns session data on success, throws on failure. */
  login(credentials: Record<string, unknown>): Promise<SessionData>
  /** Perform logout. */
  logout(): Promise<void>
  /** Fetch the current session from the backend. Returns null if unauthenticated. */
  fetchSession(): Promise<SessionData | null>
  /** Attempt to refresh the access token. Returns true if successful. */
  refresh(): Promise<boolean>
  /** Called when the session cannot be recovered (e.g. redirect to /login). */
  onUnauthorized(): void
}

// ── HttpOnly cookie adapter ───────────────────────────────────────────────────
// Relies entirely on HttpOnly cookies — the token is never accessible from JS.
// The browser sends cookies automatically; no manual token management needed.

interface HttpOnlyCookieAdapterConfig {
  loginUrl: string
  logoutUrl: string
  meUrl: string
  refreshUrl: string
  /** Redirect path after an unrecoverable 401. Defaults to '/login'. */
  unauthorizedRedirect?: string
}

// Bare axios instance used only by the adapter — no interceptors, no cycles.
const _http = axios.create({ withCredentials: true, headers: { 'Content-Type': 'application/json' } })

export class HttpOnlyCookieAdapter implements AuthAdapter {
  private readonly cfg: HttpOnlyCookieAdapterConfig

  constructor(cfg: HttpOnlyCookieAdapterConfig) {
    this.cfg = cfg
  }

  getRequestHeaders(): Record<string, string> {
    // Double-submit CSRF: echo the non-HttpOnly `csrftoken` cookie back in a
    // header the backend compares against the cookie.
    if (typeof document === 'undefined') return {}
    const match = document.cookie.match(/(?:^|;\s*)csrftoken=([^;]+)/)
    return match ? { 'X-CSRFToken': decodeURIComponent(match[1]) } : {}
  }

  async login(credentials: Record<string, unknown>): Promise<SessionData> {
    const { data } = await _http.post(this.cfg.loginUrl, credentials)
    const user: AuthUser = data.user ?? data
    return { user, permissions: user.permissions ?? [], roles: user.roles ?? [] }
  }

  async logout(): Promise<void> {
    await _http.post(this.cfg.logoutUrl).catch(() => {})
  }

  async fetchSession(): Promise<SessionData | null> {
    try {
      const { data } = await _http.get(this.cfg.meUrl)
      const user: AuthUser = data.user ?? data
      return { user, permissions: user.permissions ?? [], roles: user.roles ?? [] }
    } catch {
      return null
    }
  }

  async refresh(): Promise<boolean> {
    try {
      await _http.post(this.cfg.refreshUrl)
      return true
    } catch {
      return false
    }
  }

  onUnauthorized(): void {
    window.location.href = this.cfg.unauthorizedRedirect ?? '/login'
  }
}

// ── Singleton ─────────────────────────────────────────────────────────────────

let _adapter: AuthAdapter = new HttpOnlyCookieAdapter({
  loginUrl:   '/api/auth/login',
  logoutUrl:  '/api/auth/logout',
  meUrl:      '/api/auth/me',
  refreshUrl: '/api/auth/refresh',
})

export const getAuthAdapter = (): AuthAdapter => _adapter

/** Call this once (e.g. in main.tsx) to swap the default adapter. */
export const setAuthAdapter = (adapter: AuthAdapter): void => {
  _adapter = adapter
}
