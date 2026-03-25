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

  function clearClientState() {
    cachedProfile = null;
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
    const fetchOptions = { ...options };
    delete fetchOptions.skipAuthRefresh;
    delete fetchOptions.skipAuthRedirect;

    let resp = await rawFetch(url, fetchOptions);
    if (resp.status === 401 && !skipRefresh && !String(url).startsWith("/v2/auth/refresh")) {
      const refreshed = await refreshSession();
      if (refreshed) {
        resp = await rawFetch(url, fetchOptions);
      }
    }

    const data = await parseResponse(resp);
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
    return persistProfile(normalizeProfile(data));
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

  window.IDSAIAuth = {
    clearClientState,
    ensureSession,
    fetchCurrentProfile,
    getCachedProfile: () => cachedProfile,
    persistProfile,
    requestJSON,
    logout,
    targetByProfile,
  };
})();
