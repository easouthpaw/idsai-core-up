(() => {
  const auth = window.IDSAIAuth;
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";
  const GUIDE_STORAGE_KEY = "idsai_professor_criteria_guide_hidden";
  const EDITABLE_REVIEW_STATUS = new Set(["ACCEPTED"]);

  const ui = {
    profAvatar: document.getElementById("profAvatar"),
    profName: document.getElementById("profName"),
    profEmail: document.getElementById("profEmail"),
    logoutBtn: document.getElementById("logoutBtn"),
    projectSelect: document.getElementById("projectSelect"),
    projectStatusView: document.getElementById("projectStatusView"),
    criteriaWeightMeta: document.getElementById("criteriaWeightMeta"),
    criteriaList: document.getElementById("criteriaList"),
    guidePanel: document.getElementById("profGuidePanel"),
    guideSteps: document.getElementById("profGuideSteps"),
    guideToggleBtn: document.getElementById("profGuideToggleBtn"),
    guideRestoreBtn: document.getElementById("profGuideRestoreBtn"),
    criterionForm: document.getElementById("criterionForm"),
    openComposerBtn: document.getElementById("openComposerBtn"),
    cancelComposerBtn: document.getElementById("cancelComposerBtn"),
    criterionTitleInput: document.getElementById("criterionTitleInput"),
    criterionWeightInput: document.getElementById("criterionWeightInput"),
    criterionDescInput: document.getElementById("criterionDescInput"),
    criterionSubmitBtn: document.getElementById("criterionSubmitBtn"),
    loadTemplateBtn: document.getElementById("loadTemplateBtn"),
    cancelBottomBtn: document.getElementById("cancelBottomBtn"),
    saveTemplateBtn: document.getElementById("saveTemplateBtn"),
    saveBottomBtn: document.getElementById("saveBottomBtn"),
    pageStatus: document.getElementById("pageStatus"),
  };

  const state = {
    userID: "",
    projectID: "",
    projects: [],
    criteria: [],
    composerOpen: false,
    canEditCurrent: false,
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
    return e ? e.slice(0, 2).toUpperCase() : "PR";
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

  function setStatus(message, isError) {
    if (!ui.pageStatus) return;
    ui.pageStatus.textContent = message || "";
    ui.pageStatus.classList.toggle("err", Boolean(isError));
  }

  function guideHidden() {
    try {
      return localStorage.getItem(GUIDE_STORAGE_KEY) === "1";
    } catch (_) {
      return false;
    }
  }

  function applyGuideVisibility() {
    const hidden = guideHidden();
    if (ui.guidePanel) ui.guidePanel.hidden = hidden;
    if (ui.guideToggleBtn) ui.guideToggleBtn.hidden = hidden;
    if (ui.guideRestoreBtn) ui.guideRestoreBtn.hidden = !hidden;
  }

  function setGuideHidden(hidden) {
    try {
      if (hidden) {
        localStorage.setItem(GUIDE_STORAGE_KEY, "1");
      } else {
        localStorage.removeItem(GUIDE_STORAGE_KEY);
      }
    } catch (_) {}
    applyGuideVisibility();
  }

  function setComposerOpen(open) {
    if (open && !state.canEditCurrent) return;
    state.composerOpen = Boolean(open);
    if (ui.criterionForm) {
      ui.criterionForm.hidden = !state.composerOpen;
    }
    if (ui.openComposerBtn) {
      ui.openComposerBtn.hidden = state.composerOpen;
    }
    if (state.composerOpen && ui.criterionTitleInput) {
      ui.criterionTitleInput.focus();
    }
  }

  function resetComposerForm() {
    if (ui.criterionTitleInput) ui.criterionTitleInput.value = "";
    if (ui.criterionDescInput) ui.criterionDescInput.value = "";
    if (ui.criterionWeightInput) ui.criterionWeightInput.value = "10";
  }

  function decodePayload(token) {
    const parts = String(token || "").split(".");
    if (parts.length < 2) throw new Error("invalid token");
    let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const mod = payload.length % 4;
    if (mod > 0) payload += "=".repeat(4 - mod);
    return JSON.parse(atob(payload));
  }

  function clearSession() {
    auth.clearClientState();
  }

  function ensureSession() {
    const claims = auth.getCachedProfile();
    if (!claims) {
      window.location.href = "/dev/login";
      return null;
    }
    if (claims.is_admin) {
      window.location.href = "/dev/admin";
      return null;
    }
    if (!claims.is_professor) {
      window.location.href = "/dev/projects";
      return null;
    }
    return claims;
  }

  function authHeaders(withJSON) {
    const headers = {};
    if (withJSON) headers["Content-Type"] = "application/json";
    return headers;
  }

  async function request(method, url, body) {
    const { resp, data } = await auth.requestJSON(url, {
      method,
      headers: body ? { "Content-Type": "application/json" } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!resp.ok) {
      const errText = data && data.error ? data.error : `${method} ${url} failed (${resp.status})`;
      throw new Error(errText);
    }

    return data;
  }

  function projectIDFromQuery() {
    return String(new URLSearchParams(window.location.search).get("project_id") || "").trim();
  }

  function setProjectInURL(projectID) {
    const url = new URL(window.location.href);
    if (projectID) {
      url.searchParams.set("project_id", projectID);
    } else {
      url.searchParams.delete("project_id");
    }
    window.history.replaceState({}, "", url.toString());
  }

  function bindProfile() {
    const name = localStorage.getItem(LS_STUDENT_NAME) || "Преподаватель";
    const email = localStorage.getItem(LS_STUDENT_EMAIL) || "professor@idsai.dev";
    const avatarURL = localStorage.getItem(LS_AVATAR_URL) || "";
    if (ui.profName) ui.profName.textContent = name;
    if (ui.profEmail) ui.profEmail.textContent = email;
    renderAvatar(ui.profAvatar, initials(name, email), avatarURL);
  }

  function normalizeProjects(items) {
    const list = Array.isArray(items) ? items : [];
    return list.sort((a, b) => {
      const ad = new Date(a.updated_at || a.created_at || 0).getTime();
      const bd = new Date(b.updated_at || b.created_at || 0).getTime();
      return bd - ad;
    });
  }

  function currentProject() {
    return state.projects.find((item) => String(item.id || "") === String(state.projectID || "")) || null;
  }

  function isManagedByProfessor(project) {
    const uid = String(state.userID || "");
    if (!uid || !project) return false;
    return String(project.professor_id || "") === uid || String(project.created_by || "") === uid;
  }

  function canEditProject(project) {
    const uid = String(state.userID || "");
    if (!uid || !project) return false;
    if (String(project.created_by || "") === uid) return true;
    const professorID = String(project.professor_id || "");
    const reviewStatus = String(project.professor_review_status || "NONE").toUpperCase();
    return professorID === uid && EDITABLE_REVIEW_STATUS.has(reviewStatus);
  }

  function projectStatusLabel(project) {
    if (!project) return "Сначала выберите проект";
    const uid = String(state.userID || "");
    const status = String(project.status || "DRAFT").toUpperCase();
    if (String(project.created_by || "") === uid) {
      return `${status} · владелец проекта`;
    }
    const review = String(project.professor_review_status || "NONE").toUpperCase();
    if (String(project.professor_id || "") !== uid) {
      return `${status} · нет доступа`;
    }
    if (review === "ACCEPTED") return `${status} · доступ на редактирование`;
    if (review === "PENDING") return `${status} · приглашение не принято`;
    if (review === "REJECTED") return `${status} · приглашение отклонено`;
    return `${status} · ожидание назначения`;
  }

  function syncEditorState() {
    const project = currentProject();
    state.canEditCurrent = canEditProject(project);

    if (ui.projectStatusView) {
      ui.projectStatusView.value = projectStatusLabel(project);
    }

    if (!state.canEditCurrent) {
      setComposerOpen(false);
    }
    syncWeightMeta();
    syncComposerControls();
    renderGuide();
  }

  function renderGuide() {
    if (!ui.guideSteps) return;

    const project = currentProject();
    const totalWeight = criteriaWeightTotal();
    const status = projectStatusLabel(project);
    const reviewStatus = String(project && project.professor_review_status || "NONE").toUpperCase();

    const steps = [
      {
        tone: project ? "done" : "current",
        kicker: "Шаг 1",
        title: project ? `Выбран проект: ${project.title || "Без названия"}` : "Сначала выберите проект",
        text: project
          ? `Текущий контекст: ${status}. Теперь можно проверять доступ и собирать чек-лист.`
          : "Без выбранного проекта страница остается нейтральной и ничего не редактирует сама.",
        actions: [{ act: "focus-project", label: project ? "Сменить проект" : "Выбрать проект" }],
      },
      {
        tone: !project ? "blocked" : state.canEditCurrent ? "done" : reviewStatus === "PENDING" ? "current" : "blocked",
        kicker: "Шаг 2",
        title: state.canEditCurrent ? "Режим редактирования доступен" : "Проверьте право на редактирование",
        text: !project
          ? "Сначала определитесь с проектом, и только потом проверяйте статус приглашения."
          : state.canEditCurrent
            ? "Изменения применяются сразу для команды проекта. Можно спокойно собирать финальный чек-лист."
            : reviewStatus === "PENDING"
              ? "Приглашение еще не принято. Сначала подтвердите ревью на странице заявок."
              : "Для этого проекта у вас сейчас только режим просмотра.",
        actions: !project
          ? []
          : state.canEditCurrent
            ? [{ act: "open-composer", label: "Добавить критерий" }]
            : reviewStatus === "PENDING"
              ? [{ act: "goto-reviews", label: "К заявкам" }]
              : [{ act: "open-project", label: "Открыть проект" }],
      },
      {
        tone: !project ? "blocked" : totalWeight >= 100 ? "done" : "current",
        kicker: "Шаг 3",
        title: !project ? "Чек-лист еще не собран" : `Суммарный вес: ${totalWeight} / 100`,
        text: !project
          ? "Когда проект будет выбран, здесь появится подсказка по наполнению чек-листа."
          : totalWeight >= 100
            ? "Чек-лист полностью укомплектован. Следующий преподавательский шаг можно продолжать на странице оценивания."
            : "Доведите набор критериев до нужной полноты и только потом переключайтесь в оценивание.",
        actions: !project
          ? []
          : totalWeight >= 100
            ? [{ act: "open-grading", label: "К оцениванию" }]
            : state.canEditCurrent
              ? [{ act: "open-composer", label: "Продолжить чек-лист" }]
              : [{ act: "focus-project", label: "Проверить проект" }],
      },
    ];

    ui.guideSteps.innerHTML = steps.map((step) => (
      `<article class="prof-guide-step prof-guide-step--${escapeHTML(step.tone)}">` +
        `<small>${escapeHTML(step.kicker)}</small>` +
        `<strong>${escapeHTML(step.title)}</strong>` +
        `<p>${escapeHTML(step.text)}</p>` +
        `<div class="prof-guide-step__actions">` +
          (Array.isArray(step.actions) ? step.actions.map((action) => (
            `<button class="ghost-btn" type="button" data-guide-act="${escapeHTML(action.act)}">${escapeHTML(action.label)}</button>`
          )).join("") : "") +
        `</div>` +
      `</article>`
    )).join("");
  }

  function criteriaWeightTotal() {
    return (Array.isArray(state.criteria) ? state.criteria : []).reduce((sum, item) => {
      const weight = Number.parseInt(String(item && item.weight !== undefined ? item.weight : "0"), 10);
      if (!Number.isFinite(weight) || weight <= 0) return sum;
      return sum + weight;
    }, 0);
  }

  function syncWeightMeta() {
    if (!ui.criteriaWeightMeta) return;
    const total = criteriaWeightTotal();
    const over = total > 100;
    ui.criteriaWeightMeta.textContent = `Вес ${total} / 100`;
    ui.criteriaWeightMeta.classList.toggle("alert", over);
  }

  function syncComposerControls() {
    const total = criteriaWeightTotal();
    const remains = Math.max(0, 100 - total);
    const canAdd = state.canEditCurrent && remains > 0;
    if (!canAdd && state.composerOpen) {
      setComposerOpen(false);
      resetComposerForm();
    }
    if (ui.openComposerBtn) {
      ui.openComposerBtn.disabled = !canAdd;
      ui.openComposerBtn.title = canAdd ? "" : remains <= 0 ? "Достигнут лимит суммарного веса (100)." : "Нет прав на редактирование.";
    }
    if (ui.criterionSubmitBtn) ui.criterionSubmitBtn.disabled = !canAdd;
    if (ui.saveBottomBtn) ui.saveBottomBtn.disabled = !state.canEditCurrent;
    if (ui.criterionWeightInput) {
      const max = Math.max(1, remains);
      ui.criterionWeightInput.max = String(max);
      if (Number.parseInt(ui.criterionWeightInput.value || "0", 10) > max) {
        ui.criterionWeightInput.value = String(max);
      }
    }
  }

  async function loadProjects() {
    const queryProjectID = projectIDFromQuery();

    const [mine, pub] = await Promise.all([
      request("GET", "/v2/projects/my"),
      request("GET", "/v2/projects/public"),
    ]);

    const map = new Map();
    (Array.isArray(mine) ? mine : []).forEach((item) => map.set(item.id, item));
    (Array.isArray(pub) ? pub : []).forEach((item) => map.set(item.id, item));

    if (queryProjectID && !map.has(queryProjectID)) {
      try {
        const exact = await request("GET", `/v2/projects/${queryProjectID}`);
        if (exact && exact.id) {
          map.set(exact.id, exact);
        }
      } catch (_) {}
    }

    state.projects = normalizeProjects(Array.from(map.values())).filter(isManagedByProfessor);
    const hasQuery = queryProjectID && state.projects.some((p) => String(p.id) === queryProjectID);
    if (hasQuery) {
      state.projectID = queryProjectID;
    } else if (!state.projects.some((p) => String(p.id) === String(state.projectID || ""))) {
      state.projectID = "";
    }

    renderProjectSelect();
    syncEditorState();
  }

  function renderProjectSelect() {
    if (!ui.projectSelect) return;

    if (!state.projects.length) {
      state.projectID = "";
      ui.projectSelect.innerHTML = "<option value=''>Нет проектов для настройки</option>";
      ui.projectSelect.disabled = true;
      return;
    }

    ui.projectSelect.disabled = false;
    ui.projectSelect.innerHTML =
      `<option value="">Выберите проект</option>` +
      state.projects.map((p) => {
        const selected = String(p.id) === String(state.projectID) ? "selected" : "";
        const status = String(p.status || "DRAFT").toUpperCase();
        return `<option value="${escapeHTML(p.id)}" ${selected}>${escapeHTML(p.title || "Без названия")} · ${escapeHTML(status)}</option>`;
      }).join("");
  }

  async function loadCriteria() {
    if (!state.projectID) {
      state.criteria = [];
      renderCriteria();
      syncWeightMeta();
      syncComposerControls();
      return;
    }

    try {
      const items = await request("GET", `/v2/projects/${state.projectID}/criteria`);
      state.criteria = Array.isArray(items) ? items : [];
    } catch (err) {
      const msg = String(err && err.message ? err.message : err).toLowerCase();
      if (msg.includes("forbidden")) {
        state.criteria = [];
        setStatus("Для этого проекта у вас нет доступа к просмотру критериев.", true);
        renderCriteria();
        return;
      }
      throw err;
    }

    renderCriteria();
    syncWeightMeta();
    syncComposerControls();
  }

  function renderCriteria() {
    if (!ui.criteriaList) return;

    if (!state.projectID) {
      ui.criteriaList.innerHTML =
        `<article class="criterion-row draft">` +
          `<div class="criterion-grip" aria-hidden="true">⋮⋮</div>` +
          `<div class="criterion-idx">0</div>` +
          `<div class="criterion-main">` +
            `<strong>Проект пока не выбран</strong>` +
            `<p>Выберите проект сверху, и после этого здесь появится текущий чек-лист критериев.</p>` +
          `</div>` +
          `<div class="criterion-weight">Ожидание</div>` +
        `</article>`;
      renderGuide();
      return;
    }

    if (!state.criteria.length) {
      ui.criteriaList.innerHTML =
        `<article class="criterion-row draft">` +
          `<div class="criterion-grip" aria-hidden="true">⋮⋮</div>` +
          `<div class="criterion-idx">1</div>` +
          `<div class="criterion-main">` +
            `<strong>Начните вводить новый критерий...</strong>` +
            `<p>Добавьте первый пункт проверки, чтобы сформировать чек-лист проекта.</p>` +
          `</div>` +
          `<div class="criterion-weight">Черновик</div>` +
      `</article>`;
      renderGuide();
      return;
    }

    const rows = state.criteria
      .map((item, idx) => {
        const weight = Number(item.weight || 1);
        return `
          <article class="criterion-row">
            <div class="criterion-grip" aria-hidden="true">⋮⋮</div>
            <div class="criterion-idx">${idx + 1}</div>
            <div class="criterion-main">
              <strong>${escapeHTML(item.title || "Без названия")}</strong>
              <p>${escapeHTML(item.description || "Описание отсутствует")}</p>
            </div>
            <div class="criterion-weight">Вес ${escapeHTML(weight)}</div>
          </article>
        `;
      })
      .join("");

    ui.criteriaList.innerHTML =
      rows +
      `<article class="criterion-row draft">` +
        `<div class="criterion-grip" aria-hidden="true">⋮⋮</div>` +
        `<div class="criterion-idx">${state.criteria.length + 1}</div>` +
        `<div class="criterion-main">` +
          `<strong>Добавьте следующий критерий</strong>` +
          `<p>${state.canEditCurrent ? "Нажмите кнопку ниже, чтобы открыть форму добавления." : "Редактирование станет доступно после принятия приглашения на ревью."}</p>` +
        `</div>` +
        `<div class="criterion-weight">Черновик</div>` +
      `</article>`;
    renderGuide();
  }

  async function createCriterion(event) {
    event.preventDefault();

    if (!state.projectID) {
      setStatus("Сначала выберите проект.", true);
      return;
    }
    if (!state.canEditCurrent) {
      setStatus("Редактирование недоступно: преподаватель должен быть подтверждён в проекте.", true);
      return;
    }

    const title = String(ui.criterionTitleInput.value || "").trim();
    const description = String(ui.criterionDescInput.value || "").trim();
    const weight = Number.parseInt(String(ui.criterionWeightInput.value || "10"), 10);
    const total = criteriaWeightTotal();

    if (!title) {
      setStatus("Введите название критерия.", true);
      return;
    }

    if (!Number.isFinite(weight) || weight < 1 || weight > 100) {
      setStatus("Вес должен быть числом от 1 до 100.", true);
      return;
    }
    if (total + weight > 100) {
      setStatus(`Суммарный вес критериев не должен превышать 100. Сейчас: ${total}/100, осталось: ${Math.max(0, 100 - total)}.`, true);
      return;
    }

    ui.criterionSubmitBtn.disabled = true;
    try {
      await request("POST", `/v2/projects/${state.projectID}/criteria`, {
        title,
        description,
        weight,
      });
      resetComposerForm();
      setComposerOpen(false);
      setStatus("Критерий добавлен.", false);
      await loadCriteria();
      syncComposerControls();
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      ui.criterionSubmitBtn.disabled = false;
    }
  }

  async function onProjectChange() {
    state.projectID = String(ui.projectSelect.value || "").trim();
    setProjectInURL(state.projectID);
    setComposerOpen(false);
    syncEditorState();
    await loadCriteria();
    if (!state.projectID) {
      setStatus("Сначала выберите проект для настройки критериев.", false);
      return;
    }
    if (!state.canEditCurrent) {
      setStatus("Проект выбран в режиме read-only: примите приглашение на ревью, чтобы редактировать критерии.", false);
      return;
    }
    if (criteriaWeightTotal() >= 100) {
      setStatus("Достигнут лимит суммарного веса 100. Для добавления нового критерия уменьшите существующие веса.", true);
      return;
    }
    setStatus("Готово к редактированию критериев.", false);
  }

  function handleGuideAction(act) {
    const action = String(act || "").trim();
    if (action === "focus-project" && ui.projectSelect) {
      ui.projectSelect.focus();
      return;
    }
    if (action === "goto-reviews") {
      window.location.href = "/dev/professor/reviews";
      return;
    }
    if (action === "open-project" && state.projectID) {
      window.location.href = `/dev/projects/${encodeURIComponent(state.projectID)}`;
      return;
    }
    if (action === "open-grading") {
      if (state.projectID) {
        window.location.href = `/dev/professor/grading?project_id=${encodeURIComponent(state.projectID)}`;
      } else if (ui.projectSelect) {
        ui.projectSelect.focus();
      }
      return;
    }
    if (action === "open-composer") {
      if (!state.projectID) {
        if (ui.projectSelect) ui.projectSelect.focus();
        return;
      }
      if (!state.canEditCurrent) {
        setStatus("Редактирование пока недоступно для этого проекта.", true);
        return;
      }
      setComposerOpen(true);
      if (ui.criterionTitleInput) {
        requestAnimationFrame(() => ui.criterionTitleInput.focus());
      }
    }
  }

  function attachEvents() {
    if (ui.logoutBtn && ui.logoutBtn.dataset.bound !== "1") {
      ui.logoutBtn.dataset.bound = "1";
      ui.logoutBtn.addEventListener("click", () => {
        auth.logout();
      });
    }

    if (ui.projectSelect) {
      ui.projectSelect.addEventListener("change", () => {
        onProjectChange().catch((err) => setStatus(err.message || String(err), true));
      });
    }

    if (ui.guideToggleBtn) {
      ui.guideToggleBtn.addEventListener("click", () => {
        setGuideHidden(true);
      });
    }
    if (ui.guideRestoreBtn) {
      ui.guideRestoreBtn.addEventListener("click", () => {
        setGuideHidden(false);
      });
    }
    if (ui.guideSteps) {
      ui.guideSteps.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-guide-act]");
        if (!btn) return;
        handleGuideAction(btn.getAttribute("data-guide-act") || "");
      });
    }

    if (ui.criterionForm) {
      ui.criterionForm.addEventListener("submit", createCriterion);
    }

    if (ui.openComposerBtn) {
      ui.openComposerBtn.addEventListener("click", () => setComposerOpen(true));
    }

    if (ui.cancelComposerBtn) {
      ui.cancelComposerBtn.addEventListener("click", () => {
        setComposerOpen(false);
        resetComposerForm();
      });
    }

    if (ui.loadTemplateBtn) {
      ui.loadTemplateBtn.addEventListener("click", () => {
        setStatus("Загрузка шаблонов будет добавлена в следующем обновлении.", false);
      });
    }
    if (ui.saveTemplateBtn) {
      ui.saveTemplateBtn.addEventListener("click", () => {
        setStatus("Сохранение шаблона пока недоступно в API.", false);
      });
    }
    if (ui.cancelBottomBtn) {
      ui.cancelBottomBtn.addEventListener("click", () => {
        setComposerOpen(false);
        resetComposerForm();
      });
    }
    if (ui.saveBottomBtn) {
      ui.saveBottomBtn.addEventListener("click", () => {
        if (!state.canEditCurrent) {
          setStatus("Редактирование недоступно: нет прав на этот проект.", true);
          return;
        }
        setComposerOpen(false);
        resetComposerForm();
        setStatus("Готово. Список критериев сохранён.", false);
      });
    }
  }

  async function bootstrap() {
    const claims = await auth.ensureSession("professor");
    if (!claims) return;
    state.userID = String(claims.sub || "");

    bindProfile();
    attachEvents();
    setComposerOpen(false);
    applyGuideVisibility();
    renderGuide();

    try {
      setStatus("Загрузка данных...", false);
      await loadProjects();
      setProjectInURL(state.projectID);
      await loadCriteria();
      if (!state.projects.length) {
        setStatus("Нет проектов, где вы подтверждены как преподаватель-ревьюер.", true);
      } else if (!state.projectID) {
        setStatus("Сначала выберите проект для настройки критериев.", false);
      } else if (!state.canEditCurrent) {
        setStatus("Проект выбран в режиме read-only: примите приглашение на ревью, чтобы редактировать критерии.", false);
      } else {
        setStatus("Готово к редактированию критериев.", false);
      }
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      auth.setPageLoading(false);
    }
  }

  bootstrap().catch((err) => {
    auth.setPageLoading(false);
    setStatus(err.message || String(err), true);
  });
})();
