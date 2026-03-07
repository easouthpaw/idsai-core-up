(() => {
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

  const ui = {
    profAvatar: document.getElementById("profAvatar"),
    profName: document.getElementById("profName"),
    profEmail: document.getElementById("profEmail"),
    logoutBtn: document.getElementById("logoutBtn"),
    openProjectBtn: document.getElementById("openProjectBtn"),
    openGradingBtn: document.getElementById("openGradingBtn"),
    projectSelect: document.getElementById("projectSelect"),
    criteriaList: document.getElementById("criteriaList"),
    criterionForm: document.getElementById("criterionForm"),
    criterionTitleInput: document.getElementById("criterionTitleInput"),
    criterionWeightInput: document.getElementById("criterionWeightInput"),
    criterionDescInput: document.getElementById("criterionDescInput"),
    criterionSubmitBtn: document.getElementById("criterionSubmitBtn"),
    pageStatus: document.getElementById("pageStatus"),
  };

  const state = {
    projectID: "",
    projects: [],
    criteria: [],
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
    localStorage.removeItem(LS_ACCESS);
    localStorage.removeItem(LS_REFRESH);
    localStorage.removeItem(LS_USER);
    localStorage.removeItem(LS_IS_ADMIN);
    localStorage.removeItem(LS_IS_PROFESSOR);
  }

  function ensureSession() {
    const access = localStorage.getItem(LS_ACCESS) || "";
    if (!access) {
      window.location.href = "/dev/login";
      return null;
    }

    try {
      const claims = decodePayload(access);
      if (!claims.sub) throw new Error("missing sub");
      localStorage.setItem(LS_USER, claims.sub);
      localStorage.setItem(LS_IS_ADMIN, claims.is_admin ? "1" : "0");
      localStorage.setItem(LS_IS_PROFESSOR, claims.is_professor ? "1" : "0");
      if (claims.is_admin) {
        window.location.href = "/dev/admin";
        return null;
      }
      if (!claims.is_professor) {
        window.location.href = "/dev/projects";
        return null;
      }
      return claims;
    } catch (_) {
      clearSession();
      window.location.href = "/dev/login";
      return null;
    }
  }

  function authHeaders(withJSON) {
    const headers = {};
    if (withJSON) headers["Content-Type"] = "application/json";
    const access = localStorage.getItem(LS_ACCESS) || "";
    if (access) headers.Authorization = `Bearer ${access}`;
    return headers;
  }

  async function request(method, url, body) {
    const resp = await fetch(url, {
      method,
      headers: authHeaders(Boolean(body)),
      body: body ? JSON.stringify(body) : undefined,
    });

    if (resp.status === 401) {
      clearSession();
      window.location.href = "/dev/login";
      return null;
    }

    const text = await resp.text();
    let data = {};
    try {
      data = text ? JSON.parse(text) : {};
    } catch (_) {
      data = text;
    }

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
    if (ui.profName) ui.profName.textContent = name;
    if (ui.profEmail) ui.profEmail.textContent = email;
    if (ui.profAvatar) ui.profAvatar.textContent = initials(name, email);
  }

  function normalizeProjects(items) {
    const list = Array.isArray(items) ? items : [];
    return list.sort((a, b) => {
      const ad = new Date(a.updated_at || a.created_at || 0).getTime();
      const bd = new Date(b.updated_at || b.created_at || 0).getTime();
      return bd - ad;
    });
  }

  async function loadProjects() {
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

    renderProjectSelect();
    updateActionLinks();
  }

  function renderProjectSelect() {
    if (!ui.projectSelect) return;

    if (!state.projects.length) {
      ui.projectSelect.innerHTML = "<option value=''>Нет доступных проектов</option>";
      ui.projectSelect.disabled = true;
      return;
    }

    ui.projectSelect.disabled = false;
    ui.projectSelect.innerHTML = state.projects
      .map((p) => {
        const selected = String(p.id) === String(state.projectID) ? "selected" : "";
        const status = String(p.status || "DRAFT").toUpperCase();
        return `<option value="${escapeHTML(p.id)}" ${selected}>${escapeHTML(p.title || "Без названия")} · ${escapeHTML(status)}</option>`;
      })
      .join("");
  }

  function updateActionLinks() {
    const projectID = String(state.projectID || "");

    if (ui.openProjectBtn) {
      ui.openProjectBtn.href = projectID ? `/dev/projects/${projectID}` : "/dev/projects";
    }

    if (ui.openGradingBtn) {
      ui.openGradingBtn.href = projectID ? `/dev/professor/grading?project_id=${encodeURIComponent(projectID)}` : "#";
      ui.openGradingBtn.classList.toggle("disabled-link", !projectID);
      if (!projectID) {
        ui.openGradingBtn.setAttribute("aria-disabled", "true");
      } else {
        ui.openGradingBtn.removeAttribute("aria-disabled");
      }
    }
  }

  async function loadCriteria() {
    if (!state.projectID) {
      state.criteria = [];
      renderCriteria();
      return;
    }

    const items = await request("GET", `/v2/projects/${state.projectID}/criteria`);
    state.criteria = Array.isArray(items) ? items : [];
    renderCriteria();
  }

  function renderCriteria() {
    if (!ui.criteriaList) return;

    if (!state.criteria.length) {
      ui.criteriaList.innerHTML = '<div class="empty-state">Критерии еще не добавлены.</div>';
      return;
    }

    ui.criteriaList.innerHTML = state.criteria
      .map((item, idx) => {
        const weight = Number(item.weight || 0);
        return `
          <article class="criterion-row">
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
  }

  async function createCriterion(event) {
    event.preventDefault();

    if (!state.projectID) {
      setStatus("Сначала выберите проект.", true);
      return;
    }

    const title = String(ui.criterionTitleInput.value || "").trim();
    const description = String(ui.criterionDescInput.value || "").trim();
    const weight = Number.parseInt(String(ui.criterionWeightInput.value || "10"), 10);

    if (!title) {
      setStatus("Введите название критерия.", true);
      return;
    }

    if (!Number.isFinite(weight) || weight < 1 || weight > 100) {
      setStatus("Вес должен быть числом от 1 до 100.", true);
      return;
    }

    ui.criterionSubmitBtn.disabled = true;
    try {
      await request("POST", `/v2/projects/${state.projectID}/criteria`, {
        title,
        description,
        weight,
      });
      ui.criterionTitleInput.value = "";
      ui.criterionDescInput.value = "";
      ui.criterionWeightInput.value = "10";
      setStatus("Критерий добавлен.", false);
      await loadCriteria();
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      ui.criterionSubmitBtn.disabled = false;
    }
  }

  async function onProjectChange() {
    state.projectID = String(ui.projectSelect.value || "").trim();
    setProjectInURL(state.projectID);
    updateActionLinks();
    await loadCriteria();
  }

  function attachEvents() {
    if (ui.logoutBtn) {
      ui.logoutBtn.addEventListener("click", () => {
        clearSession();
        window.location.href = "/dev/login";
      });
    }

    if (ui.projectSelect) {
      ui.projectSelect.addEventListener("change", () => {
        onProjectChange().catch((err) => setStatus(err.message || String(err), true));
      });
    }

    if (ui.criterionForm) {
      ui.criterionForm.addEventListener("submit", createCriterion);
    }
  }

  async function bootstrap() {
    const claims = ensureSession();
    if (!claims) return;

    bindProfile();
    attachEvents();

    try {
      setStatus("Загрузка данных...", false);
      await loadProjects();
      setProjectInURL(state.projectID);
      await loadCriteria();
      setStatus("Готово к редактированию критериев.", false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  bootstrap();
})();
