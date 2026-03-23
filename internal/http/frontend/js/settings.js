(() => {
  const auth = window.IDSAIAuth;
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";

  const ui = {
    sidebarAvatar: document.getElementById("profileAvatar"),
    sidebarName: document.getElementById("studentName"),
    sidebarEmail: document.getElementById("studentEmail"),
    logoutBtn: document.getElementById("logoutBtn"),

    settingsAvatar: document.getElementById("settingsAvatarPreview"),
    avatarInput: document.getElementById("avatarInput"),
    uploadAvatarBtn: document.getElementById("uploadAvatarBtn"),
    removeAvatarBtn: document.getElementById("removeAvatarBtn"),
    avatarStatus: document.getElementById("avatarStatus"),

    fullNameInput: document.getElementById("fullNameInput"),
    saveProfileBtn: document.getElementById("saveProfileBtn"),
    fullNameStatus: document.getElementById("fullNameStatus"),

    currentEmailInput: document.getElementById("currentEmailInput"),
    newEmailInput: document.getElementById("newEmailInput"),
    startEmailChangeBtn: document.getElementById("startEmailChangeBtn"),
    resendEmailChangeBtn: document.getElementById("resendEmailChangeBtn"),
    emailTokenInput: document.getElementById("emailTokenInput"),
    confirmEmailChangeBtn: document.getElementById("confirmEmailChangeBtn"),
    emailStatus: document.getElementById("emailStatus"),

    currentPasswordInput: document.getElementById("currentPasswordInput"),
    newPasswordInput: document.getElementById("newPasswordInput"),
    confirmPasswordInput: document.getElementById("confirmPasswordInput"),
    changePasswordBtn: document.getElementById("changePasswordBtn"),
    passwordStatus: document.getElementById("passwordStatus"),
  };

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function initials(name, email) {
    const n = String(name || "").trim();
    if (n) {
      const parts = n.split(/\s+/).filter(Boolean);
      if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
      return n.slice(0, 2).toUpperCase();
    }
    const e = String(email || "").trim();
    return e ? e.slice(0, 2).toUpperCase() : "ST";
  }

  function renderAvatar(el, fallbackText, avatarURL) {
    if (!el) return;
    const url = String(avatarURL || "").trim();
    if (url) {
      el.classList.add("has-image");
      el.innerHTML = `<img src="${escapeHTML(url)}" alt="Avatar" loading="lazy" />`;
      return;
    }
    el.classList.remove("has-image");
    el.textContent = fallbackText;
  }

  function setStatus(el, message, kind) {
    if (!el) return;
    el.textContent = message || "";
    el.classList.remove("err", "ok");
    if (kind === "err" || kind === "ok") {
      el.classList.add(kind);
    }
  }

  function setLoading(btn, loading, loadingText) {
    if (!btn) return;
    if (!btn.dataset.baseText) {
      btn.dataset.baseText = btn.textContent || "";
    }
    btn.disabled = Boolean(loading);
    btn.textContent = loading ? loadingText : btn.dataset.baseText;
  }

  function toProfilePayload(data) {
    return {
      sub: String(data.user_id || ""),
      tenant_id: String(data.tenant_id || ""),
      faculty_id: String(data.faculty_id || ""),
      department_id: String(data.department_id || ""),
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

  function applyProfile(data) {
    const profile = toProfilePayload(data);
    auth.persistProfile(profile);

    const name = profile.full_name || "Student";
    const email = profile.email || "";
    const avatarURL = profile.avatar_url || "";

    localStorage.setItem(LS_STUDENT_NAME, name);
    localStorage.setItem(LS_STUDENT_EMAIL, email);
    localStorage.setItem(LS_AVATAR_URL, avatarURL);

    if (ui.sidebarName) ui.sidebarName.textContent = name;
    if (ui.sidebarEmail) ui.sidebarEmail.textContent = email;
    renderAvatar(ui.sidebarAvatar, initials(name, email), avatarURL);
    renderAvatar(ui.settingsAvatar, initials(name, email), avatarURL);

    if (ui.fullNameInput) ui.fullNameInput.value = name;
    if (ui.currentEmailInput) ui.currentEmailInput.value = email;

    if (ui.newEmailInput && profile.pending_email) {
      ui.newEmailInput.value = profile.pending_email;
    }

    const pendingStatus = profile.pending_email_status || "";
    if (pendingStatus === "verification_sent") {
      setStatus(ui.emailStatus, "Новый email ожидает подтверждения. Письмо отправлено.", "ok");
    } else if (profile.pending_email) {
      setStatus(ui.emailStatus, "Новый email ожидает подтверждения.", "ok");
    }
  }

  async function request(method, url, body, options = {}) {
    const { resp, data } = await auth.requestJSON(url, {
      method,
      headers: body !== undefined && !(body instanceof FormData)
        ? { "Content-Type": "application/json" }
        : undefined,
      body: body === undefined ? undefined : body instanceof FormData ? body : JSON.stringify(body),
      ...options,
    });
    if (!resp.ok) {
      const err = new Error(
        data && typeof data === "object" && data.error ? String(data.error) : `request failed (${resp.status})`
      );
      err.status = resp.status;
      throw err;
    }
    return data;
  }

  async function loadSettings() {
    const data = await request("GET", "/v2/auth/settings");
    applyProfile(data);
  }

  async function saveProfile() {
    const fullName = String(ui.fullNameInput.value || "").trim();
    if (fullName.length < 2) {
      setStatus(ui.fullNameStatus, "Введите корректное имя (минимум 2 символа).", "err");
      return;
    }
    setLoading(ui.saveProfileBtn, true, "Сохраняем...");
    try {
      const data = await request("PATCH", "/v2/auth/settings/profile", { full_name: fullName });
      applyProfile(data);
      setStatus(ui.fullNameStatus, "Профиль обновлен.", "ok");
    } catch (err) {
      setStatus(ui.fullNameStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.saveProfileBtn, false, "Сохранить профиль");
    }
  }

  async function startEmailChange() {
    const nextEmail = String(ui.newEmailInput.value || "").trim().toLowerCase();
    if (!nextEmail.includes("@")) {
      setStatus(ui.emailStatus, "Введите корректный email.", "err");
      return;
    }
    setLoading(ui.startEmailChangeBtn, true, "Отправляем...");
    try {
      await request("POST", "/v2/auth/settings/email/change", { email: nextEmail });
      setStatus(ui.emailStatus, "Письмо подтверждения отправлено на новый email.", "ok");
      await loadSettings();
    } catch (err) {
      setStatus(ui.emailStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.startEmailChangeBtn, false, "Отправить подтверждение");
    }
  }

  async function resendEmailChange() {
    setLoading(ui.resendEmailChangeBtn, true, "Отправляем...");
    try {
      await request("POST", "/v2/auth/settings/email/resend", {});
      setStatus(ui.emailStatus, "Письмо отправлено повторно.", "ok");
      await loadSettings();
    } catch (err) {
      setStatus(ui.emailStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.resendEmailChangeBtn, false, "Отправить повторно");
    }
  }

  async function confirmEmailChangeByToken() {
    const token = String(ui.emailTokenInput.value || "").trim();
    if (!token) {
      setStatus(ui.emailStatus, "Укажите код подтверждения.", "err");
      return;
    }
    if (!/^\d{6}$/.test(token) && !token.includes(".")) {
      setStatus(ui.emailStatus, "Код должен содержать 6 цифр.", "err");
      return;
    }
    setLoading(ui.confirmEmailChangeBtn, true, "Подтверждаем...");
    try {
      await request("POST", "/v2/auth/settings/email/confirm", { token });
      setStatus(ui.emailStatus, "Email подтвержден. Войдите заново.", "ok");
      setTimeout(() => {
        auth.clearClientState();
        window.location.href = "/dev/login?email_change=1";
      }, 900);
    } catch (err) {
      setStatus(ui.emailStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.confirmEmailChangeBtn, false, "Подтвердить email");
    }
  }

  async function uploadAvatar(file) {
    if (!file) return;
    const allowed = new Set(["image/jpeg", "image/png", "image/webp"]);
    if (!allowed.has(String(file.type || "").toLowerCase())) {
      setStatus(ui.avatarStatus, "Поддерживаются JPG/PNG/WEBP.", "err");
      return;
    }
    if (Number(file.size || 0) > 8 * 1024 * 1024) {
      setStatus(ui.avatarStatus, "Файл слишком большой (макс. 8MB).", "err");
      return;
    }

    const form = new FormData();
    form.append("avatar", file);

    setLoading(ui.uploadAvatarBtn, true, "Загружаем...");
    try {
      const data = await request("POST", "/v2/auth/settings/avatar", form);
      applyProfile(data);
      setStatus(ui.avatarStatus, "Аватар обновлен.", "ok");
    } catch (err) {
      setStatus(ui.avatarStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.uploadAvatarBtn, false, "Загрузить новое");
      if (ui.avatarInput) ui.avatarInput.value = "";
    }
  }

  async function removeAvatar() {
    setLoading(ui.removeAvatarBtn, true, "Удаляем...");
    try {
      const data = await request("DELETE", "/v2/auth/settings/avatar");
      applyProfile(data);
      setStatus(ui.avatarStatus, "Аватар удален.", "ok");
    } catch (err) {
      setStatus(ui.avatarStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.removeAvatarBtn, false, "Удалить");
    }
  }

  function validPassword(password) {
    const p = String(password || "");
    return p.length >= 8 && /[A-Za-zА-Яа-яЁё]/.test(p) && /\d/.test(p);
  }

  async function changePassword() {
    const currentPassword = String(ui.currentPasswordInput.value || "");
    const newPassword = String(ui.newPasswordInput.value || "");
    const confirmPassword = String(ui.confirmPasswordInput.value || "");

    if (!currentPassword || !newPassword || !confirmPassword) {
      setStatus(ui.passwordStatus, "Заполните все поля пароля.", "err");
      return;
    }
    if (newPassword !== confirmPassword) {
      setStatus(ui.passwordStatus, "Новый пароль и подтверждение не совпадают.", "err");
      return;
    }
    if (!validPassword(newPassword)) {
      setStatus(ui.passwordStatus, "Пароль должен быть не короче 8 символов и содержать буквы и цифры.", "err");
      return;
    }

    setLoading(ui.changePasswordBtn, true, "Сохраняем...");
    try {
      await request("POST", "/v2/auth/settings/password", {
        current_password: currentPassword,
        new_password: newPassword,
        confirm_password: confirmPassword,
      });
      setStatus(ui.passwordStatus, "Пароль обновлен. Выполняется выход из системы.", "ok");
      setTimeout(() => {
        auth.clearClientState();
        window.location.href = "/dev/login?password_changed=1";
      }, 900);
    } catch (err) {
      setStatus(ui.passwordStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.changePasswordBtn, false, "Обновить пароль");
    }
  }

  function applyQueryStatuses() {
    const params = new URLSearchParams(window.location.search || "");
    const emailChange = String(params.get("email_change") || "").trim();
    if (emailChange === "1") {
      setStatus(ui.emailStatus, "Email подтвержден. Войдите заново для обновления сессии.", "ok");
    } else if (emailChange === "expired") {
      setStatus(ui.emailStatus, "Ссылка подтверждения устарела. Отправьте письмо повторно.", "err");
    }
  }

  function wireEvents() {
    if (ui.logoutBtn) {
      ui.logoutBtn.addEventListener("click", () => auth.logout());
    }

    ui.saveProfileBtn.addEventListener("click", () => {
      void saveProfile();
    });
    ui.startEmailChangeBtn.addEventListener("click", () => {
      void startEmailChange();
    });
    ui.resendEmailChangeBtn.addEventListener("click", () => {
      void resendEmailChange();
    });
    ui.confirmEmailChangeBtn.addEventListener("click", () => {
      void confirmEmailChangeByToken();
    });
    ui.changePasswordBtn.addEventListener("click", () => {
      void changePassword();
    });

    ui.uploadAvatarBtn.addEventListener("click", () => {
      if (ui.avatarInput) ui.avatarInput.click();
    });
    if (ui.avatarInput) {
      ui.avatarInput.addEventListener("change", () => {
        const file = ui.avatarInput.files && ui.avatarInput.files[0] ? ui.avatarInput.files[0] : null;
        void uploadAvatar(file);
      });
    }
    ui.removeAvatarBtn.addEventListener("click", () => {
      void removeAvatar();
    });
  }

  async function bootstrap() {
    const profile = await auth.ensureSession();
    if (!profile) return;

    wireEvents();
    applyQueryStatuses();

    try {
      await loadSettings();
    } catch (err) {
      setStatus(ui.fullNameStatus, `Не удалось загрузить настройки: ${err.message || String(err)}`, "err");
    }
  }

  void bootstrap();
})();
