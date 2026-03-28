(() => {
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

  let cachedProfile = null;
  let refreshPromise = null;
  let profilePromise = null;
  const capabilityCache = new Map();
  const capabilityPromises = new Map();
  let alertLayer = null;
  let lastAlertMessage = "";
  let lastAlertAt = 0;

  function clearClientState() {
    cachedProfile = null;
    capabilityCache.clear();
    capabilityPromises.clear();
    localStorage.removeItem(LS_ACCESS);
    localStorage.removeItem(LS_REFRESH);
    localStorage.removeItem(LS_USER);
    localStorage.removeItem(LS_FACULTY);
    localStorage.removeItem(LS_STUDENT_NAME);
    localStorage.removeItem(LS_STUDENT_EMAIL);
    localStorage.removeItem(LS_AVATAR_URL);
    localStorage.removeItem(LS_IS_ADMIN);
    localStorage.removeItem(LS_IS_PROFESSOR);
  }

  function persistProfile(profile) {
    cachedProfile = profile || null;
    if (!profile) {
      clearClientState();
      return null;
    }

    localStorage.setItem(LS_USER, profile.sub || "");
    localStorage.setItem(LS_FACULTY, profile.faculty_id || "");
    localStorage.setItem(LS_STUDENT_NAME, profile.full_name || profile.email || "Student");
    localStorage.setItem(LS_STUDENT_EMAIL, profile.email || "");
    localStorage.setItem(LS_AVATAR_URL, profile.avatar_url || "");
    localStorage.setItem(LS_IS_ADMIN, profile.is_admin ? "1" : "0");
    localStorage.setItem(LS_IS_PROFESSOR, profile.is_professor ? "1" : "0");
    return profile;
  }

  function normalizeProfile(data) {
    if (!data || !data.user_id) return null;
    return {
      sub: String(data.user_id || ""),
      tenant_id: String(data.tenant_id || ""),
      faculty_id: String(data.faculty_id || ""),
      department_id: String(data.department_id || ""),
      department_code: String(data.department_code || ""),
      group_id: String(data.group_id || ""),
      group_code: String(data.group_code || ""),
      group_number: data.group_number !== undefined && data.group_number !== null ? Number(data.group_number) : null,
      email: String(data.email || ""),
      pending_email: String(data.pending_email || ""),
      pending_email_status: String(data.pending_email_status || ""),
      full_name: String(data.full_name || ""),
      avatar_url: String(data.avatar_url || ""),
      is_admin: Boolean(data.is_admin),
      is_professor: Boolean(data.is_professor),
      email_verified: Boolean(data.email_verified),
    };
  }

  function targetByProfile(profile) {
    if (profile && profile.is_admin) return "/dev/admin";
    if (profile && profile.is_professor) return "/dev/professor";
    return "/dev/projects";
  }

  function ensureAlertStyles() {
    if (document.getElementById("idsaiAppAlertStyles")) {
      return;
    }
    const style = document.createElement("style");
    style.id = "idsaiAppAlertStyles";
    style.textContent = `
      .idsai-app-alert-layer {
        position: fixed;
        top: 22px;
        right: 22px;
        z-index: 1400;
        width: min(360px, calc(100vw - 28px));
        display: grid;
        gap: 10px;
        pointer-events: none;
      }
      .idsai-app-alert {
        pointer-events: auto;
        display: grid;
        grid-template-columns: 1fr auto;
        gap: 10px;
        align-items: start;
        padding: 12px 14px;
        border-radius: 14px;
        border: 1px solid rgba(6, 78, 59, 0.18);
        background: linear-gradient(180deg, rgba(248, 255, 252, 0.98) 0%, rgba(236, 252, 245, 0.98) 100%);
        box-shadow: 0 18px 40px rgba(15, 23, 42, 0.16);
        color: #0f172a;
        animation: idsai-app-alert-in 180ms ease forwards;
      }
      .idsai-app-alert__body {
        min-width: 0;
      }
      .idsai-app-alert__title {
        margin: 0 0 4px;
        font: 800 14px/1.2 "Inter", "Segoe UI", sans-serif;
      }
      .idsai-app-alert__message {
        margin: 0;
        color: #475569;
        font: 500 13px/1.4 "Inter", "Segoe UI", sans-serif;
      }
      .idsai-app-alert__close {
        width: 26px;
        height: 26px;
        border: 0;
        border-radius: 999px;
        background: transparent;
        color: #64748b;
        cursor: pointer;
        font-size: 18px;
        line-height: 1;
      }
      .idsai-app-alert--warning {
        border-color: rgba(180, 83, 9, 0.18);
        background: linear-gradient(180deg, rgba(255, 251, 235, 0.98) 0%, rgba(254, 243, 199, 0.98) 100%);
      }
      .idsai-app-alert--warning .idsai-app-alert__title {
        color: #92400e;
      }
      .idsai-app-alert--error {
        border-color: rgba(190, 24, 93, 0.16);
        background: linear-gradient(180deg, rgba(255, 241, 242, 0.98) 0%, rgba(255, 228, 230, 0.98) 100%);
      }
      .idsai-app-alert--error .idsai-app-alert__title {
        color: #be123c;
      }
      .idsai-app-alert--success {
        border-color: rgba(6, 95, 70, 0.18);
        background: linear-gradient(180deg, rgba(236, 253, 245, 0.98) 0%, rgba(209, 250, 229, 0.98) 100%);
      }
      .idsai-app-alert--success .idsai-app-alert__title {
        color: #065f46;
      }
      @keyframes idsai-app-alert-in {
        from {
          opacity: 0;
          transform: translateY(-8px);
        }
        to {
          opacity: 1;
          transform: translateY(0);
        }
      }
      @media (max-width: 720px) {
        .idsai-app-alert-layer {
          top: 12px;
          right: 12px;
          left: 12px;
          width: auto;
        }
      }
    `;
    document.head.appendChild(style);
  }

  function ensureAlertLayer() {
    if (alertLayer && document.body.contains(alertLayer)) {
      return alertLayer;
    }
    if (!document.body) {
      return null;
    }
    ensureAlertStyles();
    alertLayer = document.getElementById("idsaiAppAlertLayer");
    if (alertLayer) {
      return alertLayer;
    }
    alertLayer = document.createElement("section");
    alertLayer.id = "idsaiAppAlertLayer";
    alertLayer.className = "idsai-app-alert-layer";
    alertLayer.setAttribute("aria-live", "polite");
    document.body.appendChild(alertLayer);
    return alertLayer;
  }

  function removeAlert(card) {
    if (!(card instanceof HTMLElement)) {
      return;
    }
    card.remove();
  }

  function showAlert(message, kind = "info", options = {}) {
    const text = String(message || "").trim();
    if (!text) {
      return;
    }

    const now = Date.now();
    const dedupeWindowMs = Number(options.dedupeWindowMs || 3200);
    if (text === lastAlertMessage && now - lastAlertAt < dedupeWindowMs) {
      return;
    }
    lastAlertMessage = text;
    lastAlertAt = now;

    const layer = ensureAlertLayer();
    if (!layer) {
      return;
    }

    const variant = kind === "error" || kind === "success" || kind === "warning" ? kind : "info";
    const titles = {
      info: "Уведомление",
      success: "Готово",
      warning: "Нет доступа",
      error: "Ошибка",
    };
    const card = document.createElement("article");
    card.className = `idsai-app-alert idsai-app-alert--${variant}`;
    card.innerHTML = `
      <div class="idsai-app-alert__body">
        <h4 class="idsai-app-alert__title">${titles[variant]}</h4>
        <p class="idsai-app-alert__message"></p>
      </div>
      <button type="button" class="idsai-app-alert__close" aria-label="Закрыть">×</button>
    `;
    const messageEl = card.querySelector(".idsai-app-alert__message");
    if (messageEl) {
      messageEl.textContent = text;
    }
    const closeBtn = card.querySelector(".idsai-app-alert__close");
    if (closeBtn) {
      closeBtn.addEventListener("click", () => removeAlert(card));
    }

    layer.appendChild(card);
    const ttlMs = Number(options.ttlMs || 4200);
    window.setTimeout(() => removeAlert(card), ttlMs);
  }

  function redirectToNotFound(fromURL) {
    const current = `${window.location.pathname || ""}${window.location.search || ""}`;
    if (current.startsWith("/dev/404") || current === "/404") {
      return;
    }
    const url = new URL("/dev/404", window.location.origin);
    const target = String(fromURL || current).trim();
    if (target) {
      url.searchParams.set("from", target);
    }
    window.location.href = `${url.pathname}${url.search}`;
  }

  function defaultScopeForProfile(profile) {
    if (!profile) return null;
    if (profile.is_admin) {
      return { type: "SYSTEM", id: "" };
    }
    if (profile.faculty_id) {
      return { type: "FACULTY", id: String(profile.faculty_id) };
    }
    if (profile.department_id) {
      return { type: "DEPARTMENT", id: String(profile.department_id) };
    }
    if (profile.tenant_id) {
      return { type: "TENANT", id: String(profile.tenant_id) };
    }
    return null;
  }

  function normalizeScope(scope, profile = cachedProfile) {
    const source = scope && typeof scope === "object" ? scope : defaultScopeForProfile(profile);
    if (!source || !source.type) return null;
    return {
      type: String(source.type || "").trim().toUpperCase(),
      id: String(source.id || "").trim(),
    };
  }

  function scopeKey(scope) {
    return `${scope.type}:${scope.id || ""}`;
  }

  function buildCapabilitiesURL(scope) {
    const url = new URL("/v2/auth/capabilities", window.location.origin);
    if (scope && scope.type) url.searchParams.set("scope_type", scope.type);
    if (scope && scope.id) url.searchParams.set("scope_id", scope.id);
    return `${url.pathname}${url.search}`;
  }

  function normalizeCapabilities(data) {
    if (!data || typeof data !== "object" || !Array.isArray(data.permissions)) {
      return [];
    }
    return data.permissions
      .map((item) => String(item || "").trim())
      .filter(Boolean);
  }

  async function rawFetch(url, options = {}) {
    const opts = { credentials: "same-origin", ...options };
    return fetch(url, opts);
  }

  async function parseResponse(resp) {
    const text = await resp.text();
    let data = {};
    try {
      data = text ? JSON.parse(text) : {};
    } catch (_) {
      data = text;
    }
    return data;
  }

  async function refreshSession() {
    if (refreshPromise) {
      return refreshPromise;
    }

    refreshPromise = (async () => {
      const resp = await rawFetch("/v2/auth/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      if (resp.status === 204) {
        clearClientState();
        return false;
      }
      if (!resp.ok) {
        clearClientState();
        return false;
      }
      return true;
    })();

    try {
      return await refreshPromise;
    } finally {
      refreshPromise = null;
    }
  }

  async function requestJSON(url, options = {}) {
    const skipRefresh = Boolean(options.skipAuthRefresh);
    const skipRedirect = Boolean(options.skipAuthRedirect);
    const skipAccessAlert = Boolean(options.skipAccessAlert);
    const fetchOptions = { ...options };
    const headers = { ...(options.headers || {}) };
    delete fetchOptions.skipAuthRefresh;
    delete fetchOptions.skipAuthRedirect;
    delete fetchOptions.skipAccessAlert;
    delete fetchOptions.headers;

    if (
      fetchOptions.body !== undefined &&
      fetchOptions.body !== null &&
      typeof fetchOptions.body === "object" &&
      !(fetchOptions.body instanceof FormData) &&
      !(fetchOptions.body instanceof URLSearchParams) &&
      !(fetchOptions.body instanceof Blob) &&
      !(fetchOptions.body instanceof ArrayBuffer)
    ) {
      const hasContentType = Object.keys(headers).some((key) => key.toLowerCase() === "content-type");
      if (!hasContentType) {
        headers["Content-Type"] = "application/json";
      }
      fetchOptions.body = JSON.stringify(fetchOptions.body);
    }

    fetchOptions.headers = headers;

    let resp = await rawFetch(url, fetchOptions);
    if (resp.status === 401 && !skipRefresh && !String(url).startsWith("/v2/auth/refresh")) {
      const refreshed = await refreshSession();
      if (refreshed) {
        resp = await rawFetch(url, fetchOptions);
      }
    }

    const data = await parseResponse(resp);
    if (resp.status === 403 && data && typeof data === "object") {
      const current = String(data.error || "").trim().toLowerCase();
      if (!current || current === "forbidden" || current === "access denied") {
        data.error = "Нет доступа. Попробуйте сменить контекст.";
      }
      if (!skipAccessAlert && !window.location.pathname.startsWith("/dev/login")) {
        showAlert(data.error, "warning");
      }
    }
    if (resp.status === 401 && !skipRedirect && !window.location.pathname.startsWith("/dev/login")) {
      clearClientState();
      window.location.href = "/dev/login";
    }
    return { resp, data };
  }

  async function fetchCurrentProfile() {
    if (cachedProfile) {
      return cachedProfile;
    }
    if (profilePromise) {
      return profilePromise;
    }

    profilePromise = (async () => {
    const { resp, data } = await requestJSON("/v2/auth/me", {
      method: "GET",
      skipAuthRedirect: true,
    });
    if (!resp.ok) {
      return null;
    }
      const profile = persistProfile(normalizeProfile(data));
      if (profile) {
        void loadCapabilities(defaultScopeForProfile(profile));
      }
      return profile;
    })();

    try {
      return await profilePromise;
    } finally {
      profilePromise = null;
    }
  }

  async function ensureSession(expectedRole, options = {}) {
    const redirectOnMissing = options.redirectOnMissing !== false;
    const profile = await fetchCurrentProfile();
    if (!profile) {
      if (redirectOnMissing) {
        window.location.href = "/dev/login";
      }
      return null;
    }

    if (expectedRole === "admin" && !profile.is_admin) {
      window.location.href = targetByProfile(profile);
      return null;
    }
    if (expectedRole === "professor" && profile.is_admin) {
      window.location.href = "/dev/admin";
      return null;
    }
    if (expectedRole === "professor" && !profile.is_professor) {
      window.location.href = "/dev/projects";
      return null;
    }
    if (expectedRole === "student" && profile.is_admin) {
      window.location.href = "/dev/admin";
      return null;
    }
    if (expectedRole === "student" && profile.is_professor) {
      window.location.href = "/dev/professor";
      return null;
    }

    return profile;
  }

  async function logout() {
    await requestJSON("/v2/auth/logout", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: "{}",
      skipAuthRefresh: true,
      skipAuthRedirect: true,
    });
    clearClientState();
    window.location.href = "/dev/login";
  }

  async function loadCapabilities(scope) {
    const profile = cachedProfile || await fetchCurrentProfile();
    const resolvedScope = normalizeScope(scope, profile);
    if (!resolvedScope) {
      return new Set();
    }

    const key = scopeKey(resolvedScope);
    const cached = capabilityCache.get(key);
    if (cached) {
      return cached;
    }
    if (capabilityPromises.has(key)) {
      return capabilityPromises.get(key);
    }

    const promise = (async () => {
      const { resp, data } = await requestJSON(buildCapabilitiesURL(resolvedScope), {
        method: "GET",
        skipAuthRedirect: true,
      });

      if (!resp.ok) {
        const empty = new Set();
        capabilityCache.set(key, empty);
        return empty;
      }

      const permissions = new Set(normalizeCapabilities(data));
      capabilityCache.set(key, permissions);
      return permissions;
    })();

    capabilityPromises.set(key, promise);
    try {
      return await promise;
    } finally {
      capabilityPromises.delete(key);
    }
  }

  function canCached(permission, scope) {
    const resolvedScope = normalizeScope(scope);
    if (!resolvedScope) {
      return false;
    }
    const permissions = capabilityCache.get(scopeKey(resolvedScope));
    return Boolean(permissions && permissions.has(String(permission || "").trim()));
  }

  async function can(permission, scope) {
    const permissions = await loadCapabilities(scope);
    return permissions.has(String(permission || "").trim());
  }

  window.IDSAIAuth = {
    can,
    canCached,
    clearClientState,
    ensureSession,
    fetchCurrentProfile,
    getCachedCapabilities: (scope) => {
      const resolvedScope = normalizeScope(scope);
      if (!resolvedScope) return [];
      return Array.from(capabilityCache.get(scopeKey(resolvedScope)) || []);
    },
    getCachedProfile: () => cachedProfile,
    getDefaultScope: () => normalizeScope(defaultScopeForProfile(cachedProfile)),
    loadCapabilities,
    persistProfile,
    requestJSON,
    redirectToNotFound,
    showAlert,
    logout,
    targetByProfile,
  };
})();
