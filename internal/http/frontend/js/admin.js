(() => {
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_IS_ADMIN = "idsai_is_admin";

  const viewButtons = Array.from(document.querySelectorAll(".side-link[data-view]"));
  const viewEls = Array.from(document.querySelectorAll(".view"));
  const crumbCurrentEl = document.getElementById("crumbCurrent");
  const quickSearchInputEl = document.getElementById("quickSearchInput");

  const openUsersViewBtnEl = document.getElementById("openUsersViewBtn");
  const openProjectsViewBtnEl = document.getElementById("openProjectsViewBtn");

  const usersBodyEl = document.getElementById("usersBody");
  const usersMetaEl = document.getElementById("usersMeta");
  const searchInputEl = document.getElementById("searchInput");
  const usersRefreshBtnEl = document.getElementById("usersRefreshBtn");
  const userRoleTabs = Array.from(document.querySelectorAll("#view-users .tab[data-role]"));

  const projectsBodyEl = document.getElementById("projectsBody");
  const projectsMetaEl = document.getElementById("projectsMeta");
  const projectSearchInputEl = document.getElementById("projectSearchInput");
  const projectsRefreshBtnEl = document.getElementById("projectsRefreshBtn");
  const projectStatusTabs = Array.from(document.querySelectorAll("#view-projects .tab[data-status]"));

  const quickAddStudentBtnEl = document.getElementById("quickAddStudentBtn");
  const quickAddProfessorBtnEl = document.getElementById("quickAddProfessorBtn");
  const addStudentBtnEl = document.getElementById("addStudentBtn");
  const addProfessorBtnEl = document.getElementById("addProfessorBtn");
  const logoutBtnEl = document.getElementById("logoutBtn");

  const statTotalUsersEl = document.getElementById("statTotalUsers");
  const statStudentsEl = document.getElementById("statStudents");
  const statProfessorsEl = document.getElementById("statProfessors");
  const statDisabledEl = document.getElementById("statDisabled");

  const statTotalProjectsEl = document.getElementById("statTotalProjects");
  const statReviewProjectsEl = document.getElementById("statReviewProjects");
  const statApprovedProjectsEl = document.getElementById("statApprovedProjects");
  const statRejectedProjectsEl = document.getElementById("statRejectedProjects");

  const projTotalEl = document.getElementById("projTotal");
  const projReviewEl = document.getElementById("projReview");
  const projApprovedEl = document.getElementById("projApproved");
  const projRejectedEl = document.getElementById("projRejected");

  const activityListEl = document.getElementById("activityList");

  const modalEl = document.getElementById("createModal");
  const closeModalBtnEl = document.getElementById("closeModalBtn");
  const createTitleEl = document.getElementById("createTitle");
  const createSubtitleEl = document.getElementById("createSubtitle");
  const modalStatusEl = document.getElementById("modalStatus");
  const fullNameInputEl = document.getElementById("fullNameInput");
  const emailInputEl = document.getElementById("emailInput");
  const passwordInputEl = document.getElementById("passwordInput");
  const departmentInputEl = document.getElementById("departmentInput");
  const submitCreateBtnEl = document.getElementById("submitCreateBtn");

  const projectStatusModalEl = document.getElementById("projectStatusModal");
  const projectModalCloseBtnEl = document.getElementById("projectModalCloseBtn");
  const projectModalCancelBtnEl = document.getElementById("projectModalCancelBtn");
  const projectModalSubmitBtnEl = document.getElementById("projectModalSubmitBtn");
  const projectModalTitleEl = document.getElementById("projectModalTitle");
  const projectModalSubtitleEl = document.getElementById("projectModalSubtitle");
  const projectModalExpectedEl = document.getElementById("projectModalExpected");
  const projectModalStatusEl = document.getElementById("projectModalStatus");
  const projectConfirmInputEl = document.getElementById("projectConfirmInput");

  const adminNameEl = document.getElementById("adminName");
  const adminEmailEl = document.getElementById("adminEmail");
  const adminAvatarEl = document.getElementById("adminAvatar");
  const sidebarNameEl = document.getElementById("sidebarName");
  const sidebarInitialsEl = document.getElementById("sidebarInitials");

  const state = {
    activeView: "dashboard",
    usersRole: "",
    usersSearch: "",
    projectsStatus: "",
    projectsSearch: "",
    dashboardSearch: "",
    createRole: "STUDENT",
    users: [],
    projects: [],
    dashboardUsers: [],
    dashboardProjects: [],
    pendingProjectAction: null,
    quickSearchTimer: null,
    usersSearchTimer: null,
    projectsSearchTimer: null,
  };

  function decodePayload(token) {
    const parts = String(token || "").split(".");
    if (parts.length < 2) throw new Error("invalid JWT");
    let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const mod = payload.length % 4;
    if (mod > 0) payload += "=".repeat(4 - mod);
    return JSON.parse(atob(payload));
  }

  function initials(v) {
    const text = String(v || "").trim();
    if (!text) return "AD";
    const parts = text.split(/\s+/).filter(Boolean);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return text.slice(0, 2).toUpperCase();
  }

  function clearSession() {
    localStorage.removeItem(LS_ACCESS);
    localStorage.removeItem(LS_REFRESH);
    localStorage.removeItem(LS_USER);
    localStorage.removeItem(LS_FACULTY);
    localStorage.removeItem(LS_STUDENT_NAME);
    localStorage.removeItem(LS_STUDENT_EMAIL);
    localStorage.removeItem(LS_IS_ADMIN);
  }

  function ensureAdminSession() {
    const access = localStorage.getItem(LS_ACCESS) || "";
    if (!access) {
      window.location.href = "/dev/login";
      return null;
    }

    try {
      const claims = decodePayload(access);
      if (!claims.sub || !claims.faculty_id) throw new Error("broken token");
      localStorage.setItem(LS_USER, claims.sub);
      localStorage.setItem(LS_FACULTY, claims.faculty_id);
      localStorage.setItem(LS_IS_ADMIN, claims.is_admin ? "1" : "0");
      if (!claims.is_admin) {
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

  function hydrateProfile() {
    const email = localStorage.getItem(LS_STUDENT_EMAIL) || "admin@idsai.local";
    const name = localStorage.getItem(LS_STUDENT_NAME) || "Главный администратор";
    const iv = initials(name || email);

    adminNameEl.textContent = name;
    sidebarNameEl.textContent = name;
    adminEmailEl.textContent = email;
    adminAvatarEl.textContent = iv;
    sidebarInitialsEl.textContent = iv;
  }

  function authHeaders(withJSON) {
    const access = localStorage.getItem(LS_ACCESS) || "";
    const headers = {};
    if (access) headers.Authorization = "Bearer " + access;
    if (withJSON) headers["Content-Type"] = "application/json";
    return headers;
  }

  async function requestJSON(url, opts) {
    const resp = await fetch(url, opts);
    const text = await resp.text();
    let data = {};
    try {
      data = text ? JSON.parse(text) : {};
    } catch (_) {
      data = { raw: text };
    }
    return { resp, data };
  }

  function handleAuthFail(status) {
    if (status !== 401 && status !== 403) return false;
    clearSession();
    window.location.href = "/dev/login";
    return true;
  }

  async function fetchUsers(role, search) {
    const qs = new URLSearchParams();
    if (role) qs.set("role", role);
    if (search) qs.set("q", search);

    const { resp, data } = await requestJSON(`/admin/users?${qs.toString()}`, {
      headers: authHeaders(false),
    });

    if (!resp.ok) {
      if (handleAuthFail(resp.status)) return null;
      throw new Error(data.error || `HTTP ${resp.status}`);
    }

    return Array.isArray(data.users) ? data.users : [];
  }

  async function fetchProjects(status, search) {
    const qs = new URLSearchParams();
    if (status) qs.set("status", status);
    if (search) qs.set("q", search);

    const { resp, data } = await requestJSON(`/admin/projects?${qs.toString()}`, {
      headers: authHeaders(false),
    });

    if (!resp.ok) {
      if (handleAuthFail(resp.status)) return null;
      throw new Error(data.error || `HTTP ${resp.status}`);
    }

    return Array.isArray(data.projects) ? data.projects : [];
  }

  function roleLabel(code) {
    if (code === "STUDENT") return "Студент";
    if (code === "PROFESSOR") return "Преподаватель";
    if (code === "SUPER_ADMIN") return "Админ";
    return code || "-";
  }

  function roleClass(code) {
    if (code === "STUDENT") return "role-student";
    if (code === "PROFESSOR") return "role-professor";
    if (code === "SUPER_ADMIN") return "role-admin";
    return "role-student";
  }

  function userStatusLabel(status) {
    if (status === "ACTIVE") return "Активен";
    if (status === "PENDING") return "Ожидает";
    if (status === "DISABLED") return "Неактивен";
    return status || "-";
  }

  function userStatusClass(status) {
    if (status === "ACTIVE") return "status-active";
    if (status === "PENDING") return "status-pending";
    if (status === "DISABLED") return "status-disabled";
    return "status-pending";
  }

  function userStatusDot(status) {
    if (status === "ACTIVE") return "green";
    if (status === "DISABLED") return "red";
    return "amber";
  }

  function projectStatusLabel(status) {
    const s = String(status || "").toUpperCase();
    if (s === "DRAFT") return "Черновик";
    if (s === "REVIEW") return "На проверке";
    if (s === "RECRUITMENT") return "Набор";
    if (s === "ACTIVE") return "Активный";
    if (s === "GRADING") return "Оценивание";
    if (s === "ARCHIVE") return "Отклонен";
    return s || "-";
  }

  function projectStatusClass(status) {
    const s = String(status || "").toUpperCase();
    if (s === "REVIEW") return "review";
    if (s === "ACTIVE") return "active";
    if (s === "RECRUITMENT") return "recruitment";
    if (s === "GRADING") return "grading";
    if (s === "ARCHIVE") return "archive";
    return "draft";
  }

  function projectStatusDot(status) {
    const s = String(status || "").toUpperCase();
    if (s === "ACTIVE") return "green";
    if (s === "REVIEW") return "amber";
    if (s === "ARCHIVE") return "red";
    if (s === "RECRUITMENT") return "violet";
    return "gray";
  }

  function escapeHTML(v) {
    return String(v || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function formatDate(raw) {
    if (!raw) return "-";
    const dt = new Date(raw);
    if (Number.isNaN(dt.getTime())) return "-";
    return dt.toLocaleDateString("ru-RU");
  }

  function relativeTime(rawOrEpoch) {
    const ts = typeof rawOrEpoch === "number" ? rawOrEpoch : new Date(rawOrEpoch).getTime();
    if (!Number.isFinite(ts) || ts <= 0) return "недавно";

    const diffMs = Date.now() - ts;
    const mins = Math.floor(diffMs / 60000);
    if (mins <= 0) return "только что";
    if (mins < 60) return `${mins} мин назад`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours} ч назад`;
    const days = Math.floor(hours / 24);
    return `${days} дн назад`;
  }

  function groupLabel(user) {
    const dept = String(user.department_code || "").trim();
    const fac = String(user.faculty_code || "").trim();
    if (dept && fac) return `${dept}-${fac}`;
    if (dept) return dept;
    return "—";
  }

  function computeProjectStats(projects) {
    const list = Array.isArray(projects) ? projects : [];
    return {
      total: list.length,
      review: list.filter((p) => String(p.status || "").toUpperCase() === "REVIEW").length,
      approved: list.filter((p) => String(p.status || "").toUpperCase() === "ACTIVE").length,
      rejected: list.filter((p) => String(p.status || "").toUpperCase() === "ARCHIVE").length,
    };
  }

  function renderUserStats(users) {
    const list = Array.isArray(users) ? users : [];
    const total = list.length;
    const students = list.filter((u) => u.role_code === "STUDENT").length;
    const professors = list.filter((u) => u.role_code === "PROFESSOR").length;
    const disabled = list.filter((u) => u.status === "DISABLED").length;

    statTotalUsersEl.textContent = String(total);
    statStudentsEl.textContent = String(students);
    statProfessorsEl.textContent = String(professors);
    statDisabledEl.textContent = String(disabled);
  }

  function renderProjectStats(projects) {
    const stats = computeProjectStats(projects);

    statTotalProjectsEl.textContent = String(stats.total);
    statReviewProjectsEl.textContent = String(stats.review);
    statApprovedProjectsEl.textContent = String(stats.approved);
    statRejectedProjectsEl.textContent = String(stats.rejected);

    projTotalEl.textContent = String(stats.total);
    projReviewEl.textContent = String(stats.review);
    projApprovedEl.textContent = String(stats.approved);
    projRejectedEl.textContent = String(stats.rejected);
  }

  function renderActivity(users, projects, query) {
    const q = String(query || "").trim().toLowerCase();
    const entries = [];

    const userList = Array.isArray(users) ? users.slice(0, 6) : [];
    userList.forEach((u, i) => {
      entries.push({
        kind: "user",
        dot: u.status === "DISABLED" ? "amber" : "violet",
        text: `${u.full_name || u.email || "Пользователь"} · ${roleLabel(u.role_code)}`,
        at: Date.now() - i * 5 * 60 * 1000,
      });
    });

    const projectList = Array.isArray(projects) ? projects.slice(0, 6) : [];
    projectList.forEach((p) => {
      entries.push({
        kind: "project",
        dot: projectStatusDot(p.status),
        text: `${p.title || "Проект"} · ${projectStatusLabel(p.status)}`,
        at: new Date(p.updated_at || p.created_at || "").getTime() || 0,
      });
    });

    const list = entries
      .filter((item) => (q ? item.text.toLowerCase().includes(q) : true))
      .sort((a, b) => b.at - a.at)
      .slice(0, 8);

    activityListEl.innerHTML = "";
    if (list.length === 0) {
      activityListEl.innerHTML = "<li><div class=\"activity-title\"><i class=\"dot gray\"></i>Нет данных для отображения</div><span class=\"activity-time\">—</span></li>";
      return;
    }

    list.forEach((item) => {
      const li = document.createElement("li");
      li.innerHTML = `
        <div class="activity-title">
          <i class="dot ${item.dot}"></i>
          <span>${escapeHTML(item.text)}</span>
        </div>
        <span class="activity-time">${escapeHTML(relativeTime(item.at))}</span>
      `;
      activityListEl.appendChild(li);
    });
  }

  function renderUsers(users) {
    state.users = Array.isArray(users) ? users : [];
    usersBodyEl.innerHTML = "";

    if (state.users.length === 0) {
      usersBodyEl.innerHTML = "<tr><td colspan=\"6\">Пользователи не найдены</td></tr>";
      usersMetaEl.textContent = "Показано 0 пользователей";
      return;
    }

    for (const u of state.users) {
      const actionStatus = u.status === "ACTIVE" ? "DISABLED" : "ACTIVE";
      const actionLabel = actionStatus === "ACTIVE" ? "Активировать" : "Блокировать";
      const avatar = initials(u.full_name || u.email || "U");

      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td><span class="check-cell" aria-hidden="true"></span></td>
        <td>
          <div class="user-cell">
            <span class="user-avatar">${escapeHTML(avatar)}</span>
            <div>
              <span class="user-name">${escapeHTML(u.full_name || "-")}</span>
              <span class="user-mail">${escapeHTML(u.email || "-")}</span>
            </div>
          </div>
        </td>
        <td><span class="pill ${roleClass(u.role_code)}">${escapeHTML(roleLabel(u.role_code))}</span></td>
        <td><span class="group-cell">${escapeHTML(groupLabel(u))}</span></td>
        <td><span class="pill ${userStatusClass(u.status)}"><i class="dot ${userStatusDot(u.status)}"></i>${escapeHTML(userStatusLabel(u.status))}</span></td>
        <td>
          <div class="row-actions">
            <button type="button" class="action-btn" data-act="set-user-status" data-id="${u.id}" data-status="${actionStatus}">${escapeHTML(actionLabel)}</button>
          </div>
        </td>
      `;
      usersBodyEl.appendChild(tr);
    }

    usersMetaEl.textContent = `Показано ${state.users.length} пользователей`;
  }

  function renderProjects(projects) {
    state.projects = Array.isArray(projects) ? projects : [];
    projectsBodyEl.innerHTML = "";

    if (state.projects.length === 0) {
      projectsBodyEl.innerHTML = "<tr><td colspan=\"6\">Проекты не найдены</td></tr>";
      projectsMetaEl.textContent = "Показано 0 проектов";
      return;
    }

    for (const p of state.projects) {
      const status = String(p.status || "").toUpperCase();
      const avatar = initials(p.author_name || p.author_email || "PR");
      const authorName = p.author_name || "Не указан";

      let actions = "";
      if (status === "REVIEW") {
        actions = `
          <button class="action-btn approve" data-project-act="set-status" data-id="${p.id}" data-next="ACTIVE">Одобрить</button>
          <button class="action-btn reject" data-project-act="set-status" data-id="${p.id}" data-next="ARCHIVE" data-title="${escapeHTML(p.title || "")}">Отклонить</button>
        `;
      } else if (status === "ARCHIVE") {
        actions = `<button class="action-btn" data-project-act="set-status" data-id="${p.id}" data-next="ACTIVE">Активировать</button>`;
      } else {
        actions = `<button class="action-btn reject" data-project-act="set-status" data-id="${p.id}" data-next="ARCHIVE" data-title="${escapeHTML(p.title || "")}">В архив</button>`;
      }

      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td><span class="project-id">${escapeHTML(String(p.id || "").slice(0, 8).toUpperCase())}</span></td>
        <td>
          <span class="project-title">${escapeHTML(p.title || "Без названия")}</span>
          <span class="project-desc">${escapeHTML((p.description || "").slice(0, 96) || "Описание отсутствует")}</span>
        </td>
        <td>
          <div class="author-cell">
            <span class="author-avatar">${escapeHTML(avatar)}</span>
            <div>
              <span class="author-name">${escapeHTML(authorName)}</span>
              <span class="author-mail">${escapeHTML(p.author_email || "-")}</span>
            </div>
          </div>
        </td>
        <td><span class="date-cell">${escapeHTML(formatDate(p.updated_at || p.created_at))}</span></td>
        <td><span class="status-badge ${projectStatusClass(status)}"><i class="dot ${projectStatusDot(status)}"></i>${escapeHTML(projectStatusLabel(status))}</span></td>
        <td><div class="row-actions">${actions}</div></td>
      `;
      projectsBodyEl.appendChild(tr);
    }

    projectsMetaEl.textContent = `Показано ${state.projects.length} проектов`;
  }

  function syncTopbarByView() {
    const isProjectsView = state.activeView === "projects";

    quickAddStudentBtnEl.hidden = isProjectsView;
    quickAddProfessorBtnEl.hidden = isProjectsView;

    if (state.activeView === "users") {
      quickSearchInputEl.placeholder = "Поиск пользователей...";
      quickSearchInputEl.value = state.usersSearch;
      return;
    }

    if (state.activeView === "projects") {
      quickSearchInputEl.placeholder = "Поиск проектов...";
      quickSearchInputEl.value = state.projectsSearch;
      return;
    }

    quickSearchInputEl.placeholder = "Поиск по активности...";
    quickSearchInputEl.value = state.dashboardSearch;
  }

  function setView(view, skipLoad) {
    const next = view === "users" || view === "projects" ? view : "dashboard";
    state.activeView = next;

    viewButtons.forEach((btn) => {
      btn.classList.toggle("active", btn.dataset.view === next);
    });

    viewEls.forEach((el) => {
      el.classList.toggle("active", el.id === `view-${next}`);
    });

    crumbCurrentEl.textContent = next;
    syncTopbarByView();

    if (skipLoad) return;

    if (next === "users") {
      void loadUsers();
    } else if (next === "projects") {
      void loadProjects();
    }
  }

  async function refreshDashboard() {
    const [users, projects] = await Promise.all([fetchUsers("", ""), fetchProjects("", "")]);
    if (users === null || projects === null) return;

    state.dashboardUsers = users;
    state.dashboardProjects = projects;

    renderUserStats(users);
    renderProjectStats(projects);
    renderActivity(users, projects, state.dashboardSearch);
  }

  async function loadUsers() {
    usersMetaEl.textContent = "Загрузка...";
    try {
      const users = await fetchUsers(state.usersRole, state.usersSearch);
      if (users === null) return;
      renderUsers(users);
    } catch (err) {
      usersBodyEl.innerHTML = `<tr><td colspan="6">Ошибка загрузки: ${escapeHTML(err.message || "unknown error")}</td></tr>`;
      usersMetaEl.textContent = "Ошибка";
    }
  }

  async function loadProjects() {
    projectsMetaEl.textContent = "Загрузка...";
    try {
      const projects = await fetchProjects(state.projectsStatus, state.projectsSearch);
      if (projects === null) return;
      renderProjects(projects);
    } catch (err) {
      projectsBodyEl.innerHTML = `<tr><td colspan="6">Ошибка загрузки: ${escapeHTML(err.message || "unknown error")}</td></tr>`;
      projectsMetaEl.textContent = "Ошибка";
    }
  }

  async function reloadAll() {
    await Promise.all([refreshDashboard(), loadUsers(), loadProjects()]);
  }

  function openCreateModal(role) {
    state.createRole = role;
    const title = role === "PROFESSOR" ? "Добавить преподавателя" : "Добавить студента";
    createTitleEl.textContent = title;
    createSubtitleEl.textContent = "Пользователь будет создан в статусе ACTIVE с привязкой к кафедре.";
    modalStatusEl.textContent = "";
    fullNameInputEl.value = "";
    emailInputEl.value = "";
    passwordInputEl.value = "";
    departmentInputEl.value = "CPI";
    modalEl.hidden = false;
    document.body.classList.add("modal-open");
  }

  function closeCreateModal() {
    modalStatusEl.textContent = "";
    modalEl.hidden = true;
    if (projectStatusModalEl.hidden) {
      document.body.classList.remove("modal-open");
    }
  }

  async function submitCreate() {
    const payload = {
      full_name: fullNameInputEl.value.trim(),
      email: emailInputEl.value.trim(),
      password: passwordInputEl.value,
      department_code: departmentInputEl.value,
    };

    if (!payload.full_name || !payload.email || !payload.password || !payload.department_code) {
      modalStatusEl.textContent = "Заполните все поля.";
      return;
    }

    submitCreateBtnEl.disabled = true;
    modalStatusEl.textContent = "Сохраняем...";

    const endpoint = state.createRole === "PROFESSOR" ? "/admin/users/professors" : "/admin/users/students";
    const { resp, data } = await requestJSON(endpoint, {
      method: "POST",
      headers: authHeaders(true),
      body: JSON.stringify(payload),
    });

    submitCreateBtnEl.disabled = false;

    if (!resp.ok) {
      if (handleAuthFail(resp.status)) return;
      modalStatusEl.textContent = data.error || `Ошибка: ${resp.status}`;
      return;
    }

    closeCreateModal();
    await reloadAll();
  }

  async function setUserStatus(userID, status) {
    const { resp, data } = await requestJSON(`/admin/users/${userID}/status`, {
      method: "PATCH",
      headers: authHeaders(true),
      body: JSON.stringify({ status }),
    });

    if (!resp.ok) {
      if (handleAuthFail(resp.status)) return;
      alert(data.error || `Ошибка смены статуса: ${resp.status}`);
      return;
    }

    await reloadAll();
  }

  async function setProjectStatus(projectID, status) {
    const { resp, data } = await requestJSON(`/admin/projects/${projectID}/status`, {
      method: "PATCH",
      headers: authHeaders(true),
      body: JSON.stringify({ status }),
    });

    if (!resp.ok) {
      if (handleAuthFail(resp.status)) return false;
      projectModalStatusEl.textContent = data.error || `Ошибка смены статуса: ${resp.status}`;
      return false;
    }

    return true;
  }

  function normalizeText(v) {
    return String(v || "").trim().toLowerCase();
  }

  function validateProjectConfirmInput() {
    if (!state.pendingProjectAction) {
      projectModalSubmitBtnEl.disabled = true;
      return;
    }

    const expected = normalizeText(state.pendingProjectAction.title);
    if (!expected) {
      projectModalSubmitBtnEl.disabled = false;
      return;
    }

    const given = normalizeText(projectConfirmInputEl.value);
    projectModalSubmitBtnEl.disabled = given !== expected;
  }

  function closeProjectStatusModal() {
    state.pendingProjectAction = null;
    projectModalStatusEl.textContent = "";
    projectConfirmInputEl.value = "";
    projectStatusModalEl.hidden = true;
    if (modalEl.hidden) {
      document.body.classList.remove("modal-open");
    }
  }

  function openProjectStatusModal(projectID, projectTitle, targetStatus) {
    state.pendingProjectAction = {
      projectID,
      title: projectTitle,
      status: targetStatus,
    };

    projectModalTitleEl.textContent = targetStatus === "ARCHIVE" ? "Предупреждение" : "Подтверждение действия";
    projectModalSubtitleEl.textContent = targetStatus === "ARCHIVE"
      ? "Проект будет переведен в архив и исключен из активного списка."
      : "Подтвердите смену статуса проекта.";
    projectModalExpectedEl.textContent = projectTitle ? `Ожидается: ${projectTitle}` : "";
    projectModalStatusEl.textContent = "";
    projectConfirmInputEl.value = "";
    projectModalSubmitBtnEl.textContent = targetStatus === "ARCHIVE" ? "Отклонить проект" : "Подтвердить";
    projectModalSubmitBtnEl.disabled = true;

    projectStatusModalEl.hidden = false;
    document.body.classList.add("modal-open");
    validateProjectConfirmInput();
    setTimeout(() => projectConfirmInputEl.focus(), 0);
  }

  async function submitProjectModal() {
    if (!state.pendingProjectAction) return;

    const expected = normalizeText(state.pendingProjectAction.title);
    const entered = normalizeText(projectConfirmInputEl.value);
    if (expected && expected !== entered) {
      projectModalStatusEl.textContent = "Название проекта не совпадает.";
      return;
    }

    projectModalSubmitBtnEl.disabled = true;
    projectModalStatusEl.textContent = "Обновляем статус...";

    const ok = await setProjectStatus(state.pendingProjectAction.projectID, state.pendingProjectAction.status);
    if (!ok) {
      validateProjectConfirmInput();
      return;
    }

    closeProjectStatusModal();
    await reloadAll();
  }

  function bindCreateButtons() {
    [quickAddStudentBtnEl, addStudentBtnEl].forEach((el) => {
      if (!el) return;
      el.addEventListener("click", () => openCreateModal("STUDENT"));
    });

    [quickAddProfessorBtnEl, addProfessorBtnEl].forEach((el) => {
      if (!el) return;
      el.addEventListener("click", () => openCreateModal("PROFESSOR"));
    });
  }

  function bindEvents() {
    viewButtons.forEach((btn) => {
      btn.addEventListener("click", () => {
        const next = btn.dataset.view || "dashboard";
        setView(next, false);
      });
    });

    if (openUsersViewBtnEl) {
      openUsersViewBtnEl.addEventListener("click", () => setView("users", false));
    }

    if (openProjectsViewBtnEl) {
      openProjectsViewBtnEl.addEventListener("click", () => setView("projects", false));
    }

    userRoleTabs.forEach((btn) => {
      btn.addEventListener("click", async () => {
        userRoleTabs.forEach((x) => x.classList.remove("active"));
        btn.classList.add("active");
        state.usersRole = btn.dataset.role || "";
        await loadUsers();
      });
    });

    projectStatusTabs.forEach((btn) => {
      btn.addEventListener("click", async () => {
        projectStatusTabs.forEach((x) => x.classList.remove("active"));
        btn.classList.add("active");
        state.projectsStatus = btn.dataset.status || "";
        await loadProjects();
      });
    });

    searchInputEl.addEventListener("input", () => {
      if (state.usersSearchTimer) clearTimeout(state.usersSearchTimer);
      state.usersSearchTimer = setTimeout(async () => {
        state.usersSearch = searchInputEl.value.trim();
        if (state.activeView === "users") {
          quickSearchInputEl.value = state.usersSearch;
        }
        await loadUsers();
      }, 220);
    });

    projectSearchInputEl.addEventListener("input", () => {
      if (state.projectsSearchTimer) clearTimeout(state.projectsSearchTimer);
      state.projectsSearchTimer = setTimeout(async () => {
        state.projectsSearch = projectSearchInputEl.value.trim();
        if (state.activeView === "projects") {
          quickSearchInputEl.value = state.projectsSearch;
        }
        await loadProjects();
      }, 220);
    });

    quickSearchInputEl.addEventListener("input", () => {
      if (state.quickSearchTimer) clearTimeout(state.quickSearchTimer);
      state.quickSearchTimer = setTimeout(async () => {
        const q = quickSearchInputEl.value.trim();

        if (state.activeView === "users") {
          state.usersSearch = q;
          searchInputEl.value = q;
          await loadUsers();
          return;
        }

        if (state.activeView === "projects") {
          state.projectsSearch = q;
          projectSearchInputEl.value = q;
          await loadProjects();
          return;
        }

        state.dashboardSearch = q;
        renderActivity(state.dashboardUsers, state.dashboardProjects, q);
      }, 220);
    });

    usersRefreshBtnEl.addEventListener("click", async () => {
      await Promise.all([refreshDashboard(), loadUsers()]);
    });

    projectsRefreshBtnEl.addEventListener("click", async () => {
      await Promise.all([refreshDashboard(), loadProjects()]);
    });

    bindCreateButtons();

    closeModalBtnEl.addEventListener("click", closeCreateModal);
    submitCreateBtnEl.addEventListener("click", submitCreate);

    modalEl.addEventListener("click", (e) => {
      if (e.target === modalEl) closeCreateModal();
    });

    projectModalCloseBtnEl.addEventListener("click", closeProjectStatusModal);
    projectModalCancelBtnEl.addEventListener("click", closeProjectStatusModal);
    projectModalSubmitBtnEl.addEventListener("click", async () => {
      await submitProjectModal();
    });

    projectStatusModalEl.addEventListener("click", (e) => {
      if (e.target === projectStatusModalEl) closeProjectStatusModal();
    });

    projectConfirmInputEl.addEventListener("input", () => {
      projectModalStatusEl.textContent = "";
      validateProjectConfirmInput();
    });

    window.addEventListener("keydown", (e) => {
      if (e.key === "Escape") {
        if (!projectStatusModalEl.hidden) {
          closeProjectStatusModal();
          return;
        }
        if (!modalEl.hidden) {
          closeCreateModal();
        }
      }
    });

    usersBodyEl.addEventListener("click", async (e) => {
      const target = e.target;
      if (!(target instanceof HTMLElement)) return;
      if (target.dataset.act !== "set-user-status") return;
      const userID = target.dataset.id || "";
      const status = target.dataset.status || "";
      if (!userID || !status) return;
      await setUserStatus(userID, status);
    });

    projectsBodyEl.addEventListener("click", async (e) => {
      const target = e.target;
      if (!(target instanceof HTMLElement)) return;
      if (target.dataset.projectAct !== "set-status") return;

      const projectID = target.dataset.id || "";
      const nextStatus = String(target.dataset.next || "").toUpperCase();
      const projectTitle = target.dataset.title || "";

      if (!projectID || !nextStatus) return;

      if (nextStatus === "ARCHIVE") {
        openProjectStatusModal(projectID, projectTitle, nextStatus);
        return;
      }

      const ok = await setProjectStatus(projectID, nextStatus);
      if (!ok) return;
      await reloadAll();
    });

    logoutBtnEl.addEventListener("click", () => {
      clearSession();
      window.location.href = "/dev/login";
    });
  }

  async function main() {
    const claims = ensureAdminSession();
    if (!claims) return;

    hydrateProfile();
    bindEvents();
    setView("dashboard", true);

    await reloadAll();
  }

  main();
})();
