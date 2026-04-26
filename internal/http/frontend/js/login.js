(() => {
  const auth = window.IDSAIAuth;
  const statusEl = document.getElementById("status");
  const respEl = document.getElementById("resp");
  const tabLoginEl = document.getElementById("tabLogin");
  const tabRegisterEl = document.getElementById("tabRegister");
  const panelLoginEl = document.getElementById("panelLogin");
  const panelRegisterEl = document.getElementById("panelRegister");
  const forgotPasswordBtn = document.getElementById("forgotPasswordBtn");
  const rememberMeEl = document.getElementById("rememberMe");
  const regEducationTypeButtons = Array.from(document.querySelectorAll("[data-education-type]"));
  const regUniversityFieldsEl = document.getElementById("regUniversityFields");
  const regSchoolFieldsEl = document.getElementById("regSchoolFields");
  const regInstitutionEl = document.getElementById("regInstitution");
  const regInstitutionSuggestionsEl = document.getElementById("regInstitutionSuggestions");
  const regInstitutionNoteEl = document.getElementById("regInstitutionNote");
  const regFacultyEl = document.getElementById("regFaculty");
  const regDepartmentEl = document.getElementById("regDepartment");
  const regGroupEl = document.getElementById("regGroup");
  const regGroupPreviewEl = document.getElementById("regGroupPreview");
  const regSchoolClassEl = document.getElementById("regSchoolClass");
  const registrationState = {
    faculties: [],
    departments: [],
    educationType: "UNIVERSITY",
    institutionSearchTimer: 0,
    institutionRequestID: 0,
    institutionResults: [],
    institutionActiveIndex: -1,
    institutionSelection: null,
  };
  const PASSWORD_ICON_SHOW = "/dev/static/assets/icon-eye.svg";
  const PASSWORD_ICON_HIDE = "/dev/static/assets/icon-eye-slash.svg";
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

  function escapeRegExp(value) {
    return String(value || "").replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  function highlightInstitutionMatch(value, query) {
    const text = String(value || "");
    const needle = String(query || "").trim();
    if (!text || !needle) {
      return escapeHTML(text);
    }

    const pattern = new RegExp(`(${escapeRegExp(needle)})`, "i");
    const parts = text.split(pattern);
    if (parts.length < 3) {
      return escapeHTML(text);
    }

    return parts
      .map((part, index) => {
        if (index % 2 === 1) {
          return `<strong>${escapeHTML(part)}</strong>`;
        }
        return escapeHTML(part);
      })
      .join("");
  }

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
    statusEl.className = "status " + (ok ? "ok" : "err");
    statusEl.innerHTML = msg
      ? `<span class="status__icon" aria-hidden="true">${STATUS_ICON_HTML}</span><span class="status__copy">${escapeHTML(msg)}</span>`
      : "";
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

  function syncPasswordToggle(button, input) {
    if (!(button instanceof HTMLButtonElement) || !(input instanceof HTMLInputElement)) {
      return;
    }
    const isVisible = input.type === "text";
    const icon = button.querySelector(".password-toggle__icon");
    button.classList.toggle("is-visible", isVisible);
    button.setAttribute("aria-pressed", isVisible ? "true" : "false");
    button.setAttribute("aria-label", isVisible ? "Скрыть пароль" : "Показать пароль");
    if (icon instanceof HTMLImageElement) {
      icon.src = isVisible ? PASSWORD_ICON_HIDE : PASSWORD_ICON_SHOW;
    }
  }

  function initPasswordToggles() {
    document.querySelectorAll("[data-password-toggle]").forEach((button) => {
      if (!(button instanceof HTMLButtonElement)) {
        return;
      }
      const targetID = String(button.dataset.target || "").trim();
      const input = targetID ? document.getElementById(targetID) : null;
      if (!(input instanceof HTMLInputElement)) {
        return;
      }
      syncPasswordToggle(button, input);
      button.addEventListener("click", () => {
        input.type = input.type === "password" ? "text" : "password";
        syncPasswordToggle(button, input);
        input.focus({ preventScroll: true });
        const caret = input.value.length;
        try {
          input.setSelectionRange(caret, caret);
        } catch (_) {}
      });
    });
  }

  function targetByProfile(profile) {
    return auth.targetByProfile(profile);
  }

  function selectedEducationType() {
    return registrationState.educationType === "SCHOOL" ? "SCHOOL" : "UNIVERSITY";
  }

  function isSchoolClassValid(value) {
    return /^(?:[1-9]|1[0-2])(?:[A-Za-zА-Яа-яЁё])?$/.test(String(value || "").trim());
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

  function institutionKindLabel() {
    return selectedEducationType() === "SCHOOL" ? "школу" : "вуз";
  }

  function setInstitutionNote(message, tone = "") {
    if (!regInstitutionNoteEl) return;
    regInstitutionNoteEl.textContent = message;
    regInstitutionNoteEl.dataset.tone = tone;
  }

  function clearInstitutionSuggestions() {
    registrationState.institutionResults = [];
    registrationState.institutionActiveIndex = -1;
    if (!regInstitutionSuggestionsEl) return;
    regInstitutionSuggestionsEl.hidden = true;
    regInstitutionSuggestionsEl.innerHTML = "";
  }

  function clearInstitutionSelection(options = {}) {
    registrationState.institutionSelection = null;
    if (!options.keepInput && regInstitutionEl) {
      regInstitutionEl.value = "";
    }
    refreshFacultyOptions();
  }

  function selectedInstitutionPayload() {
    const selected = registrationState.institutionSelection;
    if (selected && selected.name) {
      return {
        institution_provider: String(selected.provider || ""),
        institution_external_id: String(selected.external_id || ""),
        institution_name: String(selected.name || ""),
        institution_address: String(selected.address || ""),
      };
    }
    const manualName = String(regInstitutionEl?.value || "").trim();
    return {
      institution_provider: "",
      institution_external_id: "",
      institution_name: manualName,
      institution_address: "",
    };
  }

  function institutionOptionHTML(item, active) {
    const query = String(regInstitutionEl?.value || "").trim();
    const name = highlightInstitutionMatch(item && item.name, query);
    const address = escapeHTML(item && item.address);
    return `
      <button class="institution-suggestion${active ? " is-active" : ""}" type="button" role="option" aria-selected="${active ? "true" : "false"}">
        <svg class="institution-suggestion__icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z" fill="currentColor"/>
        </svg>
        <span class="institution-suggestion__body">
          <span class="institution-suggestion__line">
            <span class="institution-suggestion__name">${name}</span>
            ${address ? `<span class="institution-suggestion__address">${address}</span>` : ""}
          </span>
        </span>
      </button>
    `;
  }

  function renderInstitutionSuggestions() {
    if (!regInstitutionSuggestionsEl) return;
    const items = Array.isArray(registrationState.institutionResults) ? registrationState.institutionResults : [];
    if (!items.length) {
      clearInstitutionSuggestions();
      return;
    }

    regInstitutionSuggestionsEl.hidden = false;
    regInstitutionSuggestionsEl.innerHTML = items
      .map((item, index) => institutionOptionHTML(item, index === registrationState.institutionActiveIndex))
      .join("");

    Array.from(regInstitutionSuggestionsEl.querySelectorAll(".institution-suggestion")).forEach((button, index) => {
      button.addEventListener("mousedown", (event) => {
        event.preventDefault();
      });
      button.addEventListener("click", () => {
        const item = registrationState.institutionResults[index];
        if (!item) return;
        registrationState.institutionSelection = item;
        registrationState.institutionActiveIndex = index;
        if (regInstitutionEl) {
          regInstitutionEl.value = String(item.name || "");
        }
        clearInstitutionSuggestions();
        setInstitutionNote(
          item.address
            ? `Выбрали ${institutionKindLabel()}: ${item.address}`
            : `Выбрали ${institutionKindLabel()} из подсказок.`,
          "ok",
        );
        refreshFacultyOptions();
      });
    });
  }

  async function fetchInstitutionSuggestions(query) {
    const requestID = ++registrationState.institutionRequestID;
    const params = new URLSearchParams({
      q: String(query || "").trim(),
      kind: selectedEducationType(),
    });
    const { resp, data } = await auth.requestJSON(`/v2/auth/institutions/suggest?${params.toString()}`, {
      method: "GET",
      skipAuthRefresh: true,
      skipAuthRedirect: true,
    });
    if (requestID !== registrationState.institutionRequestID) {
      return;
    }
    if (!resp.ok) {
      clearInstitutionSuggestions();
      const reason = String(data && data.error || "").trim();
      setInstitutionNote(
        reason || `Подсказки для ${institutionKindLabel()} сейчас недоступны. Название можно ввести вручную.`,
        "warning",
      );
      return;
    }
    const items = Array.isArray(data.items) ? data.items : [];
    registrationState.institutionResults = items
      .map((item) => ({
        provider: String(item.provider || ""),
        external_id: String(item.external_id || ""),
        name: String(item.name || "").trim(),
        address: String(item.address || "").trim(),
      }))
      .filter((item) => item.name);

    if (!registrationState.institutionResults.length) {
      clearInstitutionSuggestions();
      setInstitutionNote(`Не нашли подходящую запись. Можно продолжить с ручным названием ${institutionKindLabel()}.`, "warning");
      return;
    }

    registrationState.institutionActiveIndex = 0;
    renderInstitutionSuggestions();
    setInstitutionNote(`Выберите ${institutionKindLabel()} из списка или продолжайте вводить вручную.`, "");
  }

  function scheduleInstitutionSuggestions() {
    if (registrationState.institutionSearchTimer) {
      clearTimeout(registrationState.institutionSearchTimer);
    }
    const query = String(regInstitutionEl?.value || "").trim();
    if (query.length < 2) {
      registrationState.institutionRequestID += 1;
      clearInstitutionSuggestions();
      setInstitutionNote(`Начните вводить название ${institutionKindLabel()}.`, "");
      return;
    }
    registrationState.institutionSearchTimer = window.setTimeout(() => {
      fetchInstitutionSuggestions(query).catch(() => {
        clearInstitutionSuggestions();
        setInstitutionNote(`Подсказки для ${institutionKindLabel()} сейчас недоступны. Название можно ввести вручную.`, "warning");
      });
    }, 220);
  }

  // Maps institution name keywords → university key suffix used in faculty codes.
  const universityKeyMap = [
    { key: "ENU",   words: ["евразийский", "eurasian", "ену", "enu", "gumilyov", "гумилева"] },
    { key: "KAZNU", words: ["казну", "kaznu", "аль-фараби", "al-farabi", "farabi", "казахский национальный"] },
    { key: "KBTU",  words: ["кбту", "kbtu", "казахстанско-британский", "british technical"] },
    { key: "NU",    words: ["назарбаев", "nazarbayev"] },
    { key: "MUIT",  words: ["муит", "muit", "международный университет информационных", "iitu"] },
    { key: "AITU",  words: ["aitu", "астана ит", "astana it"] },
    { key: "SAT",   words: ["сатбаев", "satbayev", "казнту", "kazntu", "сатпаева"] },
    { key: "SDU",   words: ["sdu", "suleyman", "демирель", "demirel"] },
  ];

  function detectUniversityKey() {
    const sel = registrationState.institutionSelection;
    if (sel) {
      // kzuniversities provider embeds the key in external_id as "kzuni:{KEY}"
      if (String(sel.provider || "") === "kzuniversities") {
        const raw = String(sel.external_id || "");
        if (raw.startsWith("kzuni:")) return raw.slice(6);
      }
      // Fallback: match by name keywords
      const nameLower = String(sel.name || "").toLowerCase();
      for (const entry of universityKeyMap) {
        if (entry.words.some((w) => nameLower.includes(w))) return entry.key;
      }
    }
    // Manual input: try to match typed text
    const typed = String(regInstitutionEl?.value || "").toLowerCase();
    if (typed) {
      for (const entry of universityKeyMap) {
        if (entry.words.some((w) => typed.includes(w))) return entry.key;
      }
    }
    return null;
  }

  function visibleDepartments() {
    const facultyID = String(regFacultyEl?.value || "").trim();
    const list = Array.isArray(registrationState.departments) ? registrationState.departments : [];
    if (!facultyID) {
      return [];
    }
    return list.filter((item) => String(item.faculty_id || "") === facultyID);
  }

  function visibleFaculties() {
    const all = Array.isArray(registrationState.faculties) ? registrationState.faculties : [];
    const uniKey = detectUniversityKey();
    // No institution identified — return empty so dropdown stays locked
    if (!uniKey) return null;
    return all.filter((item) => {
      const code = String(item.code || "").toUpperCase();
      const lastPart = code.includes("_") ? code.split("_").pop() : code;
      return lastPart === uniKey;
    });
  }

  function setFacultyOptions(items, locked) {
    if (!regFacultyEl) return;
    regFacultyEl.innerHTML = "";
    const first = document.createElement("option");
    first.value = "";
    first.textContent = locked ? "Сначала выберите учреждение" : "Выберите факультет";
    regFacultyEl.appendChild(first);
    regFacultyEl.disabled = locked || !items || !items.length;

    if (!locked && items) {
      items.forEach((item) => {
        const id = String(item.id || "").trim();
        const name = String(item.name || "").trim();
        const code = String(item.code || "").trim().toUpperCase();
        if (!id) return;
        const opt = document.createElement("option");
        opt.value = id;
        opt.textContent = name || code;
        regFacultyEl.appendChild(opt);
      });
    }
  }

  function refreshFacultyOptions() {
    const filtered = visibleFaculties(); // null = no institution selected
    const locked = filtered === null;
    const prevFacultyID = String(regFacultyEl?.value || "").trim();

    setFacultyOptions(filtered, locked);

    // Restore previous selection if it's still in the filtered list
    if (!locked && filtered && prevFacultyID) {
      const stillValid = filtered.some((f) => String(f.id || "") === prevFacultyID);
      if (stillValid && regFacultyEl) regFacultyEl.value = prevFacultyID;
    }

    // Auto-select when exactly one faculty matches
    if (!locked && filtered && filtered.length === 1 && regFacultyEl && !regFacultyEl.value) {
      regFacultyEl.value = String(filtered[0].id || "").trim();
    }

    // Reset department whenever faculty list changes
    if (regDepartmentEl) {
      regDepartmentEl.value = "";
      setDepartmentOptions([]);
    }
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

  function setDepartmentOptions(items) {
    if (!regDepartmentEl) return;
    const list = Array.isArray(items) ? items : [];
    regDepartmentEl.innerHTML = "";
    const first = document.createElement("option");
    first.value = "";
    first.textContent = regFacultyEl && !regFacultyEl.value ? "Сначала выберите факультет" : "Выберите кафедру";
    regDepartmentEl.appendChild(first);

    list.forEach((item) => {
      const code = String(item.code || "").toUpperCase();
      const name = String(item.name || "").trim();
      const opt = document.createElement("option");
      opt.value = code;
      opt.textContent = name ? `${code} — ${name}` : code;
      regDepartmentEl.appendChild(opt);
    });

    regDepartmentEl.disabled = list.length === 0;
    if (!list.length) {
      regDepartmentEl.value = "";
    }
  }

  function syncRegistrationFields() {
    const educationType = selectedEducationType();
    const isUniversity = educationType === "UNIVERSITY";

    regEducationTypeButtons.forEach((button) => {
      if (!(button instanceof HTMLButtonElement)) {
        return;
      }
      const active = String(button.dataset.educationType || "").toUpperCase() === educationType;
      button.classList.toggle("active", active);
      button.setAttribute("aria-pressed", active ? "true" : "false");
    });

    if (regUniversityFieldsEl) {
      regUniversityFieldsEl.hidden = !isUniversity;
    }
    if (regSchoolFieldsEl) {
      regSchoolFieldsEl.hidden = isUniversity;
    }
    if (regInstitutionEl) {
      regInstitutionEl.placeholder = isUniversity
        ? "Начните вводить название вуза"
        : "Начните вводить название школы";
    }
    if (!isUniversity) {
      if (regFacultyEl) regFacultyEl.value = "";
      if (regDepartmentEl) regDepartmentEl.value = "";
      if (regGroupEl) regGroupEl.value = "";
      updateRegistrationGroupField();
      return;
    }

    setDepartmentOptions(visibleDepartments());
    updateRegistrationGroupField();
  }

  function setRegistrationEducationType(nextType) {
    registrationState.educationType = String(nextType || "").toUpperCase() === "SCHOOL" ? "SCHOOL" : "UNIVERSITY";
    registrationState.institutionRequestID += 1;
    clearInstitutionSelection();
    clearInstitutionSuggestions();
    setInstitutionNote(`Начните вводить название ${institutionKindLabel()}.`, "");
    syncRegistrationFields();
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

  async function loadFaculties() {
    const { resp, data } = await auth.requestJSON("/v2/auth/faculties", {
      method: "GET",
      skipAuthRefresh: true,
      skipAuthRedirect: true,
    });
    if (!resp.ok) {
      throw new Error((data && data.error) || "Не удалось загрузить факультеты");
    }
    const items = Array.isArray(data.faculties) ? data.faculties : [];
    registrationState.faculties = items;
    refreshFacultyOptions();
    syncRegistrationFields();
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
    syncRegistrationFields();
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
    const rememberMe = Boolean(rememberMeEl && rememberMeEl.checked);

    const out = await callJSON("/v2/auth/login", { email, password, remember_me: rememberMe });
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
      faculty_code: out.data.faculty_code,
      department_id: out.data.department_id,
      department_code: out.data.department_code,
      group_id: out.data.group_id,
      group_code: out.data.group_code,
      group_number: out.data.group_number,
      education_type: out.data.education_type,
      school_class: out.data.school_class,
      institution_provider: out.data.institution_provider,
      institution_external_id: out.data.institution_external_id,
      institution_name: out.data.institution_name,
      institution_address: out.data.institution_address,
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
    const educationType = selectedEducationType();
    const institution = selectedInstitutionPayload();
    const facultyID = String(regFacultyEl?.value || "").trim();
    const department = String(regDepartmentEl?.value || "").trim().toUpperCase();
    const groupCode = buildGroupCode(department, regGroupEl?.value);
    const schoolClass = String(regSchoolClassEl?.value || "").trim().toUpperCase();
    const email = document.getElementById("regEmail").value.trim();
    const password = document.getElementById("regPassword").value;
    const password2 = document.getElementById("regPassword2").value;

    if (!email || !password) {
      setStatus("Заполните обязательные поля регистрации.", false);
      return;
    }
    if (password !== password2) {
      setStatus("Пароли не совпадают", false);
      return;
    }
    if (!institution.institution_name) {
      setStatus(`Укажите ${selectedEducationType() === "SCHOOL" ? "школу" : "вуз"}.`, false);
      return;
    }

    let payload;
    if (educationType === "SCHOOL") {
      if (!isSchoolClassValid(schoolClass)) {
        setStatus("Для школы укажите класс в формате 5, 9A или 11Б.", false);
        return;
      }
      payload = {
        email,
        password,
        full_name: fullName,
        education_type: "SCHOOL",
        school_class: schoolClass,
        ...institution,
      };
    } else {
      if (!facultyID || !department || !groupCode) {
        setStatus("Для вуза выберите факультет, кафедру и укажите группу.", false);
        return;
      }
      if (!/^[A-Z]{2,8}-\d{1,4}$/.test(groupCode)) {
        setStatus("Номер группы должен содержать от 1 до 4 цифр.", false);
        return;
      }
      payload = {
        email,
        password,
        full_name: fullName,
        education_type: "UNIVERSITY",
        faculty_id: facultyID,
        department_code: department,
        group_code: groupCode,
        ...institution,
      };
    }

    const out = await callJSON("/v2/auth/register", payload);
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

  initPasswordToggles();
  handleQueryState();
  regEducationTypeButtons.forEach((button) => {
    if (!(button instanceof HTMLButtonElement)) {
      return;
    }
    button.addEventListener("click", () => {
      setRegistrationEducationType(button.dataset.educationType || "UNIVERSITY");
    });
  });
  regFacultyEl?.addEventListener("change", () => {
    if (regDepartmentEl) {
      regDepartmentEl.value = "";
    }
    setDepartmentOptions(visibleDepartments());
    updateRegistrationGroupField();
  });
  regDepartmentEl?.addEventListener("change", () => {
    updateRegistrationGroupField();
  });
  regGroupEl?.addEventListener("input", () => {
    updateRegistrationGroupField();
  });
  regInstitutionEl?.addEventListener("input", () => {
    const typed = String(regInstitutionEl.value || "").trim();
    const selected = registrationState.institutionSelection;
    if (selected && typed !== String(selected.name || "").trim()) {
      clearInstitutionSelection({ keepInput: true });
    }
    scheduleInstitutionSuggestions();
    refreshFacultyOptions();
  });
  regInstitutionEl?.addEventListener("focus", () => {
    if (Array.isArray(registrationState.institutionResults) && registrationState.institutionResults.length) {
      renderInstitutionSuggestions();
    }
  });
  regInstitutionEl?.addEventListener("blur", () => {
    window.setTimeout(() => {
      clearInstitutionSuggestions();
    }, 200);
  });
  regInstitutionEl?.addEventListener("keydown", (event) => {
    const items = Array.isArray(registrationState.institutionResults) ? registrationState.institutionResults : [];
    if (!items.length) {
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      registrationState.institutionActiveIndex = (registrationState.institutionActiveIndex + 1 + items.length) % items.length;
      renderInstitutionSuggestions();
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      registrationState.institutionActiveIndex = (registrationState.institutionActiveIndex - 1 + items.length) % items.length;
      renderInstitutionSuggestions();
      return;
    }
    if (event.key === "Enter") {
      const next = items[registrationState.institutionActiveIndex];
      if (!next) return;
      event.preventDefault();
      registrationState.institutionSelection = next;
      regInstitutionEl.value = String(next.name || "");
      clearInstitutionSuggestions();
      setInstitutionNote(
        next.address
          ? `Выбрали ${institutionKindLabel()}: ${next.address}`
          : `Выбрали ${institutionKindLabel()} из подсказок.`,
        "ok",
      );
      return;
    }
    if (event.key === "Escape") {
      clearInstitutionSuggestions();
    }
  });

  setRegistrationEducationType(registrationState.educationType);

  Promise.all([loadFaculties(), loadDepartments()])
    .catch((e) => {
      setStatus("Не удалось загрузить данные для регистрации", false);
      showJSON(e.message || String(e));
    });

  restoreExistingSession().catch(() => {});
})();
