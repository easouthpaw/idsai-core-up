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
  let dialogLayer = null;
  let activeDialog = null;
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

  function ensureDialogStyles() {
    if (document.getElementById("idsaiAppDialogStyles")) {
      return;
    }
    const style = document.createElement("style");
    style.id = "idsaiAppDialogStyles";
    style.textContent = `
      .idsai-app-dialog-layer {
        position: fixed;
        inset: 0;
        z-index: 1600;
        pointer-events: none;
      }
      .idsai-app-dialog-backdrop {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 24px;
        background: rgba(15, 23, 42, 0.56);
        backdrop-filter: blur(8px);
        pointer-events: auto;
      }
      .idsai-app-dialog {
        width: min(520px, calc(100vw - 32px));
        max-height: calc(100vh - 48px);
        overflow: auto;
        border-radius: 24px;
        border: 1px solid rgba(148, 163, 184, 0.28);
        background:
          radial-gradient(circle at top right, rgba(191, 219, 254, 0.18), transparent 34%),
          linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 250, 252, 0.98) 100%);
        box-shadow: 0 28px 80px rgba(15, 23, 42, 0.34);
        color: #0f172a;
      }
      .idsai-app-dialog--danger {
        border-color: rgba(239, 68, 68, 0.24);
      }
      .idsai-app-dialog__header {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 16px;
        padding: 22px 22px 8px;
      }
      .idsai-app-dialog__title {
        margin: 0;
        font: 800 26px/1.12 "Manrope", "Inter", "Segoe UI", sans-serif;
        letter-spacing: -0.03em;
      }
      .idsai-app-dialog__subtitle {
        margin: 8px 0 0;
        color: #475569;
        font: 500 14px/1.55 "Inter", "Segoe UI", sans-serif;
        white-space: pre-line;
      }
      .idsai-app-dialog__close {
        width: 38px;
        height: 38px;
        flex: 0 0 auto;
        border: 1px solid rgba(148, 163, 184, 0.24);
        border-radius: 999px;
        background: rgba(255, 255, 255, 0.92);
        color: #475569;
        font: 400 22px/1 sans-serif;
        cursor: pointer;
      }
      .idsai-app-dialog__close:hover {
        background: #f8fafc;
      }
      .idsai-app-dialog__body {
        padding: 0 22px 6px;
      }
      .idsai-app-dialog__fields {
        display: grid;
        gap: 14px;
      }
      .idsai-app-dialog__field {
        display: grid;
        gap: 7px;
      }
      .idsai-app-dialog__label {
        color: #334155;
        font: 700 12px/1.2 "IBM Plex Mono", "SFMono-Regular", monospace;
        letter-spacing: 0.12em;
        text-transform: uppercase;
      }
      .idsai-app-dialog__input,
      .idsai-app-dialog__textarea {
        width: 100%;
        border: 1px solid rgba(148, 163, 184, 0.35);
        border-radius: 14px;
        background: rgba(255, 255, 255, 0.98);
        color: #0f172a;
        font: 500 15px/1.45 "Inter", "Segoe UI", sans-serif;
        transition: border-color 140ms ease, box-shadow 140ms ease, transform 140ms ease;
      }
      .idsai-app-dialog__input {
        min-height: 52px;
        padding: 0 15px;
      }
      .idsai-app-dialog__textarea {
        min-height: 120px;
        padding: 14px 15px;
        resize: vertical;
      }
      .idsai-app-dialog__input:focus,
      .idsai-app-dialog__textarea:focus {
        outline: none;
        border-color: rgba(37, 99, 235, 0.52);
        box-shadow: 0 0 0 4px rgba(191, 219, 254, 0.56);
        transform: translateY(-1px);
      }
      .idsai-app-dialog__hint {
        color: #64748b;
        font: 500 12px/1.5 "Inter", "Segoe UI", sans-serif;
      }
      .idsai-app-dialog__error {
        min-height: 20px;
        margin: 8px 0 0;
        color: #be123c;
        font: 700 13px/1.45 "Inter", "Segoe UI", sans-serif;
      }
      .idsai-app-dialog__actions {
        display: flex;
        justify-content: flex-end;
        gap: 10px;
        padding: 14px 22px 22px;
      }
      .idsai-app-dialog__button {
        min-width: 132px;
        min-height: 46px;
        border: 1px solid transparent;
        border-radius: 14px;
        padding: 0 18px;
        font: 800 14px/1 "Inter", "Segoe UI", sans-serif;
        cursor: pointer;
        transition: transform 140ms ease, filter 140ms ease, background 140ms ease;
      }
      .idsai-app-dialog__button:hover {
        transform: translateY(-1px);
      }
      .idsai-app-dialog__button--cancel {
        border-color: rgba(148, 163, 184, 0.32);
        background: rgba(255, 255, 255, 0.96);
        color: #334155;
      }
      .idsai-app-dialog__button--primary {
        background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
        color: #fff;
      }
      .idsai-app-dialog__button--danger {
        background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
      }
      .idsai-app-dialog__button:disabled {
        cursor: default;
        filter: grayscale(0.14);
        opacity: 0.72;
        transform: none;
      }
      @media (max-width: 640px) {
        .idsai-app-dialog-backdrop {
          padding: 14px;
        }
        .idsai-app-dialog {
          width: 100%;
          max-height: calc(100vh - 28px);
          border-radius: 20px;
        }
        .idsai-app-dialog__header,
        .idsai-app-dialog__body,
        .idsai-app-dialog__actions {
          padding-left: 16px;
          padding-right: 16px;
        }
        .idsai-app-dialog__title {
          font-size: 22px;
        }
        .idsai-app-dialog__actions {
          flex-direction: column-reverse;
        }
        .idsai-app-dialog__button {
          width: 100%;
        }
      }
    `;
    document.head.appendChild(style);
  }

  function ensureDialogLayer() {
    if (dialogLayer && document.body.contains(dialogLayer)) {
      return dialogLayer;
    }
    if (!document.body) {
      return null;
    }
    ensureDialogStyles();
    dialogLayer = document.getElementById("idsaiAppDialogLayer");
    if (dialogLayer) {
      return dialogLayer;
    }
    dialogLayer = document.createElement("section");
    dialogLayer.id = "idsaiAppDialogLayer";
    dialogLayer.className = "idsai-app-dialog-layer";
    document.body.appendChild(dialogLayer);
    return dialogLayer;
  }

  function finishDialog(dialog, result) {
    if (!dialog || dialog.closed) {
      return;
    }
    dialog.closed = true;
    if (activeDialog === dialog) {
      activeDialog = null;
    }
    document.removeEventListener("keydown", dialog.onKeyDown, true);
    dialog.backdrop.remove();
    const focusTarget = dialog.restoreFocus;
    if (focusTarget instanceof HTMLElement && document.contains(focusTarget)) {
      focusTarget.focus({ preventScroll: true });
    }
    dialog.resolve(result);
  }

  function collectDialogValues(fields, inputs) {
    const out = {};
    fields.forEach((field) => {
      const key = String(field.name || "").trim();
      if (!key) return;
      const el = inputs.get(key);
      if (!(el instanceof HTMLElement)) return;
      if (el instanceof HTMLInputElement && el.type === "checkbox") {
        out[key] = el.checked;
        return;
      }
      out[key] = "value" in el ? String(el.value || "") : "";
    });
    return out;
  }

  function showDialog(options = {}) {
    if (activeDialog) {
      finishDialog(activeDialog, options.form ? null : false);
    }

    const layer = ensureDialogLayer();
    if (!layer) {
      return Promise.resolve(options.form ? null : false);
    }

    const fields = Array.isArray(options.fields) ? options.fields : [];
    const wantsForm = fields.length > 0 || Boolean(options.form);
    const confirmText = String(options.confirmText || (wantsForm ? "Сохранить" : "Подтвердить")).trim();
    const cancelText = String(options.cancelText || "Отмена").trim();
    const showCancel = options.showCancel !== false;
    const restoreFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const titleID = `idsaiAppDialogTitle-${Date.now()}`;
    const descID = `idsaiAppDialogDesc-${Date.now()}`;

    return new Promise((resolve) => {
      const backdrop = document.createElement("div");
      backdrop.className = "idsai-app-dialog-backdrop";
      const formID = wantsForm ? `idsaiAppDialogForm-${Date.now()}` : "";

      const dialog = document.createElement("section");
      dialog.className = `idsai-app-dialog${options.danger ? " idsai-app-dialog--danger" : ""}`;
      dialog.setAttribute("role", "dialog");
      dialog.setAttribute("aria-modal", "true");
      dialog.setAttribute("aria-labelledby", titleID);
      dialog.setAttribute("aria-describedby", descID);

      const header = document.createElement("header");
      header.className = "idsai-app-dialog__header";

      const titleWrap = document.createElement("div");
      const titleEl = document.createElement("h3");
      titleEl.id = titleID;
      titleEl.className = "idsai-app-dialog__title";
      titleEl.textContent = String(options.title || "Подтвердите действие");
      titleWrap.appendChild(titleEl);

      const subtitleEl = document.createElement("p");
      subtitleEl.id = descID;
      subtitleEl.className = "idsai-app-dialog__subtitle";
      subtitleEl.textContent = String(options.message || "").trim();
      titleWrap.appendChild(subtitleEl);

      const closeBtn = document.createElement("button");
      closeBtn.type = "button";
      closeBtn.className = "idsai-app-dialog__close";
      closeBtn.setAttribute("aria-label", "Закрыть");
      closeBtn.textContent = "×";

      header.appendChild(titleWrap);
      header.appendChild(closeBtn);
      dialog.appendChild(header);

      const body = document.createElement(wantsForm ? "form" : "div");
      body.className = "idsai-app-dialog__body";
      if (wantsForm) {
        body.id = formID;
        body.setAttribute("novalidate", "novalidate");
      }

      const fieldInputs = new Map();
      if (fields.length > 0) {
        const fieldsWrap = document.createElement("div");
        fieldsWrap.className = "idsai-app-dialog__fields";
        fields.forEach((field) => {
          const key = String(field.name || "").trim();
          if (!key) return;

          const fieldWrap = document.createElement("label");
          fieldWrap.className = "idsai-app-dialog__field";
          fieldWrap.setAttribute("for", `idsai-field-${key}`);

          const label = document.createElement("span");
          label.className = "idsai-app-dialog__label";
          label.textContent = String(field.label || key);
          fieldWrap.appendChild(label);

          const isTextarea = String(field.type || "").toLowerCase() === "textarea";
          const input = isTextarea ? document.createElement("textarea") : document.createElement("input");
          input.id = `idsai-field-${key}`;
          input.className = isTextarea ? "idsai-app-dialog__textarea" : "idsai-app-dialog__input";
          input.name = key;
          if (!isTextarea && field.type) {
            input.type = String(field.type);
          }
          if (field.placeholder) input.placeholder = String(field.placeholder);
          if (field.autocomplete) input.autocomplete = String(field.autocomplete);
          if (field.inputmode) input.setAttribute("inputmode", String(field.inputmode));
          if (field.pattern) input.setAttribute("pattern", String(field.pattern));
          if (field.minLength !== undefined) input.minLength = Number(field.minLength);
          if (field.maxLength !== undefined) input.maxLength = Number(field.maxLength);
          if (field.rows !== undefined && input instanceof HTMLTextAreaElement) input.rows = Number(field.rows);
          if (field.required) input.required = true;
          if ("value" in field && field.value !== undefined && field.value !== null) {
            input.value = String(field.value);
          }

          fieldWrap.appendChild(input);
          if (field.hint) {
            const hint = document.createElement("small");
            hint.className = "idsai-app-dialog__hint";
            hint.textContent = String(field.hint);
            fieldWrap.appendChild(hint);
          }
          fieldsWrap.appendChild(fieldWrap);
          fieldInputs.set(key, input);
        });
        body.appendChild(fieldsWrap);
      }

      const errorEl = document.createElement("p");
      errorEl.className = "idsai-app-dialog__error";
      errorEl.hidden = true;
      body.appendChild(errorEl);
      dialog.appendChild(body);

      const actions = document.createElement("footer");
      actions.className = "idsai-app-dialog__actions";

      const cancelBtn = document.createElement("button");
      cancelBtn.type = "button";
      cancelBtn.className = "idsai-app-dialog__button idsai-app-dialog__button--cancel";
      cancelBtn.textContent = cancelText || "Отмена";

      const confirmBtn = document.createElement("button");
      confirmBtn.type = wantsForm ? "submit" : "button";
      confirmBtn.className = `idsai-app-dialog__button idsai-app-dialog__button--primary${options.danger ? " idsai-app-dialog__button--danger" : ""}`;
      confirmBtn.textContent = confirmText || "Подтвердить";
      if (wantsForm) {
        confirmBtn.setAttribute("form", formID);
      }

      if (showCancel) {
        actions.appendChild(cancelBtn);
      }
      actions.appendChild(confirmBtn);
      dialog.appendChild(actions);

      backdrop.appendChild(dialog);
      layer.appendChild(backdrop);

      const dialogState = {
        backdrop,
        resolve,
        restoreFocus,
        closed: false,
        onKeyDown(event) {
          if (event.key === "Escape") {
            event.preventDefault();
            finishDialog(dialogState, wantsForm ? null : false);
          }
        },
      };
      activeDialog = dialogState;

      const setError = (message) => {
        const text = String(message || "").trim();
        errorEl.textContent = text;
        errorEl.hidden = !text;
      };

      const submit = async () => {
        setError("");
        if (!wantsForm) {
          finishDialog(dialogState, true);
          return;
        }

        const values = collectDialogValues(fields, fieldInputs);
        if (typeof options.validate === "function") {
          const validation = await options.validate(values);
          if (typeof validation === "string" && validation.trim()) {
            setError(validation);
            return;
          }
        }
        finishDialog(dialogState, values);
      };

      closeBtn.addEventListener("click", () => finishDialog(dialogState, wantsForm ? null : false));
      cancelBtn.addEventListener("click", () => finishDialog(dialogState, wantsForm ? null : false));
      confirmBtn.addEventListener("click", (event) => {
        if (!wantsForm) {
          event.preventDefault();
          void submit();
        }
      });
      if (wantsForm) {
        body.addEventListener("submit", (event) => {
          event.preventDefault();
          void submit();
        });
      }
      backdrop.addEventListener("click", (event) => {
        if (event.target === backdrop) {
          finishDialog(dialogState, wantsForm ? null : false);
        }
      });
      document.addEventListener("keydown", dialogState.onKeyDown, true);

      const firstField = fields.length > 0 ? fieldInputs.get(String(fields[0].name || "").trim()) : null;
      window.setTimeout(() => {
        if (firstField && "focus" in firstField) {
          firstField.focus({ preventScroll: true });
          if ("select" in firstField && typeof firstField.select === "function") {
            firstField.select();
          }
          return;
        }
        confirmBtn.focus({ preventScroll: true });
      }, 0);
    });
  }

  async function showConfirmDialog(options = {}) {
    return Boolean(await showDialog({ ...options, form: false }));
  }

  async function showFormDialog(options = {}) {
    const result = await showDialog({ ...options, form: true });
    return result && typeof result === "object" ? result : null;
  }

  async function showMessageDialog(options = {}) {
    await showDialog({
      title: options.title || "Сообщение",
      message: options.message || "",
      confirmText: options.confirmText || "Понятно",
      showCancel: false,
      danger: Boolean(options.danger),
      form: false,
    });
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
    showConfirmDialog,
    showFormDialog,
    showMessageDialog,
    logout,
    targetByProfile,
  };
})();
