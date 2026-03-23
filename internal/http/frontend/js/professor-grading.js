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

  const EDITABLE_STATUSES = new Set(["REVIEW", "GRADING"]);
  const FINALIZABLE_STATUSES = new Set(["GRADING"]);

  const ui = {
    profAvatar: document.getElementById("profAvatar"),
    profName: document.getElementById("profName"),
    profEmail: document.getElementById("profEmail"),
    logoutBtn: document.getElementById("logoutBtn"),
    projectTitle: document.getElementById("projectTitle"),
    projectMeta: document.getElementById("projectMeta"),
    projectStatusBadge: document.getElementById("projectStatusBadge"),
    openCriteriaBtn: document.getElementById("openCriteriaBtn"),
    gradingList: document.getElementById("gradingList"),
    summaryCoverage: document.getElementById("summaryCoverage"),
    summaryMet: document.getElementById("summaryMet"),
    summaryScore: document.getElementById("summaryScore"),
    gradingProgressFill: document.getElementById("gradingProgressFill"),
    gradingProgressText: document.getElementById("gradingProgressText"),
    publishGradingBtn: document.getElementById("publishGradingBtn"),
    saveGradingBtn: document.getElementById("saveGradingBtn"),
    pageStatus: document.getElementById("pageStatus"),
  };

  const state = {
    projectID: "",
    projects: [],
    project: null,
    criteria: [],
    grading: new Map(),
    canEdit: false,
    canPublish: false,
    isComplete: false,
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
      el.innerHTML = `<img src="${escapeHTML(url)}" alt="Avatar" loading="lazy" />`;
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

  function normalizeProjects(items) {
    const list = Array.isArray(items) ? items : [];
    return list.sort((a, b) => {
      const ad = new Date(a.updated_at || a.created_at || 0).getTime();
      const bd = new Date(b.updated_at || b.created_at || 0).getTime();
      return bd - ad;
    });
  }

  async function loadProjectList() {
    const [mine, pub] = await Promise.all([
      request("GET", "/v2/projects/my"),
      request("GET", "/v2/projects/public"),
    ]);

    const map = new Map();
    (Array.isArray(mine) ? mine : []).forEach((item) => map.set(item.id, item));
    (Array.isArray(pub) ? pub : []).forEach((item) => map.set(item.id, item));

    state.projects = normalizeProjects(Array.from(map.values()));

    const queryProjectID = projectIDFromQuery();
    const hasQuery = queryProjectID && state.projects.some((p) => String(p.id) === queryProjectID);
    if (hasQuery) {
      state.projectID = queryProjectID;
    } else if (!state.projectID && state.projects.length > 0) {
      state.projectID = String(state.projects[0].id || "");
    }
  }

  function bindProfile() {
    const name = localStorage.getItem(LS_STUDENT_NAME) || "Преподаватель";
    const email = localStorage.getItem(LS_STUDENT_EMAIL) || "professor@idsai.dev";
    const avatarURL = localStorage.getItem(LS_AVATAR_URL) || "";
    if (ui.profName) ui.profName.textContent = name;
    if (ui.profEmail) ui.profEmail.textContent = email;
    renderAvatar(ui.profAvatar, initials(name, email), avatarURL);
  }

  function statusClass(status) {
    const s = String(status || "").toUpperCase();
    if (s === "REVIEW") return "review";
    if (s === "ACTIVE") return "active";
    if (s === "RECRUITMENT") return "recruitment";
    return "default";
  }

  function ensureGradeEntry(criterionID) {
    const id = String(criterionID || "");
    if (!id) return { criterion_id: "", is_met: null, comment: "" };
    if (!state.grading.has(id)) {
      state.grading.set(id, {
        criterion_id: id,
        is_met: null,
        comment: "",
      });
    }
    return state.grading.get(id);
  }

  function renderProjectHeader() {
    const project = state.project;
    if (!project) {
      ui.projectTitle.textContent = "Проект не выбран";
      ui.projectMeta.textContent = "";
      ui.projectStatusBadge.textContent = "-";
      ui.projectStatusBadge.className = "status-pill default";
      return;
    }

    const status = String(project.status || "DRAFT").toUpperCase();
    ui.projectTitle.textContent = project.title || "Без названия";
    ui.projectMeta.textContent = `Статус: ${status} · Обновлен: ${new Date(project.updated_at || project.created_at || Date.now()).toLocaleString()}`;
    ui.projectStatusBadge.textContent = status;
    ui.projectStatusBadge.className = `status-pill ${statusClass(status)}`;
    state.canEdit = EDITABLE_STATUSES.has(status);
    state.canPublish = FINALIZABLE_STATUSES.has(status);

    const criteriaHref = state.projectID
      ? `/dev/professor/criteria?project_id=${encodeURIComponent(state.projectID)}`
      : "/dev/professor/criteria";
    ui.openCriteriaBtn.href = criteriaHref;
  }

  function renderSummary() {
    const total = state.criteria.length;

    let answered = 0;
    let met = 0;
    let weightTotal = 0;
    let weightMet = 0;

    state.criteria.forEach((criterion) => {
      const id = String(criterion.id || "");
      const grade = ensureGradeEntry(id);
      const weight = Number(criterion.weight || 0) > 0 ? Number(criterion.weight) : 1;
      weightTotal += weight;

      if (grade.is_met === true || grade.is_met === false) {
        answered += 1;
      }
      if (grade.is_met === true) {
        met += 1;
        weightMet += weight;
      }
    });

    const coverage = total > 0 ? Math.round((answered * 100) / total) : 0;
    const score = weightTotal > 0 ? Math.round((weightMet * 100) / weightTotal) : 0;
    state.isComplete = total > 0 && answered === total;

    ui.summaryCoverage.textContent = `${coverage}%`;
    ui.summaryMet.textContent = `${met}/${total}`;
    ui.summaryScore.textContent = `${score}/100`;
    if (ui.gradingProgressFill) {
      ui.gradingProgressFill.style.width = `${coverage}%`;
    }
    if (ui.gradingProgressText) {
      ui.gradingProgressText.textContent = `${coverage}%`;
    }
  }

  function renderGradingList() {
    if (!ui.gradingList) return;

    if (!state.criteria.length) {
      ui.gradingList.innerHTML = '<div class="empty-state">Критерии пока не настроены.</div>';
      ui.saveGradingBtn.disabled = true;
      state.isComplete = false;
      if (ui.publishGradingBtn) ui.publishGradingBtn.disabled = true;
      renderSummary();
      return;
    }

    ui.gradingList.innerHTML = state.criteria
      .map((criterion, idx) => {
        const id = String(criterion.id || "");
        const grade = ensureGradeEntry(id);
        const yesActive = grade.is_met === true ? "active" : "";
        const noActive = grade.is_met === false ? "active" : "";
        const disabledAttr = state.canEdit ? "" : "disabled";

        return `
          <article class="grading-item" data-criterion-id="${escapeHTML(id)}">
            <div class="grading-head">
              <div>
                <p class="criterion-number">Критерий ${idx + 1}</p>
                <strong>${escapeHTML(criterion.title || "Без названия")}</strong>
              </div>
              <div class="grading-right">
                <span class="criterion-weight">Вес ${escapeHTML(criterion.weight || 1)}</span>
                <div class="grade-switch">
                  <button class="grade-btn yes ${yesActive}" data-grade-value="yes" ${disabledAttr}>Да</button>
                  <button class="grade-btn no ${noActive}" data-grade-value="no" ${disabledAttr}>Нет</button>
                </div>
              </div>
            </div>
            <p class="grading-desc">${escapeHTML(criterion.description || "Описание отсутствует")}</p>
            <label for="comment-${escapeHTML(id)}">Комментарий преподавателя (необязательно)</label>
            <textarea id="comment-${escapeHTML(id)}" class="grade-comment" data-grade-comment="${escapeHTML(id)}" placeholder="Добавьте комментарий при необходимости..." ${disabledAttr}>${escapeHTML(grade.comment || "")}</textarea>
          </article>
        `;
      })
      .join("");

    ui.saveGradingBtn.disabled = !state.canEdit;
    renderSummary();
    if (ui.publishGradingBtn) ui.publishGradingBtn.disabled = !(state.canPublish && state.isComplete);
  }

  function applyGradingPayload(items) {
    state.grading.clear();
    (Array.isArray(items) ? items : []).forEach((item) => {
      const id = String(item.criterion_id || "").trim();
      if (!id) return;
      state.grading.set(id, {
        criterion_id: id,
        is_met: item.is_met === true ? true : item.is_met === false ? false : null,
        comment: String(item.comment || ""),
      });
    });
  }

  async function loadPageData() {
    if (!state.projectID) {
      state.project = null;
      state.criteria = [];
      applyGradingPayload([]);
      renderProjectHeader();
      renderGradingList();
      setStatus("Нет доступных проектов.", true);
      return;
    }

    const [project, criteria, gradingResp] = await Promise.all([
      request("GET", `/v2/projects/${state.projectID}`),
      request("GET", `/v2/projects/${state.projectID}/criteria`),
      request("GET", `/v2/projects/${state.projectID}/grading`),
    ]);

    state.project = project;
    state.criteria = Array.isArray(criteria) ? criteria : [];
    applyGradingPayload(gradingResp && Array.isArray(gradingResp.items) ? gradingResp.items : []);

    renderProjectHeader();
    renderGradingList();

    if (!state.canEdit) {
      const status = String(state.project?.status || "DRAFT").toUpperCase();
      if (status === "ARCHIVE") {
        setStatus("Оценивание завершено. Проект находится в статусе ARCHIVE.", false);
      } else {
        setStatus(`Оценивание недоступно в статусе ${status}. Ожидается статус REVIEW/GRADING.`, true);
      }
    } else if (state.canPublish && !state.isComplete) {
      setStatus("Для завершения отметьте Да/Нет по каждому критерию.", false);
    } else {
      setStatus("Оценка готова к сохранению.", false);
    }
  }

  function setCriterionDecision(criterionID, value) {
    const entry = ensureGradeEntry(criterionID);
    entry.is_met = value;
    state.grading.set(String(criterionID), entry);
  }

  function setCriterionComment(criterionID, comment) {
    const entry = ensureGradeEntry(criterionID);
    entry.comment = String(comment || "").trim();
    state.grading.set(String(criterionID), entry);
  }

  async function saveGrading() {
    if (!state.projectID) {
      setStatus("Проект не выбран.", true);
      return;
    }

    if (!state.canEdit) {
      setStatus("Оценивание сейчас недоступно для этого статуса проекта.", true);
      return;
    }

    const items = state.criteria.map((criterion) => {
      const id = String(criterion.id || "");
      const entry = ensureGradeEntry(id);
      return {
        criterion_id: id,
        is_met: entry.is_met,
        comment: String(entry.comment || ""),
      };
    });

    ui.saveGradingBtn.disabled = true;
    try {
      const resp = await request("PUT", `/v2/projects/${state.projectID}/grading`, { items });
      applyGradingPayload(resp && Array.isArray(resp.items) ? resp.items : []);
      renderGradingList();
      setStatus("Оценка сохранена.", false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      ui.saveGradingBtn.disabled = !state.canEdit;
    }
  }

  async function publishGrading() {
    if (!state.projectID) {
      setStatus("Проект не выбран.", true);
      return;
    }
    if (!state.canPublish) {
      setStatus("Завершение оценивания недоступно для текущего статуса проекта.", true);
      return;
    }
    const confirmed = window.confirm("Завершить оценивание и перевести проект в ARCHIVE?");
    if (!confirmed) return;

    if (ui.publishGradingBtn) ui.publishGradingBtn.disabled = true;
    try {
      await request("POST", `/v2/projects/${state.projectID}/grading/publish`, {});
      await loadPageData();
      setStatus("Оценивание завершено. Проект переведен в ARCHIVE.", false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      if (ui.publishGradingBtn) ui.publishGradingBtn.disabled = !(state.canPublish && state.isComplete);
    }
  }

  function attachEvents() {
    if (ui.logoutBtn) {
      ui.logoutBtn.addEventListener("click", () => {
        auth.logout();
      });
    }

    if (ui.gradingList) {
      ui.gradingList.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-grade-value]");
        if (!btn) return;
        if (!state.canEdit) return;

        const row = btn.closest("[data-criterion-id]");
        if (!row) return;
        const criterionID = String(row.getAttribute("data-criterion-id") || "");
        if (!criterionID) return;

        const value = String(btn.getAttribute("data-grade-value") || "");
        if (value === "yes") {
          setCriterionDecision(criterionID, true);
        } else if (value === "no") {
          setCriterionDecision(criterionID, false);
        }
        renderGradingList();
      });

      ui.gradingList.addEventListener("input", (event) => {
        const input = event.target.closest("textarea[data-grade-comment]");
        if (!input) return;
        const criterionID = String(input.getAttribute("data-grade-comment") || "");
        if (!criterionID) return;
        setCriterionComment(criterionID, input.value);
        renderSummary();
      });
    }

    if (ui.saveGradingBtn) {
      ui.saveGradingBtn.addEventListener("click", () => {
        saveGrading().catch((err) => setStatus(err.message || String(err), true));
      });
    }
    if (ui.publishGradingBtn) {
      ui.publishGradingBtn.addEventListener("click", () => {
        publishGrading().catch((err) => setStatus(err.message || String(err), true));
      });
    }
  }

  async function bootstrap() {
    const claims = await auth.ensureSession("professor");
    if (!claims) return;

    bindProfile();
    attachEvents();

    try {
      setStatus("Загрузка данных...", false);
      await loadProjectList();
      setProjectInURL(state.projectID);
      await loadPageData();
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  bootstrap();
})();
