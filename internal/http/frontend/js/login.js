(() => {
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_IS_ADMIN = "idsai_is_admin";

  const statusEl = document.getElementById("status");
  const respEl = document.getElementById("resp");
  const tabLoginEl = document.getElementById("tabLogin");
  const tabRegisterEl = document.getElementById("tabRegister");
  const panelLoginEl = document.getElementById("panelLogin");
  const panelRegisterEl = document.getElementById("panelRegister");

  function setStatus(msg, ok) {
    statusEl.textContent = msg;
    statusEl.className = "status " + (ok ? "ok" : "err");
  }

  function showJSON(v) {
    try {
      respEl.textContent = JSON.stringify(v, null, 2);
    } catch (_) {
      respEl.textContent = String(v);
    }
  }

  function setTab(tab) {
    const isLogin = tab === "login";
    tabLoginEl.classList.toggle("active", isLogin);
    tabRegisterEl.classList.toggle("active", !isLogin);
    panelLoginEl.classList.toggle("active", isLogin);
    panelRegisterEl.classList.toggle("active", !isLogin);
  }

  function decodePayload(token) {
    const parts = token.split(".");
    if (parts.length < 2) throw new Error("invalid JWT");
    let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const mod = payload.length % 4;
    if (mod > 0) payload += "=".repeat(4 - mod);
    return JSON.parse(atob(payload));
  }

  function deriveNameFromEmail(email) {
    const localPart = String(email || "").trim().split("@")[0] || "";
    if (!localPart) return "Студент";
    return localPart;
  }

  function saveSession(tokens) {
    localStorage.setItem(LS_ACCESS, tokens.access_token || "");
    localStorage.setItem(LS_REFRESH, tokens.refresh_token || "");

    const claims = decodePayload(tokens.access_token || "");
    if (!claims.sub || !claims.faculty_id) {
      throw new Error("token has no sub/faculty_id");
    }

    localStorage.setItem(LS_USER, claims.sub);
    localStorage.setItem(LS_FACULTY, claims.faculty_id);
    localStorage.setItem(LS_IS_ADMIN, claims.is_admin ? "1" : "0");
    return claims;
  }

  function targetByClaims(claims) {
    return claims && claims.is_admin ? "/dev/admin" : "/dev/projects";
  }

  async function callJSON(url, payload) {
    const started = performance.now();
    const resp = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const elapsed = Math.round(performance.now() - started);
    const text = await resp.text();
    let data = text;
    try {
      data = JSON.parse(text);
    } catch (_) {}

    return { resp, data, elapsed };
  }

  async function login() {
    const email = document.getElementById("loginEmail").value.trim();
    const password = document.getElementById("loginPassword").value;

    const out = await callJSON("/v2/auth/login", { email, password });
    if (!out.resp.ok) {
      setStatus("Ошибка входа: " + out.resp.status + " (" + out.elapsed + " ms)", false);
      showJSON(out.data);
      return;
    }

    const claims = saveSession(out.data);
    localStorage.setItem(LS_STUDENT_EMAIL, email);
    if (!localStorage.getItem(LS_STUDENT_NAME)) {
      localStorage.setItem(LS_STUDENT_NAME, deriveNameFromEmail(email));
    }
    setStatus("Вход выполнен. Переход в кабинет...", true);
    showJSON(out.data);
    window.location.href = targetByClaims(claims);
  }

  async function register() {
    const fullName = document.getElementById("regFullName").value.trim();
    const department = document.getElementById("regDepartment").value;
    const email = document.getElementById("regEmail").value.trim();
    const password = document.getElementById("regPassword").value;
    const password2 = document.getElementById("regPassword2").value;

    if (!email || !password || !department) {
      setStatus("Заполни обязательные поля регистрации", false);
      return;
    }

    if (password !== password2) {
      setStatus("Пароли не совпадают", false);
      return;
    }

    const out = await callJSON("/v2/auth/register", {
      email,
      password,
      full_name: fullName,
      department_code: department,
    });

    if (!out.resp.ok) {
      setStatus("Ошибка регистрации: " + out.resp.status + " (" + out.elapsed + " ms)", false);
      showJSON(out.data);
      return;
    }

    const claims = saveSession(out.data);
    localStorage.setItem(LS_STUDENT_EMAIL, email);
    localStorage.setItem(LS_STUDENT_NAME, fullName || deriveNameFromEmail(email));
    setStatus("Регистрация успешна. Переход в кабинет...", true);
    showJSON(out.data);
    window.location.href = targetByClaims(claims);
  }

  tabLoginEl.addEventListener("click", () => setTab("login"));
  tabRegisterEl.addEventListener("click", () => setTab("register"));

  document.getElementById("loginBtn").addEventListener("click", async () => {
    try {
      await login();
    } catch (e) {
      setStatus("Сбой запроса входа", false);
      showJSON(e.message || String(e));
    }
  });

  document.getElementById("registerBtn").addEventListener("click", async () => {
    try {
      await register();
    } catch (e) {
      setStatus("Сбой запроса регистрации", false);
      showJSON(e.message || String(e));
    }
  });

  const token = localStorage.getItem(LS_ACCESS) || "";
  if (token) {
    try {
      const claims = decodePayload(token);
      if (claims.sub && claims.faculty_id) {
        localStorage.setItem(LS_USER, claims.sub);
        localStorage.setItem(LS_FACULTY, claims.faculty_id);
        localStorage.setItem(LS_IS_ADMIN, claims.is_admin ? "1" : "0");
        window.location.href = targetByClaims(claims);
      }
    } catch (_) {}
  }
})();
