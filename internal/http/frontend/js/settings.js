(() => {
  const auth = window.IDSAIAuth;
  const i18n = window.IDSAI18n;
  const roleSidebar = window.IDSAIRoleSidebar;
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";

  const ui = {
    sidebarHost: document.querySelector("[data-role-sidebar]"),
    sidebarAvatar: document.getElementById("profileAvatar"),
    sidebarName: document.getElementById("studentName"),
    sidebarEmail: document.getElementById("studentEmail"),
    logoutBtn: document.getElementById("logoutBtn"),

    heroAvatar: document.getElementById("settingsHeroAvatar"),
    heroName: document.getElementById("settingsHeroName"),
    heroEmail: document.getElementById("settingsHeroEmail"),
    heroRole: document.getElementById("settingsHeroRole"),
    heroScope: document.getElementById("settingsHeroScope"),

    settingsAvatar: document.getElementById("settingsAvatarPreview"),
    avatarInput: document.getElementById("avatarInput"),
    uploadAvatarBtn: document.getElementById("uploadAvatarBtn"),
    removeAvatarBtn: document.getElementById("removeAvatarBtn"),
    avatarStatus: document.getElementById("avatarStatus"),

    fullNameInput: document.getElementById("fullNameInput"),
    saveProfileBtn: document.getElementById("saveProfileBtn"),
    fullNameStatus: document.getElementById("fullNameStatus"),
    currentInstitutionLabel: document.getElementById("currentInstitutionLabel"),
    currentInstitutionInput: document.getElementById("currentInstitutionInput"),

    currentDepartmentInput: document.getElementById("currentDepartmentInput"),
    currentGroupInput: document.getElementById("currentGroupInput"),
    groupChangeBox: document.getElementById("groupChangeBox"),
    requestDepartmentInput: document.getElementById("requestDepartmentInput"),
    requestGroupInput: document.getElementById("requestGroupInput"),
    requestGroupPreview: document.getElementById("requestGroupPreview"),
    submitGroupChangeBtn: document.getElementById("submitGroupChangeBtn"),
    groupChangeStatus: document.getElementById("groupChangeStatus"),
    groupRequestsList: document.getElementById("groupRequestsList"),

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

  const groupsState = {
    departments: [],
  };
  const STATUS_ICON_HTML =
    `<svg viewBox="0 0 24 24" fill="none" focusable="false" aria-hidden="true">` +
      `<path d="M13 16h-1v-4h1m0-4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"></path>` +
    `</svg>`;

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

  function roleLabel(profile) {
    if (profile && profile.is_admin) {
      return "Администратор";
    }
    if (profile && profile.is_professor) {
      return "Преподаватель";
    }
    return "Студент";
  }

  function isSchoolProfile(profile) {
    return Boolean(profile && String(profile.education_type || "").toUpperCase() === "SCHOOL");
  }

  function institutionLabel(profile) {
    return isSchoolProfile(profile) ? "Школа" : "Вуз";
  }

  function scopeLabel(profile) {
    if (profile && profile.is_admin) {
      return "Полный доступ";
    }
    if (isSchoolProfile(profile) && profile && profile.school_class) {
      return `Класс ${profile.school_class}`;
    }
    if (profile && profile.group_code) {
      return `Группа ${profile.group_code}`;
    }
    if (profile && profile.department_code) {
      return `Кафедра ${profile.department_code}`;
    }
    return "Персональный доступ";
  }

  function renderAvatar(el, fallbackText, avatarURL) {
    if (!el) return;
    const url = String(avatarURL || "").trim();
    if (url) {
      el.classList.add("has-image");
      el.innerHTML = `<img src="${escapeHTML(url)}" alt="Avatar" width="64" height="64" loading="lazy" />`;
      return;
    }
    el.classList.remove("has-image");
    el.textContent = fallbackText;
  }

  function setStatus(el, message, kind) {
    if (!el) return;
    const tone = kind === "ok" || kind === "err" || kind === "info" || kind === "warn" ? kind : "";
    el.innerHTML = message
      ? `<span class="inline-status__icon" aria-hidden="true">${STATUS_ICON_HTML}</span><span class="inline-status__copy">${escapeHTML(message)}</span>`
      : "";
    el.classList.remove("err", "ok", "info", "warn");
    if (tone) {
      el.classList.add(tone);
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

  function buildGroupCode(departmentCode, rawGroup) {
    const department = String(departmentCode || "").trim().toUpperCase();
    const value = String(rawGroup || "").trim().toUpperCase();
    if (!department || !value) return "";
    if (/^\d{1,4}$/.test(value)) {
      return `${department}-${value}`;
    }
    return value;
  }

  function updateRequestedGroupState(options = {}) {
    const keepValue = Boolean(options.keepValue);
    const departmentCode = String(ui.requestDepartmentInput?.value || "").trim().toUpperCase();
    const hasDepartment = Boolean(departmentCode);

    if (ui.requestGroupInput) {
      ui.requestGroupInput.disabled = !hasDepartment;
      ui.requestGroupInput.placeholder = hasDepartment ? "Например 101" : "Сначала выберите кафедру";
      if (!hasDepartment || !keepValue) {
        ui.requestGroupInput.value = hasDepartment && keepValue ? ui.requestGroupInput.value : "";
      }
    }

    const rawGroup = String(ui.requestGroupInput?.value || "").trim().toUpperCase();
    if (ui.requestGroupPreview) {
      ui.requestGroupPreview.textContent = hasDepartment && rawGroup
        ? buildGroupCode(departmentCode, rawGroup)
        : hasDepartment
          ? `${departmentCode}-...`
          : "-";
    }
  }

  function toProfilePayload(data) {
    return {
      sub: String(data.user_id || ""),
      tenant_id: String(data.tenant_id || ""),
      faculty_id: String(data.faculty_id || ""),
      faculty_code: String(data.faculty_code || ""),
      department_id: String(data.department_id || ""),
      department_code: String(data.department_code || ""),
      group_id: String(data.group_id || ""),
      group_code: String(data.group_code || ""),
      group_number: data.group_number !== undefined && data.group_number !== null ? Number(data.group_number) : null,
      education_type: String(data.education_type || ""),
      school_class: String(data.school_class || ""),
      institution_provider: String(data.institution_provider || ""),
      institution_external_id: String(data.institution_external_id || ""),
      institution_name: String(data.institution_name || ""),
      institution_address: String(data.institution_address || ""),
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

  function bindLogout() {
    if (!ui.logoutBtn || ui.logoutBtn.dataset.bound === "1") return;
    ui.logoutBtn.dataset.bound = "1";
    ui.logoutBtn.addEventListener("click", () => auth.logout());
  }

  function syncSidebar(profile) {
    if (roleSidebar && typeof roleSidebar.renderSidebar === "function" && ui.sidebarHost) {
      roleSidebar.renderSidebar(ui.sidebarHost, {
        profile,
        role: "auto",
        active: "settings",
        adminViewMode: "links",
      });
    }

    ui.logoutBtn = document.getElementById("logoutBtn");
    bindLogout();
  }

  function applyProfile(data) {
    const profile = toProfilePayload(data);
    auth.persistProfile(profile);

    const name = profile.full_name || "Пользователь";
    const email = profile.email || "";
    const avatarURL = profile.avatar_url || "";

    localStorage.setItem(LS_STUDENT_NAME, name);
    localStorage.setItem(LS_STUDENT_EMAIL, email);
    localStorage.setItem(LS_AVATAR_URL, avatarURL);

    syncSidebar(profile);

    if (ui.sidebarName) ui.sidebarName.textContent = name;
    if (ui.sidebarEmail) ui.sidebarEmail.textContent = email;
    renderAvatar(ui.sidebarAvatar, initials(name, email), avatarURL);
    renderAvatar(ui.heroAvatar, initials(name, email), avatarURL);
    renderAvatar(ui.settingsAvatar, initials(name, email), avatarURL);

    if (ui.heroName) ui.heroName.textContent = name;
    if (ui.heroEmail) ui.heroEmail.textContent = email || "email не указан";
    if (ui.heroRole) ui.heroRole.textContent = roleLabel(profile);
    if (ui.heroScope) ui.heroScope.textContent = scopeLabel(profile);

    if (ui.fullNameInput) ui.fullNameInput.value = name;
    if (ui.currentInstitutionLabel) {
      ui.currentInstitutionLabel.textContent = institutionLabel(profile);
    }
    if (ui.currentInstitutionInput) {
      ui.currentInstitutionInput.value = profile.institution_name || "—";
      ui.currentInstitutionInput.title = profile.institution_address || profile.institution_name || "";
    }
    if (ui.currentEmailInput) ui.currentEmailInput.value = email;
    if (ui.currentDepartmentInput) {
      ui.currentDepartmentInput.value = isSchoolProfile(profile) ? "Школьное направление" : (profile.department_code || "—");
    }
    if (ui.currentGroupInput) {
      ui.currentGroupInput.value = isSchoolProfile(profile) ? (profile.school_class || "—") : (profile.group_code || "—");
    }

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

  function requestStatusLabel(status) {
    const value = String(status || "").toUpperCase();
    if (value === "PENDING") return "Ожидает";
    if (value === "APPROVED") return "Одобрено";
    if (value === "REJECTED") return "Отклонено";
    return value || "—";
  }

  function formatDateTime(raw) {
    if (!raw) return "—";
    const dt = new Date(raw);
    if (Number.isNaN(dt.getTime())) return "—";
    return i18n ? i18n.formatDateTime(dt) : dt.toLocaleString("ru-RU");
  }

  function renderGroupRequests(items) {
    if (!ui.groupRequestsList) return;
    const list = Array.isArray(items) ? items : [];
    ui.groupRequestsList.innerHTML = "";
    if (!list.length) {
      const empty = document.createElement("li");
      empty.textContent = "Заявок пока нет.";
      ui.groupRequestsList.appendChild(empty);
      return;
    }

    list.forEach((item) => {
      const li = document.createElement("li");
      const status = requestStatusLabel(item.status);
      const from = String(item.current_group_code || "—");
      const to = String(item.requested_group_code || "—");
      const createdAt = formatDateTime(item.created_at);
      const reviewedAt = formatDateTime(item.reviewed_at);
      const comment = String(item.admin_comment || "").trim();
      li.innerHTML = `
        <strong>${escapeHTML(status)}</strong>
        <div>${escapeHTML(from)} → ${escapeHTML(to)}</div>
        <div>Создано: ${escapeHTML(createdAt)}</div>
        <div>Проверено: ${escapeHTML(reviewedAt)}</div>
        ${comment ? `<div>Комментарий: ${escapeHTML(comment)}</div>` : ""}
      `;
      ui.groupRequestsList.appendChild(li);
    });
  }

  async function loadGroupRequests() {
    const data = await request("GET", "/v2/auth/settings/group-change-requests");
    const requests = Array.isArray(data.requests) ? data.requests : [];
    renderGroupRequests(requests);
  }

  function setDepartmentOptions(items) {
    if (!ui.requestDepartmentInput) return;
    const list = Array.isArray(items) ? items : [];
    ui.requestDepartmentInput.innerHTML = "";
    const first = document.createElement("option");
    first.value = "";
    first.textContent = "Выберите кафедру";
    ui.requestDepartmentInput.appendChild(first);
    list.forEach((item) => {
      const code = String(item.code || "").toUpperCase();
      const name = String(item.name || "");
      const option = document.createElement("option");
      option.value = code;
      option.textContent = name ? `${code} — ${name}` : code;
      ui.requestDepartmentInput.appendChild(option);
    });
  }

  async function loadDepartments() {
    const data = await request("GET", "/v2/auth/departments");
    const departments = Array.isArray(data.departments) ? data.departments : [];
    groupsState.departments = departments;
    setDepartmentOptions(departments);
    updateRequestedGroupState();
  }

  async function submitGroupChangeRequest() {
    const departmentCode = String(ui.requestDepartmentInput?.value || "").trim().toUpperCase();
    const groupCode = buildGroupCode(departmentCode, ui.requestGroupInput?.value);
    if (!departmentCode || !groupCode) {
      setStatus(ui.groupChangeStatus, "Выберите кафедру и введите номер группы.", "err");
      return;
    }
    if (!/^[A-Z]{2,8}-\d{1,4}$/.test(groupCode)) {
      setStatus(ui.groupChangeStatus, "Номер группы должен содержать от 1 до 4 цифр.", "err");
      return;
    }

    setLoading(ui.submitGroupChangeBtn, true, "Отправляем...");
    try {
      await request("POST", "/v2/auth/settings/group-change-requests", {
        department_code: departmentCode,
        group_code: groupCode,
      });
      setStatus(ui.groupChangeStatus, "Заявка отправлена администратору.", "ok");
      await Promise.all([loadSettings(), loadGroupRequests()]);
    } catch (err) {
      setStatus(ui.groupChangeStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.submitGroupChangeBtn, false, "Отправить заявку на смену группы");
    }
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

    if (ui.requestDepartmentInput) {
      ui.requestDepartmentInput.addEventListener("change", () => {
        updateRequestedGroupState();
      });
    }
    if (ui.requestGroupInput) {
      ui.requestGroupInput.addEventListener("input", () => {
        updateRequestedGroupState({ keepValue: true });
      });
    }
    if (ui.submitGroupChangeBtn) {
      ui.submitGroupChangeBtn.addEventListener("click", () => {
        void submitGroupChangeRequest();
      });
    }
  }

  async function bootstrap() {
    const profile = await auth.ensureSession();
    if (!profile) return;

    syncSidebar(profile);

    const isStudent = !profile.is_admin && !profile.is_professor;
    const isSchool = isSchoolProfile(profile);
    if ((!isStudent || isSchool) && ui.groupChangeBox) {
      ui.groupChangeBox.hidden = true;
    }

    wireEvents();
    applyQueryStatuses();

    try {
      await loadSettings();
      if (isStudent && !isSchool) {
        await loadDepartments();
        const currentDepartment = String(auth.getCachedProfile()?.department_code || "").toUpperCase();
        if (currentDepartment && ui.requestDepartmentInput) {
          ui.requestDepartmentInput.value = currentDepartment;
          updateRequestedGroupState({ keepValue: true });
        } else {
          updateRequestedGroupState();
        }
        await loadGroupRequests();
      }
    } catch (err) {
      setStatus(ui.fullNameStatus, `Не удалось загрузить настройки: ${err.message || String(err)}`, "err");
    } finally {
      auth.setPageLoading(false);
    }
  }

  void bootstrap().catch((err) => {
    auth.setPageLoading(false);
    setStatus(ui.fullNameStatus, `Не удалось загрузить настройки: ${err.message || String(err)}`, "err");
  });
})();
