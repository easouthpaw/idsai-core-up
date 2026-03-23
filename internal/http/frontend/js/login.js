(() => {
  const auth = window.IDSAIAuth;
  const statusEl = document.getElementById("status");
  const respEl = document.getElementById("resp");
  const tabLoginEl = document.getElementById("tabLogin");
  const tabRegisterEl = document.getElementById("tabRegister");
  const panelLoginEl = document.getElementById("panelLogin");
  const panelRegisterEl = document.getElementById("panelRegister");
  const forgotPasswordBtn = document.getElementById("forgotPasswordBtn");

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

  function targetByProfile(profile) {
    return auth.targetByProfile(profile);
  }

  async function callJSON(url, payload) {
    return auth.requestJSON(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      skipAuthRefresh: true,
      skipAuthRedirect: true,
    });
  }

  async function resendVerification(email) {
    const out = await callJSON("/v2/auth/verify-email/resend", { email });
    if (!out.resp.ok) {
      setStatus("Не удалось повторно отправить письмо подтверждения", false);
      showJSON(out.data);
      return;
    }
    setStatus("Письмо подтверждения отправлено повторно.", true);
    showJSON(out.data);
  }

  async function login() {
    const email = document.getElementById("loginEmail").value.trim();
    const password = document.getElementById("loginPassword").value;

    const out = await callJSON("/v2/auth/login", { email, password });
    if (!out.resp.ok) {
      if (out.resp.status === 403 && out.data && out.data.code === "email_verification_required") {
        setStatus("Подтвердите email перед входом.", false);
        showJSON(out.data);
        if (window.confirm("Email еще не подтвержден. Отправить письмо повторно?")) {
          await resendVerification(email);
        }
        return;
      }
      setStatus("Ошибка входа: " + out.resp.status, false);
      showJSON(out.data);
      return;
    }

    const profile = auth.persistProfile({
      sub: out.data.user_id,
      tenant_id: out.data.tenant_id,
      faculty_id: out.data.faculty_id,
      department_id: out.data.department_id,
      email: out.data.email,
      full_name: out.data.full_name,
      is_admin: out.data.is_admin,
      is_professor: out.data.is_professor,
      email_verified: out.data.email_verified,
    });

    setStatus("Вход выполнен. Переход в кабинет...", true);
    showJSON({ status: "authenticated", user_id: out.data.user_id, role: targetByProfile(profile) });
    window.location.href = targetByProfile(profile);
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
      setStatus("Ошибка регистрации: " + out.resp.status, false);
      showJSON(out.data);
      return;
    }

    auth.clearClientState();
    setTab("login");
    setStatus("Аккаунт создан. Подтвердите email по ссылке из письма, затем войдите.", true);
    showJSON(out.data);
  }

  async function requestPasswordReset() {
    const suggestedEmail = document.getElementById("loginEmail").value.trim();
    const email = window.prompt("Введите email для сброса пароля:", suggestedEmail);
    if (email === null) return;

    const out = await callJSON("/v2/auth/password-reset/request", { email });
    if (!out.resp.ok) {
      setStatus("Не удалось отправить письмо для сброса пароля.", false);
      showJSON(out.data);
      return;
    }

    setStatus("Если такой аккаунт существует, письмо для сброса уже отправлено.", true);
    showJSON(out.data);
  }

  async function confirmPasswordResetFromCookie() {
    const password = window.prompt("Введите новый пароль (минимум 10 символов):", "");
    if (password === null) return;
    const password2 = window.prompt("Повторите новый пароль:", "");
    if (password2 === null) return;
    if (password !== password2) {
      setStatus("Пароли не совпадают.", false);
      return;
    }

    const out = await callJSON("/v2/auth/password-reset/confirm", { password });
    if (!out.resp.ok) {
      setStatus("Не удалось обновить пароль.", false);
      showJSON(out.data);
      return;
    }

    auth.clearClientState();
    setStatus("Пароль обновлен. Теперь можно войти с новым паролем.", true);
    showJSON(out.data);
    window.history.replaceState({}, "", "/dev/login");
  }

  async function restoreExistingSession() {
    const profile = await auth.ensureSession(undefined, { redirectOnMissing: false });
    if (!profile) return;
    window.location.href = targetByProfile(profile);
  }

  function handleQueryState() {
    const params = new URLSearchParams(window.location.search);
    const verified = params.get("verified");
    const reset = params.get("reset");

    if (verified === "1") {
      setStatus("Email подтвержден. Теперь можно войти.", true);
    } else if (verified === "0") {
      setStatus("Ссылка подтверждения недействительна или истекла.", false);
    }

    if (reset === "1") {
      setStatus("Ссылка сброса подтверждена. Задайте новый пароль.", true);
      setTimeout(() => {
        confirmPasswordResetFromCookie().catch((e) => {
          setStatus("Сбой сброса пароля", false);
          showJSON(e.message || String(e));
        });
      }, 80);
      return;
    }

    if (reset === "expired") {
      setStatus("Ссылка сброса пароля недействительна или уже истекла.", false);
    }
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

  forgotPasswordBtn?.addEventListener("click", () => {
    requestPasswordReset().catch((e) => {
      setStatus("Сбой запроса сброса пароля", false);
      showJSON(e.message || String(e));
    });
  });

  handleQueryState();
  restoreExistingSession().catch(() => {});
})();
