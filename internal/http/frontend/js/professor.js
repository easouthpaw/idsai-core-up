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

  const projectsBody = document.getElementById("projectsBody");
  const reviewsBody = document.getElementById("reviewsBody");
  const refreshBtn = document.getElementById("refreshBtn");
  const statusEl = document.getElementById("pageStatus");
  const logoutBtn = document.getElementById("logoutBtn");

  const statTotalEl = document.getElementById("statTotal");
  const statReviewEl = document.getElementById("statReview");
  const statActiveEl = document.getElementById("statActive");
  const statRecruitmentEl = document.getElementById("statRecruitment");

  let claims = null;
  let projects = [];
  let reviewInvites = [];

  function decodePayload(token) {
    const parts = String(token || "").split(".");
    if (parts.length < 2) throw new Error("invalid JWT");
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

  function setStatus(msg, isError) {
    if (!statusEl) return;
    statusEl.textContent = msg || "";
    statusEl.classList.toggle("err", Boolean(isError));
  }

  async function request(method, url, body) {
    const { resp, data } = await auth.requestJSON(url, {
      method,
      headers: body ? { "Content-Type": "application/json" } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!resp.ok) {
      const errMsg = data && data.error ? data.error : `${method} ${url} failed (${resp.status})`;
      throw new Error(errMsg);
    }
    return data;
  }

  async function loadProjects() {
    const [mine, pub] = await Promise.all([
      request("GET", "/v2/projects/my"),
      request("GET", "/v2/projects/public"),
    ]);
    const map = new Map();
    (Array.isArray(mine) ? mine : []).forEach((item) => map.set(item.id, item));
    (Array.isArray(pub) ? pub : []).forEach((item) => map.set(item.id, item));
    projects = Array.from(map.values()).sort((a, b) => {
      const ad = new Date(a.updated_at || a.created_at || 0).getTime();
      const bd = new Date(b.updated_at || b.created_at || 0).getTime();
      return bd - ad;
    });
    return projects;
  }

  async function loadReviewInvites() {
    const items = await request("GET", "/v2/professor/review-invites?limit=100");
    reviewInvites = Array.isArray(items) ? items : [];
    return reviewInvites;
  }

  function statusClass(status) {
    const s = String(status || "").toUpperCase();
    if (s === "REVIEW") return "review";
    if (s === "ACTIVE") return "active";
    if (s === "RECRUITMENT") return "recruitment";
    return "default";
  }

  function formatDate(v) {
    if (!v) return "—";
    try {
      return new Date(v).toLocaleString();
    } catch (_) {
      return String(v);
    }
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

  function bindProfile() {
    const name = localStorage.getItem(LS_STUDENT_NAME) || "Преподаватель";
    const email = localStorage.getItem(LS_STUDENT_EMAIL) || "professor@idsai.dev";
    const avatarURL = localStorage.getItem(LS_AVATAR_URL) || "";
    const avatar = document.getElementById("profAvatar");
    const nameEl = document.getElementById("profName");
    const emailEl = document.getElementById("profEmail");
    renderAvatar(avatar, initials(name, email), avatarURL);
    if (nameEl) nameEl.textContent = name;
    if (emailEl) emailEl.textContent = email;
  }

  function updateStats(items) {
    if (!statTotalEl) return;
    const total = items.length;
    const review = items.filter((x) => String(x.status || "").toUpperCase() === "REVIEW").length;
    const active = items.filter((x) => String(x.status || "").toUpperCase() === "ACTIVE").length;
    const recruitment = items.filter((x) => String(x.status || "").toUpperCase() === "RECRUITMENT").length;
    statTotalEl.textContent = String(total);
    statReviewEl.textContent = String(review);
    statActiveEl.textContent = String(active);
    statRecruitmentEl.textContent = String(recruitment);
  }

  function renderDashboard(items) {
    if (!projectsBody) return;
    if (!items.length) {
      projectsBody.innerHTML = '<tr><td colspan="4">Пока нет доступных проектов.</td></tr>';
      return;
    }

    projectsBody.innerHTML = items.map((p) => {
      const s = String(p.status || "").toUpperCase();
      const hasProfessor = Boolean(p.professor_id);
      return `
        <tr>
          <td>
            <strong>${escapeHTML(p.title || "Без названия")}</strong>
            <div class="muted">${escapeHTML(p.description || "Описание не заполнено")}</div>
          </td>
          <td><span class="status-pill ${statusClass(s)}">${escapeHTML(s || "DRAFT")}</span></td>
          <td>${escapeHTML(formatDate(p.updated_at || p.created_at))}</td>
          <td>
            <div class="actions">
              <button class="action-btn" data-act="open" data-id="${escapeHTML(p.id)}">Открыть</button>
              <button class="action-btn" data-act="recruitment" data-id="${escapeHTML(p.id)}">Набор</button>
              <button class="action-btn" data-act="attach" data-id="${escapeHTML(p.id)}" ${hasProfessor ? "disabled" : ""}>Я преподаватель</button>
              <button class="action-btn" data-act="criteria" data-id="${escapeHTML(p.id)}">Критерии</button>
              <button class="action-btn" data-act="grade" data-id="${escapeHTML(p.id)}">Оценивание</button>
              <button class="action-btn primary" data-act="start" data-id="${escapeHTML(p.id)}" ${s === "ACTIVE" ? "disabled" : ""}>Старт</button>
            </div>
          </td>
        </tr>
      `;
    }).join("");
  }

  function renderReviews(items) {
    if (!reviewsBody) return;
    const queue = Array.isArray(items) ? items : [];

    if (!queue.length) {
      reviewsBody.innerHTML = '<tr><td colspan="4">Нет заявок на ревью.</td></tr>';
      return;
    }

    reviewsBody.innerHTML = queue.map((p) => {
      const s = String(p.status || "").toUpperCase();
      return `
        <tr>
          <td>
            <strong>${escapeHTML(p.title || "Без названия")}</strong>
            <div class="muted">${escapeHTML(p.description || "Описание не заполнено")}</div>
          </td>
          <td><span class="muted">${escapeHTML(p.group_id || "Общая группа")}</span></td>
          <td><span class="status-pill ${statusClass(s)}">${escapeHTML(s)}</span></td>
          <td>
            <div class="actions">
              <button class="action-btn primary" data-act="accept" data-id="${escapeHTML(p.id)}">Принять</button>
              <button class="action-btn" data-act="reject" data-id="${escapeHTML(p.id)}">Отклонить</button>
              <button class="action-btn" data-act="open" data-id="${escapeHTML(p.id)}">Подробнее</button>
            </div>
          </td>
        </tr>
      `;
    }).join("");
  }

  async function actionRecruitment(projectID) {
    await request("POST", `/v2/projects/${projectID}/recruitment/open`, {});
    setStatus("Набор команды открыт.", false);
  }

  async function actionAttachProfessor(projectID) {
    await request("POST", `/v2/projects/${projectID}/professor`, { professor_id: claims.sub });
    setStatus("Вы прикреплены к проекту как преподаватель.", false);
  }

  function actionOpenCriteria(projectID) {
    window.location.href = `/dev/professor/criteria?project_id=${encodeURIComponent(projectID)}`;
  }

  function actionOpenGrading(projectID) {
    window.location.href = `/dev/professor/grading?project_id=${encodeURIComponent(projectID)}`;
  }

  async function actionStartProject(projectID) {
    await request("POST", `/v2/projects/${projectID}/approve`, {});
    setStatus("Проект запущен (ACTIVE).", false);
  }

  async function actionRespondProfessorInvite(projectID, accept) {
    await request("POST", `/v2/projects/${projectID}/professor/respond`, { accept: Boolean(accept) });
    setStatus(accept ? "Приглашение на ревью принято." : "Приглашение отклонено.", false);
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  async function refreshPage() {
    try {
      setStatus("Обновление данных...", false);
      let items = [];
      if (projectsBody || statTotalEl) {
        items = await loadProjects();
        updateStats(items);
        renderDashboard(items);
      }
      if (reviewsBody) {
        const queue = await loadReviewInvites();
        renderReviews(queue);
        setStatus(`Загружено приглашений: ${queue.length}.`, false);
        return;
      }
      setStatus(`Загружено проектов: ${items.length}.`, false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  async function handleAction(act, projectID) {
    if (!act || !projectID) return;
    try {
      if (act === "open") {
        window.location.href = `/dev/projects/${projectID}`;
        return;
      }
      if (act === "accept") {
        await actionRespondProfessorInvite(projectID, true);
      } else if (act === "reject") {
        await actionRespondProfessorInvite(projectID, false);
      } else if (act === "recruitment") {
        await actionRecruitment(projectID);
      } else if (act === "attach") {
        await actionAttachProfessor(projectID);
      } else if (act === "criteria") {
        actionOpenCriteria(projectID);
        return;
      } else if (act === "grade") {
        actionOpenGrading(projectID);
        return;
      } else if (act === "start") {
        await actionStartProject(projectID);
      }
      await refreshPage();
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  function attachTableListeners() {
    if (projectsBody) {
      projectsBody.addEventListener("click", (e) => {
        const btn = e.target.closest("button[data-act][data-id]");
        if (!btn) return;
        handleAction(btn.dataset.act, btn.dataset.id);
      });
    }
    if (reviewsBody) {
      reviewsBody.addEventListener("click", (e) => {
        const btn = e.target.closest("button[data-act][data-id]");
        if (!btn) return;
        handleAction(btn.dataset.act, btn.dataset.id);
      });
    }
  }

  function attachCommonActions() {
    if (refreshBtn) {
      refreshBtn.addEventListener("click", () => {
        refreshPage();
      });
    }
    if (logoutBtn) {
      logoutBtn.addEventListener("click", () => {
        auth.logout();
      });
    }
  }

  void (async () => {
    claims = await auth.ensureSession("professor");
    if (!claims) return;

    bindProfile();
    attachTableListeners();
    attachCommonActions();
    refreshPage();
  })();
})();
