(() => {
  const auth = window.IDSAIAuth;
  const roleSidebar = window.IDSAIRoleSidebar;

  const DEFAULT_EXTENDED_PROFILE = {
    headline: "",
    about: "",
    preferred_role: "",
    semester: "",
    availability: "",
    goals: "",
    github: "",
    telegram: "",
    portfolio: "",
    stacks: [],
    interests: [],
    updated_at: "",
  };

  const MAX_STACKS = 12;

  const ui = {
    sidebarHost: document.querySelector("[data-role-sidebar]"),
    logoutBtn: null,

    headerProfileAction: document.getElementById("headerProfileAction"),

    profileAvatarPreview: document.getElementById("profileAvatarPreview"),
    profileAvatarNote: document.getElementById("profileAvatarNote"),
    avatarInput: document.getElementById("avatarInput"),
    uploadAvatarBtn: document.getElementById("uploadAvatarBtn"),
    removeAvatarBtn: document.getElementById("removeAvatarBtn"),
    avatarStatus: document.getElementById("avatarStatus"),

    profileWorkspaceBadge: document.getElementById("profileWorkspaceBadge"),
    heroName: document.getElementById("heroName"),
    heroHeadline: document.getElementById("heroHeadline"),
    heroEmail: document.getElementById("heroEmail"),
    heroDepartment: document.getElementById("heroDepartment"),
    heroGroup: document.getElementById("heroGroup"),
    heroStackPreview: document.getElementById("heroStackPreview"),
    profileCompletionBadge: document.getElementById("profileCompletionBadge"),

    editProfileBtn: document.getElementById("editProfileBtn"),
    cancelEditBtn: document.getElementById("cancelEditBtn"),
    openSettingsBtn: document.getElementById("openSettingsBtn"),

    fullNameInput: document.getElementById("fullNameInput"),
    headlineInput: document.getElementById("headlineInput"),
    aboutInput: document.getElementById("aboutInput"),
    preferredRoleSelect: document.getElementById("preferredRoleSelect"),
    semesterInput: document.getElementById("semesterInput"),
    availabilitySelect: document.getElementById("availabilitySelect"),
    goalsInput: document.getElementById("goalsInput"),

    stackInput: document.getElementById("stackInput"),
    addStackBtn: document.getElementById("addStackBtn"),
    stackList: document.getElementById("stackList"),
    stackSuggestionBtns: Array.from(document.querySelectorAll("[data-stack-suggestion]")),

    githubInput: document.getElementById("githubInput"),
    telegramInput: document.getElementById("telegramInput"),
    portfolioInput: document.getElementById("portfolioInput"),
    interestGrid: document.querySelector(".interest-grid"),
    interestCheckboxes: Array.from(document.querySelectorAll('input[name="interests"]')),

    emailInput: document.getElementById("emailInput"),
    departmentInput: document.getElementById("departmentInput"),
    groupInput: document.getElementById("groupInput"),

    statsSection: document.querySelector(".profile-stats"),
    stackCountStat: document.getElementById("stackCountStat"),
    interestCountStat: document.getElementById("interestCountStat"),
    linksCountStat: document.getElementById("linksCountStat"),
    availabilityStat: document.getElementById("availabilityStat"),
    stackStatCard: document.getElementById("stackStatCard"),
    interestStatCard: document.getElementById("interestStatCard"),
    linksStatCard: document.getElementById("linksStatCard"),
    availabilityStatCard: document.getElementById("availabilityStatCard"),

    mainPanel: document.getElementById("profileMainPanel"),
    stackPanel: document.getElementById("profileStackPanel"),
    linksPanel: document.getElementById("profileLinksPanel"),
    interestsPanel: document.getElementById("profileInterestsPanel"),
    accountPanel: document.getElementById("profileAccountPanel"),

    saveProfileBtn: document.getElementById("saveProfileBtn"),
    profileStatus: document.getElementById("profileStatus"),
  };

  const state = {
    viewer: null,
    profile: null,
    savedExtended: { ...DEFAULT_EXTENDED_PROFILE },
    extended: { ...DEFAULT_EXTENDED_PROFILE },
    isOwnProfile: true,
    isEditMode: false,
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
    const text = String(name || "").trim();
    if (text) {
      const parts = text.split(/\s+/).filter(Boolean);
      if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
      return text.slice(0, 2).toUpperCase();
    }
    const mail = String(email || "").trim();
    return mail ? mail.slice(0, 2).toUpperCase() : "ST";
  }

  function renderAvatar(el, fallbackText, avatarURL) {
    if (!el) return;
    const url = String(avatarURL || "").trim();
    if (url) {
      el.classList.add("has-image");
      el.innerHTML = `<img src="${escapeHTML(url)}" alt="Avatar" width="296" height="296" loading="lazy" />`;
      return;
    }
    el.classList.remove("has-image");
    el.textContent = fallbackText;
  }

  function setStatus(el, message, kind) {
    if (!el) return;
    el.textContent = message || "";
    el.classList.remove("ok", "err");
    if (kind === "ok" || kind === "err") {
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

  function baseProfile(data) {
    const source = data && typeof data === "object" ? data : {};
    return {
      sub: String(source.user_id || source.sub || ""),
      tenant_id: String(source.tenant_id || ""),
      faculty_id: String(source.faculty_id || ""),
      department_id: String(source.department_id || ""),
      department_code: String(source.department_code || ""),
      group_id: String(source.group_id || ""),
      group_code: String(source.group_code || ""),
      group_number: source.group_number !== undefined && source.group_number !== null ? Number(source.group_number) : null,
      email: String(source.email || ""),
      pending_email: String(source.pending_email || ""),
      pending_email_status: String(source.pending_email_status || ""),
      full_name: String(source.full_name || ""),
      avatar_url: String(source.avatar_url || ""),
      is_admin: Boolean(source.is_admin),
      is_professor: Boolean(source.is_professor),
      email_verified: Boolean(source.email_verified),
    };
  }

  function dedupeStrings(items, limit = Infinity) {
    const seen = new Set();
    const out = [];
    (Array.isArray(items) ? items : []).forEach((item) => {
      const value = String(item || "").trim();
      if (!value) return;
      const key = value.toLowerCase();
      if (seen.has(key)) return;
      seen.add(key);
      out.push(value);
    });
    return out.slice(0, limit);
  }

  function normalizeExtendedProfile(data) {
    const source = data && typeof data === "object" ? data : {};
    return {
      headline: String(source.headline || "").trim(),
      about: String(source.about || "").trim(),
      preferred_role: String(source.preferred_role || "").trim(),
      semester: String(source.semester || "").trim(),
      availability: String(source.availability || "").trim(),
      goals: String(source.goals || "").trim(),
      github: String(source.github || source.github_url || "").trim(),
      telegram: String(source.telegram || "").trim(),
      portfolio: String(source.portfolio || source.portfolio_url || "").trim(),
      stacks: dedupeStrings(source.stacks, MAX_STACKS),
      interests: dedupeStrings(source.interests),
      updated_at: String(source.updated_at || "").trim(),
    };
  }

  function cloneExtendedProfile(data) {
    return normalizeExtendedProfile(data);
  }

  function linkCount(ext) {
    return [ext.github, ext.telegram, ext.portfolio].filter(Boolean).length;
  }

  function completionPercent(profile, ext) {
    const checks = [
      Boolean(profile && profile.full_name),
      Boolean(profile && profile.avatar_url),
      Boolean(ext.headline),
      Boolean(ext.about),
      Boolean(ext.preferred_role),
      Boolean(ext.semester),
      Boolean(ext.availability),
      Boolean(ext.goals),
      ext.stacks.length > 0,
      ext.interests.length > 0,
      linkCount(ext) > 0,
    ];
    const done = checks.filter(Boolean).length;
    return Math.round((done / checks.length) * 100);
  }

  function profileRoleMeta(profile) {
    if (profile && profile.is_admin) {
      return {
        workspace: "Admin Console",
        label: "Администратор",
        fallbackName: "Администратор",
      };
    }
    if (profile && profile.is_professor) {
      return {
        workspace: "Professor Workspace",
        label: "Преподаватель",
        fallbackName: "Преподаватель",
      };
    }
    return {
      workspace: "Student Workspace",
      label: "Студент",
      fallbackName: "Студент",
    };
  }

  function request(method, url, body) {
    return auth.requestJSON(url, {
      method,
      headers: body !== undefined && !(body instanceof FormData) ? { "Content-Type": "application/json" } : undefined,
      body: body === undefined ? undefined : body instanceof FormData ? body : JSON.stringify(body),
    }).then(({ resp, data }) => {
      if (!resp.ok) {
        const err = new Error(data && typeof data === "object" && data.error ? String(data.error) : `request failed (${resp.status})`);
        err.status = resp.status;
        throw err;
      }
      return data;
    });
  }

  function bindLogout() {
    if (!ui.logoutBtn || ui.logoutBtn.dataset.bound === "1") return;
    ui.logoutBtn.dataset.bound = "1";
    ui.logoutBtn.addEventListener("click", () => {
      auth.logout();
    });
  }

  function syncSidebar() {
    if (roleSidebar && typeof roleSidebar.renderSidebar === "function" && ui.sidebarHost) {
      roleSidebar.renderSidebar(ui.sidebarHost, {
        profile: state.viewer,
        role: "auto",
        active: "profile",
        adminViewMode: "links",
      });
    }
    ui.logoutBtn = document.getElementById("logoutBtn");
    bindLogout();
  }

  function fillExtendedForm(ext) {
    ui.headlineInput.value = ext.headline;
    ui.aboutInput.value = ext.about;
    ui.preferredRoleSelect.value = ext.preferred_role;
    ui.semesterInput.value = ext.semester;
    ui.availabilitySelect.value = ext.availability;
    ui.goalsInput.value = ext.goals;
    ui.githubInput.value = ext.github;
    ui.telegramInput.value = ext.telegram;
    ui.portfolioInput.value = ext.portfolio;

    const selected = new Set(ext.interests);
    ui.interestCheckboxes.forEach((checkbox) => {
      checkbox.checked = selected.has(String(checkbox.value || ""));
    });
  }

  function fillFormFromSaved() {
    if (!state.profile) return;
    ui.fullNameInput.value = state.profile.full_name || "";
    ui.emailInput.value = state.profile.email || "";
    ui.departmentInput.value = state.profile.department_code || "—";
    ui.groupInput.value = state.profile.group_code || "—";
    fillExtendedForm(state.extended);
  }

  function currentFullName() {
    const draft = String(ui.fullNameInput.value || "").trim();
    const meta = profileRoleMeta(state.profile);
    return draft || state.profile.full_name || state.profile.email || meta.fallbackName;
  }

  function currentExtendedProfile() {
    return state.isEditMode ? state.extended : state.savedExtended;
  }

  function renderStacks(stacks) {
    ui.stackList.innerHTML = "";
    stacks.forEach((stack) => {
      const chip = document.createElement("span");
      chip.className = "stack-chip";
      if (state.isOwnProfile && state.isEditMode) {
        chip.innerHTML = `${escapeHTML(stack)}<button type="button" aria-label="Удалить ${escapeHTML(stack)}" data-remove-stack="${escapeHTML(stack)}">×</button>`;
      } else {
        chip.textContent = stack;
      }
      ui.stackList.appendChild(chip);
    });
  }

  function renderStackPreview(stacks) {
    ui.heroStackPreview.innerHTML = "";
    stacks.slice(0, 6).forEach((stack) => {
      const chip = document.createElement("span");
      chip.className = "hero-stack-chip";
      chip.textContent = stack;
      ui.heroStackPreview.appendChild(chip);
    });
  }

  function renderInterestState(ext) {
    let selectedCount = 0;
    ui.interestCheckboxes.forEach((checkbox) => {
      const option = checkbox.closest(".interest-option");
      if (!option) return;
      const checked = Boolean(checkbox.checked);
      option.classList.toggle("is-selected", checked);
      if (checked) selectedCount += 1;
    });
    if (ui.interestGrid) {
      ui.interestGrid.classList.toggle("is-empty", !ext.interests.length);
    }
    return selectedCount;
  }

  function setHidden(el, hidden) {
    if (el) {
      el.hidden = Boolean(hidden);
    }
  }

  function setFieldHidden(input, hidden) {
    if (!input) return;
    const field = input.closest(".field");
    if (field) field.hidden = Boolean(hidden);
  }

  function syncViewVisibility(profile, ext) {
    const isViewMode = !state.isEditMode;
    const hasStacks = ext.stacks.length > 0;
    const hasLinks = linkCount(ext) > 0;
    const hasInterests = ext.interests.length > 0;
    const hasAvailability = Boolean(ext.availability);

    setHidden(ui.heroHeadline, isViewMode && !ext.headline);
    setHidden(ui.heroDepartment, isViewMode && !profile.department_code);
    setHidden(ui.heroGroup, isViewMode && !profile.group_code);
    setHidden(ui.heroStackPreview, isViewMode && !hasStacks);

    setFieldHidden(ui.headlineInput, isViewMode && !ext.headline);
    setFieldHidden(ui.aboutInput, isViewMode && !ext.about);
    setFieldHidden(ui.preferredRoleSelect, isViewMode && !ext.preferred_role);
    setFieldHidden(ui.semesterInput, isViewMode && !ext.semester);
    setFieldHidden(ui.availabilitySelect, isViewMode && !ext.availability);
    setFieldHidden(ui.goalsInput, isViewMode && !ext.goals);

    setFieldHidden(ui.githubInput, isViewMode && !ext.github);
    setFieldHidden(ui.telegramInput, isViewMode && !ext.telegram);
    setFieldHidden(ui.portfolioInput, isViewMode && !ext.portfolio);

    setFieldHidden(ui.departmentInput, isViewMode && !profile.department_code);
    setFieldHidden(ui.groupInput, isViewMode && !profile.group_code);

    setHidden(ui.stackPanel, isViewMode && !hasStacks);
    setHidden(ui.linksPanel, isViewMode && !hasLinks);
    setHidden(ui.interestsPanel, isViewMode && !hasInterests);

    setHidden(ui.stackStatCard, isViewMode && !hasStacks);
    setHidden(ui.interestStatCard, isViewMode && !hasInterests);
    setHidden(ui.linksStatCard, isViewMode && !hasLinks);
    setHidden(ui.availabilityStatCard, isViewMode && !hasAvailability);
    setHidden(ui.statsSection, isViewMode && !hasStacks && !hasInterests && !hasLinks && !hasAvailability);
  }

  function renderSummary() {
    if (!state.profile) return;

    const profile = state.profile;
    const ext = currentExtendedProfile();
    const meta = profileRoleMeta(profile);
    const name = currentFullName();
    const completion = completionPercent({ ...profile, full_name: name }, ext);
    const headline = ext.headline || (state.isEditMode ? "Добавьте специализацию" : "");

    ui.profileWorkspaceBadge.textContent = meta.workspace;
    ui.heroName.textContent = name;
    ui.heroHeadline.textContent = headline;
    ui.heroEmail.textContent = profile.email || "user@idsai.dev";

    // Update department and group text (inside the span child of gh-meta-item)
    if (ui.heroDepartment) {
      const span = ui.heroDepartment.querySelectorAll("span");
      if (span.length > 1) span[1].textContent = `Кафедра: ${profile.department_code || "—"}`;
    }
    if (ui.heroGroup) {
      const span = ui.heroGroup.querySelectorAll("span");
      if (span.length > 1) span[1].textContent = `Группа: ${profile.group_code || "—"}`;
    }

    ui.profileCompletionBadge.textContent = `Профиль ${completion}%`;

    ui.stackCountStat.textContent = String(ext.stacks.length);
    ui.interestCountStat.textContent = String(renderInterestState(ext));
    ui.linksCountStat.textContent = String(linkCount(ext));
    ui.availabilityStat.textContent = ext.availability || "Не указан";

    renderStackPreview(ext.stacks);
    renderStacks(ext.stacks);
    renderAvatar(ui.profileAvatarPreview, initials(name, profile.email), profile.avatar_url);
    syncViewVisibility(profile, ext);

    document.title = `${name} | IDSAI Corp. Profile`;
  }

  function setFieldsEditable(editable) {
    [
      ui.fullNameInput,
      ui.headlineInput,
      ui.aboutInput,
      ui.preferredRoleSelect,
      ui.semesterInput,
      ui.availabilitySelect,
      ui.goalsInput,
      ui.stackInput,
      ui.githubInput,
      ui.telegramInput,
      ui.portfolioInput,
    ].forEach((field) => {
      if (!field) return;
      field.disabled = !editable;
    });

    if (ui.addStackBtn) ui.addStackBtn.disabled = !editable;
    ui.stackSuggestionBtns.forEach((button) => {
      button.disabled = !editable;
    });
    ui.interestCheckboxes.forEach((checkbox) => {
      checkbox.disabled = !editable;
    });
  }

  function updateActionState() {
    const canEdit = state.isOwnProfile;
    const showEdit = canEdit && !state.isEditMode;
    const showEditControls = canEdit && state.isEditMode;

    ui.editProfileBtn.hidden = !showEdit;
    ui.cancelEditBtn.hidden = !showEditControls;
    ui.uploadAvatarBtn.hidden = !showEditControls;
    ui.removeAvatarBtn.hidden = !showEditControls;
    if (ui.openSettingsBtn) ui.openSettingsBtn.hidden = !showEditControls;
    ui.saveProfileBtn.hidden = !showEditControls;
    if (ui.profileAvatarNote) {
      ui.profileAvatarNote.hidden = !showEditControls;
    }

    if (ui.headerProfileAction) {
      if (state.isOwnProfile) {
        ui.headerProfileAction.hidden = false;
        ui.headerProfileAction.textContent = "Аккаунт и безопасность";
        ui.headerProfileAction.href = "/dev/settings";
      } else {
        ui.headerProfileAction.hidden = false;
        ui.headerProfileAction.textContent = "Мой профиль";
        ui.headerProfileAction.href = "/dev/profile";
      }
    }
  }

  function setMode(editMode) {
    state.isEditMode = Boolean(editMode && state.isOwnProfile);
    document.body.classList.toggle("profile-mode-view", !state.isEditMode);
    document.body.classList.toggle("profile-mode-edit", state.isEditMode);
    document.body.classList.toggle("profile-viewing-other", !state.isOwnProfile);
    setFieldsEditable(state.isEditMode);
    updateActionState();
    renderSummary();
  }

  function applyViewerProfile(data) {
    state.viewer = baseProfile(data);
    auth.persistProfile(state.viewer);
    syncSidebar();
  }

  function applyProfileData(data) {
    state.profile = baseProfile(data);
    state.savedExtended = cloneExtendedProfile(data);
    state.extended = cloneExtendedProfile(data);
    fillFormFromSaved();
    renderSummary();

    if (state.isOwnProfile) {
      applyViewerProfile(data);
    }
  }

  function collectExtendedProfile() {
    return normalizeExtendedProfile({
      headline: ui.headlineInput.value,
      about: ui.aboutInput.value,
      preferred_role: ui.preferredRoleSelect.value,
      semester: ui.semesterInput.value,
      availability: ui.availabilitySelect.value,
      goals: ui.goalsInput.value,
      github: ui.githubInput.value,
      telegram: ui.telegramInput.value,
      portfolio: ui.portfolioInput.value,
      stacks: state.extended.stacks,
      interests: ui.interestCheckboxes
        .filter((checkbox) => checkbox.checked)
        .map((checkbox) => String(checkbox.value || "")),
      updated_at: state.extended.updated_at,
    });
  }

  function updateDraftFromFields() {
    if (!state.isOwnProfile || !state.isEditMode) return;
    state.extended = collectExtendedProfile();
    renderSummary();
  }

  function resetDraftToSaved() {
    state.extended = cloneExtendedProfile(state.savedExtended);
    fillFormFromSaved();
    renderSummary();
  }

  function captureDraft() {
    return {
      fullName: String(ui.fullNameInput.value || ""),
      extended: cloneExtendedProfile(collectExtendedProfile()),
    };
  }

  function restoreDraft(draft) {
    if (!draft) return;
    state.extended = cloneExtendedProfile(draft.extended);
    fillFormFromSaved();
    ui.fullNameInput.value = draft.fullName;
    renderSummary();
  }

  function addStack(rawValue) {
    if (!state.isOwnProfile || !state.isEditMode) return;
    const value = String(rawValue || "").trim();
    if (!value) return;
    state.extended = {
      ...state.extended,
      stacks: dedupeStrings([...state.extended.stacks, value], MAX_STACKS),
    };
    ui.stackInput.value = "";
    renderSummary();
  }

  function removeStack(rawValue) {
    if (!state.isOwnProfile || !state.isEditMode) return;
    const target = String(rawValue || "").trim().toLowerCase();
    state.extended = {
      ...state.extended,
      stacks: state.extended.stacks.filter((item) => String(item || "").trim().toLowerCase() !== target),
    };
    renderSummary();
  }

  async function saveProfile() {
    if (!state.isOwnProfile) return;

    const fullName = String(ui.fullNameInput.value || "").trim();
    if (fullName.length < 2) {
      setStatus(ui.profileStatus, "Введите корректное имя (минимум 2 символа).", "err");
      return;
    }

    state.extended = collectExtendedProfile();

    const payload = {
      full_name: fullName,
      headline: state.extended.headline,
      about: state.extended.about,
      preferred_role: state.extended.preferred_role,
      semester: state.extended.semester,
      availability: state.extended.availability,
      goals: state.extended.goals,
      github_url: state.extended.github,
      telegram: state.extended.telegram,
      portfolio_url: state.extended.portfolio,
      stacks: state.extended.stacks,
      interests: state.extended.interests,
    };

    setLoading(ui.saveProfileBtn, true, "Сохраняем...");
    try {
      const data = await request("PATCH", "/v2/auth/settings/profile", payload);
      applyProfileData(data);
      setMode(false);
      setStatus(ui.profileStatus, "Профиль обновлен.", "ok");
    } catch (err) {
      setStatus(ui.profileStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.saveProfileBtn, false, "Сохранить профиль");
    }
  }

  async function uploadAvatar(file) {
    if (!state.isOwnProfile || !state.isEditMode || !file) return;
    const draft = captureDraft();

    const allowed = new Set(["image/jpeg", "image/png", "image/webp"]);
    if (!allowed.has(String(file.type || "").toLowerCase())) {
      setStatus(ui.avatarStatus, "Поддерживаются JPG, PNG и WEBP.", "err");
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
      applyProfileData(data);
      restoreDraft(draft);
      setStatus(ui.avatarStatus, "Аватар обновлен.", "ok");
    } catch (err) {
      setStatus(ui.avatarStatus, err.message || String(err), "err");
    } finally {
      if (ui.avatarInput) ui.avatarInput.value = "";
      setLoading(ui.uploadAvatarBtn, false, "Обновить аватар");
    }
  }

  async function removeAvatar() {
    if (!state.isOwnProfile || !state.isEditMode) return;
    const draft = captureDraft();

    setLoading(ui.removeAvatarBtn, true, "Удаляем...");
    try {
      const data = await request("DELETE", "/v2/auth/settings/avatar");
      applyProfileData(data);
      restoreDraft(draft);
      setStatus(ui.avatarStatus, "Аватар удален.", "ok");
    } catch (err) {
      setStatus(ui.avatarStatus, err.message || String(err), "err");
    } finally {
      setLoading(ui.removeAvatarBtn, false, "Удалить аватар");
    }
  }

  function wireEvents() {
    [
      ui.fullNameInput,
      ui.headlineInput,
      ui.aboutInput,
      ui.preferredRoleSelect,
      ui.semesterInput,
      ui.availabilitySelect,
      ui.goalsInput,
      ui.githubInput,
      ui.telegramInput,
      ui.portfolioInput,
    ].forEach((field) => {
      if (!field) return;
      field.addEventListener("input", updateDraftFromFields);
      field.addEventListener("change", updateDraftFromFields);
    });

    ui.interestCheckboxes.forEach((checkbox) => {
      checkbox.addEventListener("change", updateDraftFromFields);
    });

    ui.editProfileBtn.addEventListener("click", () => {
      resetDraftToSaved();
      setStatus(ui.profileStatus, "", "");
      setStatus(ui.avatarStatus, "", "");
      setMode(true);
      ui.fullNameInput.focus();
    });

    ui.cancelEditBtn.addEventListener("click", () => {
      resetDraftToSaved();
      setStatus(ui.profileStatus, "", "");
      setStatus(ui.avatarStatus, "", "");
      setMode(false);
    });

    if (ui.addStackBtn) {
      ui.addStackBtn.addEventListener("click", () => {
        addStack(ui.stackInput.value);
      });
    }

    if (ui.stackInput) {
      ui.stackInput.addEventListener("keydown", (event) => {
        if (event.key !== "Enter") return;
        event.preventDefault();
        addStack(ui.stackInput.value);
      });
    }

    ui.stackSuggestionBtns.forEach((button) => {
      button.addEventListener("click", () => {
        addStack(button.dataset.stackSuggestion || "");
      });
    });

    ui.stackList.addEventListener("click", (event) => {
      const button = event.target.closest("[data-remove-stack]");
      if (!button) return;
      removeStack(button.dataset.removeStack || "");
    });

    ui.saveProfileBtn.addEventListener("click", () => {
      void saveProfile();
    });

    ui.uploadAvatarBtn.addEventListener("click", () => {
      if (!state.isOwnProfile || !state.isEditMode) return;
      ui.avatarInput.click();
    });

    ui.avatarInput.addEventListener("change", () => {
      const file = ui.avatarInput.files && ui.avatarInput.files[0] ? ui.avatarInput.files[0] : null;
      if (!file) return;
      void uploadAvatar(file);
    });

    ui.removeAvatarBtn.addEventListener("click", () => {
      void removeAvatar();
    });
  }

  function requestedProfileID() {
    const params = new URLSearchParams(window.location.search || "");
    return String(params.get("user_id") || "").trim();
  }

  async function loadTargetProfile(userID) {
    return request("GET", `/v2/auth/profiles/${encodeURIComponent(userID)}`);
  }

  async function bootstrap() {
    const viewer = await auth.ensureSession(undefined);
    if (!viewer) return;

    applyViewerProfile(viewer);

    const requestedID = requestedProfileID();
    state.isOwnProfile = !requestedID || requestedID === viewer.sub;
    const targetUserID = state.isOwnProfile ? viewer.sub : requestedID;

    try {
      const data = await loadTargetProfile(targetUserID);
      applyProfileData(data);
      setMode(false);
    } catch (err) {
      if (err && err.status === 404 && auth && typeof auth.redirectToNotFound === "function") {
        auth.redirectToNotFound();
        return;
      }
      setStatus(ui.profileStatus, `Не удалось загрузить профиль: ${err.message || String(err)}`, "err");
    }
  }

  wireEvents();
  bootstrap();
})();
