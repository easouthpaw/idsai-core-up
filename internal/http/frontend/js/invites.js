(() => {
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

  const ui = {
    profileAvatar: document.getElementById("profileAvatar"),
    studentName: document.getElementById("studentName"),
    studentEmail: document.getElementById("studentEmail"),
    logoutBtn: document.getElementById("logoutBtn"),

    tabIncoming: document.getElementById("tabIncoming"),
    tabOutgoing: document.getElementById("tabOutgoing"),
    incomingCount: document.getElementById("incomingCount"),
    outgoingCount: document.getElementById("outgoingCount"),

    incomingPane: document.getElementById("incomingPane"),
    outgoingPane: document.getElementById("outgoingPane"),
    incomingList: document.getElementById("incomingList"),
    outgoingList: document.getElementById("outgoingList"),
    pageStatus: document.getElementById("pageStatus"),
  };

  const state = {
    activeTab: "incoming",
    incoming: [],
    outgoing: [],
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

  function formatDate(raw) {
    if (!raw) return "-";
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) return String(raw);
    return d.toLocaleString();
  }

  function setStatus(message, isError) {
    if (!ui.pageStatus) return;
    ui.pageStatus.textContent = message || "";
    ui.pageStatus.classList.toggle("err", Boolean(isError));
  }

  function decodePayload(token) {
    const parts = token.split(".");
    if (parts.length < 2) throw new Error("invalid JWT");
    let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const mod = payload.length % 4;
    if (mod > 0) payload += "=".repeat(4 - mod);
    return JSON.parse(atob(payload));
  }

  function clearSession() {
    localStorage.removeItem(LS_ACCESS);
    localStorage.removeItem(LS_REFRESH);
    localStorage.removeItem(LS_USER);
    localStorage.removeItem(LS_FACULTY);
    localStorage.removeItem(LS_IS_ADMIN);
    localStorage.removeItem(LS_IS_PROFESSOR);
    localStorage.removeItem(LS_STUDENT_NAME);
    localStorage.removeItem(LS_STUDENT_EMAIL);
  }

  function ensureSession() {
    const access = localStorage.getItem(LS_ACCESS) || "";
    if (!access) {
      window.location.href = "/dev/login";
      return null;
    }

    try {
      const claims = decodePayload(access);
      if (!claims.sub || !claims.faculty_id) throw new Error("broken claims");
      localStorage.setItem(LS_USER, claims.sub);
      localStorage.setItem(LS_FACULTY, claims.faculty_id);
      localStorage.setItem(LS_IS_ADMIN, claims.is_admin ? "1" : "0");
      localStorage.setItem(LS_IS_PROFESSOR, claims.is_professor ? "1" : "0");
      if (claims.is_admin) {
        window.location.href = "/dev/admin";
        return null;
      }
      if (claims.is_professor) {
        window.location.href = "/dev/professor";
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
      headers: authHeaders(body !== undefined),
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    const text = await resp.text();
    let data = text;
    try {
      data = JSON.parse(text);
    } catch (_) {}

    if (!resp.ok) {
      const err = new Error(
        typeof data === "object" && data && data.error
          ? String(data.error)
          : `${resp.status} ${resp.statusText}`
      );
      err.status = resp.status;
      err.data = data;
      throw err;
    }

    return data;
  }

  function bindProfile() {
    const name = localStorage.getItem(LS_STUDENT_NAME) || "Student";
    const email = localStorage.getItem(LS_STUDENT_EMAIL) || "student@university.edu";

    ui.studentName.textContent = name;
    ui.studentEmail.textContent = email;
    ui.profileAvatar.textContent = initials(name, email);
  }

  function badgeClass(status) {
    const code = String(status || "").toLowerCase();
    if (code === "invited") return "invited";
    if (code === "applied") return "applied";
    if (code === "active") return "active";
    if (code === "rejected") return "rejected";
    if (code === "removed") return "removed";
    return "";
  }

  function renderIncoming() {
    ui.incomingList.innerHTML = "";
    if (!state.incoming.length) {
      ui.incomingList.innerHTML = '<article class="empty-card">Входящих приглашений пока нет.</article>';
      return;
    }

    state.incoming.forEach((item) => {
      const card = document.createElement("article");
      card.className = "invite-card";
      card.setAttribute("data-project-id", item.project_id || "");

      const inviter = item.inviter_name || item.inviter_email || "Команда проекта";
      const inviteComment = String(item.invite_comment || "").trim();

      card.innerHTML =
        `<div class="invite-head">` +
          `<div>` +
            `<h3 class="invite-title">${escapeHTML(item.project_title || "Без названия")}</h3>` +
            `<p class="invite-meta">От: ${escapeHTML(inviter)} · ${escapeHTML(formatDate(item.created_at))}</p>` +
          `</div>` +
          `<div class="invite-badges">` +
            `<span class="inv-badge invited">INVITED</span>` +
            `<span class="inv-badge">${escapeHTML(String(item.project_status || "DRAFT").toUpperCase())}</span>` +
          `</div>` +
        `</div>` +
        (inviteComment ? `<div class="invite-comment">${escapeHTML(inviteComment)}</div>` : "") +
        `<div class="invite-actions">` +
          `<button class="invite-btn" data-action="open">Открыть проект</button>` +
          `<button class="invite-btn accept" data-action="accept">Принять</button>` +
          `<button class="invite-btn reject" data-action="reject">Отклонить</button>` +
        `</div>`;

      ui.incomingList.appendChild(card);
    });
  }

  function renderOutgoing() {
    ui.outgoingList.innerHTML = "";
    if (!state.outgoing.length) {
      ui.outgoingList.innerHTML = '<article class="empty-card">Исходящих заявок пока нет.</article>';
      return;
    }

    state.outgoing.forEach((item) => {
      const status = String(item.status || "APPLIED").toUpperCase();
      const card = document.createElement("article");
      card.className = "invite-card";
      card.innerHTML =
        `<div class="invite-head">` +
          `<div>` +
            `<h3 class="invite-title">${escapeHTML(item.project_title || "Без названия")}</h3>` +
            `<p class="invite-meta">Создано: ${escapeHTML(formatDate(item.created_at))}` +
              (item.responded_at ? ` · Ответ: ${escapeHTML(formatDate(item.responded_at))}` : "") +
            `</p>` +
          `</div>` +
          `<div class="invite-badges">` +
            `<span class="inv-badge ${escapeHTML(badgeClass(status))}">${escapeHTML(status)}</span>` +
            `<span class="inv-badge">${escapeHTML(String(item.project_status || "DRAFT").toUpperCase())}</span>` +
          `</div>` +
        `</div>` +
        `<div class="invite-actions">` +
          `<a class="invite-btn" href="/dev/projects/${encodeURIComponent(item.project_id || "")}">Открыть проект</a>` +
        `</div>`;
      ui.outgoingList.appendChild(card);
    });
  }

  function renderCounts() {
    ui.incomingCount.textContent = String(state.incoming.length);
    ui.outgoingCount.textContent = String(state.outgoing.length);
  }

  function setTab(tab) {
    state.activeTab = tab === "outgoing" ? "outgoing" : "incoming";
    const incomingActive = state.activeTab === "incoming";

    ui.tabIncoming.classList.toggle("active", incomingActive);
    ui.tabOutgoing.classList.toggle("active", !incomingActive);
    ui.tabIncoming.setAttribute("aria-selected", incomingActive ? "true" : "false");
    ui.tabOutgoing.setAttribute("aria-selected", incomingActive ? "false" : "true");

    ui.incomingPane.classList.toggle("active", incomingActive);
    ui.outgoingPane.classList.toggle("active", !incomingActive);
  }

  async function loadInvites() {
    const [incomingResp, outgoingResp] = await Promise.all([
      request("GET", "/v2/invites/incoming?limit=100"),
      request("GET", "/v2/invites/outgoing?limit=100"),
    ]);

    state.incoming = Array.isArray(incomingResp && incomingResp.items) ? incomingResp.items : [];
    state.outgoing = Array.isArray(outgoingResp && outgoingResp.items) ? outgoingResp.items : [];

    renderCounts();
    renderIncoming();
    renderOutgoing();
  }

  async function onIncomingAction(button) {
    const card = button.closest("[data-project-id]");
    if (!card) return;

    const projectID = String(card.getAttribute("data-project-id") || "").trim();
    if (!projectID) return;

    const action = button.getAttribute("data-action");

    if (action === "open") {
      window.location.href = `/dev/projects/${projectID}`;
      return;
    }

    const accept = action === "accept";
    const actionLabel = accept ? "принять" : "отклонить";
    if (!window.confirm(`Подтвердите действие: ${actionLabel} приглашение?`)) {
      return;
    }

    const allButtons = Array.from(card.querySelectorAll("button[data-action]"));
    allButtons.forEach((btn) => {
      btn.disabled = true;
    });

    try {
      await request("POST", `/v2/projects/${projectID}/members/respond`, { accept });
      setStatus(accept ? "Приглашение принято." : "Приглашение отклонено.", false);
      await loadInvites();
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      allButtons.forEach((btn) => {
        btn.disabled = false;
      });
    }
  }

  function wireEvents() {
    ui.logoutBtn.addEventListener("click", () => {
      clearSession();
      window.location.href = "/dev/login";
    });

    ui.tabIncoming.addEventListener("click", () => setTab("incoming"));
    ui.tabOutgoing.addEventListener("click", () => setTab("outgoing"));

    ui.incomingList.addEventListener("click", async (event) => {
      const button = event.target.closest("button[data-action]");
      if (!button) return;
      await onIncomingAction(button);
    });
  }

  async function bootstrap() {
    const claims = ensureSession();
    if (!claims) return;

    bindProfile();
    wireEvents();

    try {
      setStatus("Загрузка заявок...", false);
      await loadInvites();
      setTab("incoming");
      setStatus("Заявки загружены.", false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  bootstrap();
})();
