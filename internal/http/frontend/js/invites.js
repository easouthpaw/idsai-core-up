(() => {
  const auth = window.IDSAIAuth;
  const i18n = window.IDSAI18n;
  const roleSidebar = window.IDSAIRoleSidebar;
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";
  const LS_STUDENT_SECTION = "idsai_student_section";
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

  function projectStatusLabel(status) {
    const code = String(status || "").toUpperCase();
    if (code === "DRAFT" || code === "REVIEW") return "ПОДГОТОВКА";
    if (code === "RECRUITMENT") return "НАБОР";
    if (code === "ACTIVE") return "В РАБОТЕ";
    if (code === "GRADING") return "ОЦЕНИВАНИЕ";
    if (code === "COMPLETED" || code === "ARCHIVE") return "ЗАВЕРШЕН";
    return code || "СТАТУС";
  }

  function formatDate(raw) {
    if (!raw) return "-";
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) return String(raw);
    return i18n ? i18n.formatDateTime(d) : d.toLocaleString();
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
    if (claims.is_professor) {
      window.location.href = "/dev/professor";
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
      headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

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

  function confirmAction(options) {
    if (auth && typeof auth.showConfirmDialog === "function") {
      return auth.showConfirmDialog(options);
    }
    return Promise.resolve(window.confirm(String((options && options.message) || "")));
  }

  function syncSidebar(profile) {
    const host = document.querySelector("[data-role-sidebar]");
    if (!host || !roleSidebar || typeof roleSidebar.renderSidebar !== "function") {
      return;
    }

    host.dataset.sidebarActive = "invites";
    roleSidebar.renderSidebar(host, {
      role: "student",
      active: "invites",
      profile,
      scope: typeof auth.getDefaultScope === "function" ? auth.getDefaultScope() : null,
    });

    const logoutBtn = document.getElementById("logoutBtn");
    if (logoutBtn) {
      logoutBtn.onclick = () => {
        auth.logout();
      };
    }
  }

  function bindProfile(profile) {
    syncSidebar(profile || auth.getCachedProfile());
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

  function incomingStatusLabel(status) {
    const code = String(status || "").toUpperCase();
    if (code === "APPLIED") return "ЗАЯВКА";
    if (code === "INVITED") return "ПРИГЛАШЕНИЕ";
    return code || "СТАТУС";
  }

  function outgoingStatusLabel(status) {
    const code = String(status || "").toUpperCase();
    if (code === "APPLIED") return "НА РАССМОТРЕНИИ";
    if (code === "ACTIVE") return "ПРИНЯТО";
    if (code === "REJECTED") return "ОТКЛОНЕНО";
    if (code === "REMOVED") return "ОТОЗВАНО";
    if (code === "INVITED") return "ПРИГЛАШЕНИЕ";
    return code || "СТАТУС";
  }

  function outgoingDecisionText(status) {
    const code = String(status || "").toUpperCase();
    if (code === "ACTIVE") return "Решение: заявка принята.";
    if (code === "REJECTED") return "Решение: заявка отклонена.";
    if (code === "APPLIED") return "Решение: ожидает ответа тимлида.";
    if (code === "REMOVED") return "Решение: заявка отозвана/снята.";
    return "";
  }

  function outgoingDecisionClass(status) {
    const code = String(status || "").toUpperCase();
    if (code === "ACTIVE") return "accepted";
    if (code === "REJECTED") return "rejected";
    return "pending";
  }

  function renderIncoming() {
    ui.incomingList.innerHTML = "";
    if (!state.incoming.length) {
      ui.incomingList.innerHTML = '<article class="empty-card">Входящих заявок и приглашений пока нет.</article>';
      return;
    }

    state.incoming.forEach((item) => {
      const card = document.createElement("article");
      card.className = "invite-card";
      card.setAttribute("data-project-id", item.project_id || "");
      card.setAttribute("data-member-user-id", item.user_id || "");
      card.setAttribute("data-membership-status", String(item.status || "INVITED").toUpperCase());

      const inviter = item.inviter_name || item.inviter_email || "Команда проекта";
      const inviteComment = String(item.invite_comment || "").trim();
      const memberStatus = String(item.status || "INVITED").toUpperCase();
      const memberStatusLabel = incomingStatusLabel(memberStatus);
      const isDirectInvite = memberStatus === "INVITED";
      const actionHTML = isDirectInvite
        ? `<button class="invite-btn" data-action="open">Открыть проект</button>` +
          `<button class="invite-btn accept" data-action="accept">Принять</button>` +
          `<button class="invite-btn reject" data-action="reject">Отклонить</button>`
        : `<a class="invite-btn" href="/dev/projects/${encodeURIComponent(item.project_id || "")}?nav=invites">Открыть проект</a>` +
          `<button class="invite-btn accept" data-action="accept">Принять</button>` +
          `<button class="invite-btn reject" data-action="reject">Отклонить</button>`;

      card.innerHTML =
        `<div class="invite-head">` +
          `<div>` +
            `<h3 class="invite-title">${escapeHTML(item.project_title || "Без названия")}</h3>` +
            `<p class="invite-meta">От: ${escapeHTML(inviter)} · ${escapeHTML(formatDate(item.created_at))}</p>` +
          `</div>` +
          `<div class="invite-badges">` +
            `<span class="inv-badge ${escapeHTML(badgeClass(memberStatus))}">${escapeHTML(memberStatusLabel)}</span>` +
            `<span class="inv-badge">${escapeHTML(projectStatusLabel(item.project_status || "DRAFT"))}</span>` +
          `</div>` +
        `</div>` +
        (inviteComment ? `<div class="invite-comment">${escapeHTML(inviteComment)}</div>` : "") +
        `<div class="invite-actions">` +
          actionHTML +
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
      const statusLabel = outgoingStatusLabel(status);
      const decisionText = outgoingDecisionText(status);
      const decisionClass = outgoingDecisionClass(status);
      const card = document.createElement("article");
      card.className = `invite-card ${decisionClass === "accepted" ? "decision-accepted" : ""} ${decisionClass === "rejected" ? "decision-rejected" : ""}`.trim();
      card.innerHTML =
        `<div class="invite-head">` +
          `<div>` +
            `<h3 class="invite-title">${escapeHTML(item.project_title || "Без названия")}</h3>` +
            `<p class="invite-meta">Создано: ${escapeHTML(formatDate(item.created_at))}` +
              (item.responded_at ? ` · Ответ: ${escapeHTML(formatDate(item.responded_at))}` : "") +
            `</p>` +
            (decisionText ? `<p class="invite-meta invite-decision ${escapeHTML(decisionClass)}">${escapeHTML(decisionText)}</p>` : "") +
          `</div>` +
          `<div class="invite-badges">` +
            `<span class="inv-badge ${escapeHTML(badgeClass(status))}">${escapeHTML(statusLabel)}</span>` +
            `<span class="inv-badge">${escapeHTML(projectStatusLabel(item.project_status || "DRAFT"))}</span>` +
          `</div>` +
        `</div>` +
        `<div class="invite-actions">` +
          `<a class="invite-btn" href="/dev/projects/${encodeURIComponent(item.project_id || "")}?nav=invites">Открыть проект</a>` +
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
    const memberUserID = String(card.getAttribute("data-member-user-id") || "").trim();
    const memberStatus = String(card.getAttribute("data-membership-status") || "INVITED").toUpperCase();

    const action = button.getAttribute("data-action");

    if (action === "open") {
      localStorage.setItem(LS_STUDENT_SECTION, "invites");
      window.location.href = `/dev/projects/${projectID}?nav=invites`;
      return;
    }
    if ((action === "accept" || action === "reject") && memberStatus === "APPLIED") {
      if (!memberUserID) {
        setStatus("Не удалось определить пользователя заявки.", true);
        return;
      }
      const accept = action === "accept";
      const actionLabel = accept ? "принять" : "отклонить";
      if (!await confirmAction({
        title: accept ? "Принять заявку" : "Отклонить заявку",
        message: `Подтвердите действие: ${actionLabel} заявку участника в проект.`,
        confirmText: accept ? "Принять заявку" : "Отклонить заявку",
        danger: !accept,
      })) {
        return;
      }
      const allButtons = Array.from(card.querySelectorAll("button[data-action]"));
      allButtons.forEach((btn) => {
        btn.disabled = true;
      });
      try {
        if (accept) {
          await request("POST", `/v2/projects/${projectID}/members/${memberUserID}/approve`, {});
        } else {
          await request("POST", `/v2/projects/${projectID}/members/${memberUserID}/reject`, {});
        }
        setStatus(accept ? "Заявка принята. Участник добавлен в команду." : "Заявка отклонена.", false);
        await loadInvites();
      } catch (err) {
        setStatus(err.message || String(err), true);
      } finally {
        allButtons.forEach((btn) => {
          btn.disabled = false;
        });
      }
      return;
    }
    if (memberStatus !== "INVITED") {
      return;
    }

    const accept = action === "accept";
    const actionLabel = accept ? "принять" : "отклонить";
    if (!await confirmAction({
      title: accept ? "Принять приглашение" : "Отклонить приглашение",
      message: `Подтвердите действие: ${actionLabel} приглашение в проект.`,
      confirmText: accept ? "Принять приглашение" : "Отклонить приглашение",
      danger: !accept,
    })) {
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
    ui.tabIncoming.addEventListener("click", () => setTab("incoming"));
    ui.tabOutgoing.addEventListener("click", () => setTab("outgoing"));

    ui.incomingList.addEventListener("click", async (event) => {
      const button = event.target.closest("button[data-action]");
      if (!button) return;
      await onIncomingAction(button);
    });
  }

  async function bootstrap() {
    const claims = await auth.ensureSession("student");
    if (!claims) return;

    localStorage.setItem(LS_STUDENT_SECTION, "invites");
    bindProfile(claims);
    wireEvents();

    try {
      setStatus("Загрузка заявок...", false);
      await loadInvites();
      setTab("incoming");
      setStatus("Заявки загружены.", false);
      auth.setPageLoading(false);
    } catch (err) {
      setStatus(err.message || String(err), true);
      auth.setPageLoading(false);
    }
  }

  bootstrap().catch((err) => {
    auth.setPageLoading(false);
    setStatus(err.message || String(err), true);
  });
})();
