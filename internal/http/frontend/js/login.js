(() => {
  const auth = window.IDSAIAuth;
  const statusEl = document.getElementById("status");
  const respEl = document.getElementById("resp");
  const tabLoginEl = document.getElementById("tabLogin");
  const tabRegisterEl = document.getElementById("tabRegister");
  const panelLoginEl = document.getElementById("panelLogin");
  const panelRegisterEl = document.getElementById("panelRegister");
  const forgotPasswordBtn = document.getElementById("forgotPasswordBtn");
  const regDepartmentEl = document.getElementById("regDepartment");
  const regGroupEl = document.getElementById("regGroup");
  const regGroupPreviewEl = document.getElementById("regGroupPreview");
  const registrationState = {
    departments: [],
  };

  function showConfirmDialog(options) {
    if (auth && typeof auth.showConfirmDialog === "function") {
      return auth.showConfirmDialog(options);
    }
    return Promise.resolve(window.confirm(String((options && options.message) || "")));
  }

  function showFormDialog(options) {
    if (auth && typeof auth.showFormDialog === "function") {
      return auth.showFormDialog(options);
    }
    return Promise.resolve(null);
  }

  function passwordPolicyError(password, minLength = 8) {
    const raw = String(password || "");
    if (raw.length < minLength) {
      return `Пароль должен быть не короче ${minLength} символов.`;
    }
    if (!/[A-Za-zА-Яа-яЁё]/.test(raw) || !/\d/.test(raw)) {
      return "Пароль должен содержать буквы и цифры.";
    }
    return "";
  }

  function passwordResetRequestError(message) {
    const text = String(message || "").trim().toLowerCase();
    if (!text) {
      return "Не удалось отправить письмо для сброса пароля.";
    }
    if (text.includes("account not found or unavailable for password reset")) {
      return "Аккаунт с таким email не найден или для него недоступен сброс пароля.";
    }
    if (text.includes("too many attempts")) {
      return "Слишком много попыток. Попробуйте немного позже.";
    }
    return "Не удалось отправить письмо для сброса пароля.";
  }

  async function collectPasswordResetValues(options = {}) {
    const requireCode = Boolean(options.requireCode);
    const email = String(options.email || "").trim();
    return showFormDialog({
      title: requireCode ? "Подтвердите код сброса" : "Придумайте новый пароль",
      message: requireCode
        ? `Мы отправили 6-значный код на ${email || "ваш email"}. Введите его и задайте новый пароль.`
        : "Введите новый пароль для аккаунта. После подтверждения можно будет сразу войти с ним.",
      confirmText: "Обновить пароль",
      fields: [
        ...(requireCode ? [{
          name: "code",
          label: "Код из письма",
          type: "text",
          value: "",
          placeholder: "123456",
          required: true,
          inputmode: "numeric",
          pattern: "\\d{6}",
          maxLength: 6,
          autocomplete: "one-time-code",
        }] : []),
        {
          name: "password",
          label: "Новый пароль",
          type: "password",
          value: "",
          placeholder: "Минимум 8 символов",
          required: true,
          autocomplete: "new-password",
        },
        {
          name: "password2",
          label: "Повторите пароль",
          type: "password",
          value: "",
          placeholder: "Повторите новый пароль",
          required: true,
          autocomplete: "new-password",
        },
      ],
      validate(values) {
        if (requireCode && !/^\d{6}$/.test(String(values.code || "").trim())) {
          return "Код должен содержать ровно 6 цифр.";
        }
        const password = String(values.password || "");
        const password2 = String(values.password2 || "");
        const policyError = passwordPolicyError(password, 8);
        if (policyError) {
          return policyError;
        }
        if (password !== password2) {
          return "Пароли не совпадают.";
        }
        return "";
      },
    });
  }

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

  function buildGroupCode(departmentCode, rawGroup) {
    const department = String(departmentCode || "").trim().toUpperCase();
    const value = String(rawGroup || "").trim().toUpperCase();
    if (!department || !value) return "";
    if (/^\d{1,4}$/.test(value)) {
      return `${department}-${value}`;
    }
    return value;
  }

  function updateRegistrationGroupField() {
    const departmentCode = String(regDepartmentEl?.value || "").trim().toUpperCase();
    const rawGroup = String(regGroupEl?.value || "").trim().toUpperCase();
    const hasDepartment = Boolean(departmentCode);

    if (regGroupEl) {
      regGroupEl.disabled = !hasDepartment;
      regGroupEl.placeholder = hasDepartment ? "Например 101" : "Сначала выберите кафедру";
      if (!hasDepartment) {
        regGroupEl.value = "";
      }
    }

    if (regGroupPreviewEl) {
      regGroupPreviewEl.textContent = hasDepartment && rawGroup
        ? buildGroupCode(departmentCode, rawGroup)
        : hasDepartment
          ? `${departmentCode}-...`
          : "-";
    }
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

  function setDepartmentOptions(items) {
    const list = Array.isArray(items) ? items : [];
    regDepartmentEl.innerHTML = "";
    const first = document.createElement("option");
    first.value = "";
    first.textContent = "Выберите кафедру";
    regDepartmentEl.appendChild(first);

    list.forEach((item) => {
      const code = String(item.code || "").toUpperCase();
      const name = String(item.name || "").trim();
      const opt = document.createElement("option");
      opt.value = code;
      opt.textContent = name ? `${code} — ${name}` : code;
      regDepartmentEl.appendChild(opt);
    });
  }

  async function loadDepartments() {
    const { resp, data } = await auth.requestJSON("/v2/auth/departments", {
      method: "GET",
      skipAuthRefresh: true,
      skipAuthRedirect: true,
    });
    if (!resp.ok) {
      throw new Error((data && data.error) || "Не удалось загрузить кафедры");
    }
    const items = Array.isArray(data.departments) ? data.departments : [];
    registrationState.departments = items;
    setDepartmentOptions(items);
    updateRegistrationGroupField();
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
        if (await showConfirmDialog({
          title: "Email еще не подтвержден",
          message: "Мы можем повторно отправить письмо подтверждения на этот адрес прямо сейчас.",
          confirmText: "Отправить письмо",
        })) {
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
      department_code: out.data.department_code,
      group_id: out.data.group_id,
      group_code: out.data.group_code,
      group_number: out.data.group_number,
      email: out.data.email,
      pending_email: out.data.pending_email,
      pending_email_status: out.data.pending_email_status,
      full_name: out.data.full_name,
      avatar_url: out.data.avatar_url,
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
    const department = String(regDepartmentEl.value || "").trim().toUpperCase();
    const groupCode = buildGroupCode(department, regGroupEl.value);
    const email = document.getElementById("regEmail").value.trim();
    const password = document.getElementById("regPassword").value;
    const password2 = document.getElementById("regPassword2").value;

    if (!email || !password || !department || !groupCode) {
      setStatus("Заполните обязательные поля и укажите номер группы.", false);
      return;
    }
    if (!/^[A-Z]{2,8}-\d{1,4}$/.test(groupCode)) {
      setStatus("Номер группы должен содержать от 1 до 4 цифр.", false);
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
      group_code: groupCode,
    });
    if (!out.resp.ok) {
      setStatus("Ошибка регистрации: " + out.resp.status, false);
      showJSON(out.data);
      return;
    }

    auth.clearClientState();
    setTab("login");
    if (out.data && out.data.status === "registered") {
      setStatus("Аккаунт создан. Теперь можно войти.", true);
    } else {
      setStatus("Аккаунт создан. Подтвердите email по ссылке из письма, затем войдите.", true);
    }
    showJSON(out.data);
  }

  async function requestPasswordReset() {
    const suggestedEmail = document.getElementById("loginEmail").value.trim();
    const values = await showFormDialog({
      title: "Сброс пароля",
      message: "Введите email, на который нужно отправить код для смены пароля.",
      confirmText: "Отправить код",
      fields: [{
        name: "email",
        label: "Email",
        type: "email",
        value: suggestedEmail,
        placeholder: "you@example.com",
        required: true,
        autocomplete: "email",
      }],
      validate(form) {
        const email = String(form.email || "").trim().toLowerCase();
        if (!email) {
          return "Email обязателен для сброса пароля.";
        }
        return "";
      },
    });
    if (!values) return;

    const normalizedEmail = String(values.email || "").trim().toLowerCase();
    if (!normalizedEmail) {
      setStatus("Email обязателен для сброса пароля.", false);
      return;
    }

    const out = await callJSON("/v2/auth/password-reset/request", { email: normalizedEmail });
    if (!out.resp.ok) {
      const errMessage = out.data && typeof out.data === "object" ? out.data.error : "";
      setStatus(passwordResetRequestError(errMessage), false);
      showJSON(out.data);
      return;
    }

    setStatus("Если такой аккаунт существует, код для сброса уже отправлен на email.", true);
    showJSON(out.data);
    await confirmPasswordResetByCode(normalizedEmail);
  }

  async function confirmPasswordResetFromCookie() {
    const values = await collectPasswordResetValues({ requireCode: false });
    if (!values) return;

    const password = String(values.password || "");
    if (!password) {
      setStatus("Новый пароль не был задан.", false);
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

  async function confirmPasswordResetByCode(email) {
    const values = await collectPasswordResetValues({ requireCode: true, email });
    if (!values) return;

    const normalizedCode = String(values.code || "").trim();
    const password = String(values.password || "");
    if (!normalizedCode || !password) {
      setStatus("Нужно указать код и новый пароль.", false);
      return;
    }

    const out = await callJSON("/v2/auth/password-reset/confirm", {
      email,
      code: normalizedCode,
      password,
    });
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
    const emailChange = params.get("email_change");
    const passwordChanged = params.get("password_changed");

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
      return;
    }

    if (emailChange === "1") {
      setStatus("Email успешно подтвержден и обновлен. Войдите с новым адресом.", true);
      return;
    }

    if (passwordChanged === "1") {
      setStatus("Пароль обновлен. Войдите снова с новым паролем.", true);
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
  regDepartmentEl?.addEventListener("change", () => {
    updateRegistrationGroupField();
  });
  regGroupEl?.addEventListener("input", () => {
    updateRegistrationGroupField();
  });

  loadDepartments()
    .catch((e) => {
      setStatus("Не удалось загрузить кафедры для регистрации", false);
      showJSON(e.message || String(e));
    });

  restoreExistingSession().catch(() => {});
})();
