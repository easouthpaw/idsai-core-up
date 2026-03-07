(() => {
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_SELECTED_PROJECT = "idsai_selected_project";

  const LS_PROJECT_META_PREFIX = "idsai_project_meta:";
  const LS_TASK_META_PREFIX = "idsai_task_meta:";
  const LS_MEMBER_PERMS_PREFIX = "idsai_member_perms:";

  const PERMISSION_GROUPS = [
    {
      title: "Команда",
      items: ["team.view", "team.kick", "team.invite", "team.role.assign"],
    },
    {
      title: "Задачи",
      items: ["task.create", "task.update.own", "task.status.change", "task.assign", "task.update.any"],
    },
    {
      title: "Сдачи",
      items: ["submission.create", "submission.edit.own", "submission.review", "submission.approve"],
    },
    {
      title: "Комментарии",
      items: ["comment.create", "comment.edit.own", "comment.delete.own"],
    },
  ];

  const ROLE_PRESETS = {
    TEAM_LEAD: new Set(PERMISSION_GROUPS.flatMap((g) => g.items)),
    DEVELOPER: new Set(["team.view", "task.create", "task.update.own", "task.status.change", "submission.create", "comment.create", "comment.edit.own"]),
    FRONTEND: new Set(["team.view", "task.create", "task.update.own", "task.status.change", "submission.create", "comment.create"]),
    BACKEND: new Set(["team.view", "task.create", "task.update.own", "task.status.change", "submission.create", "comment.create"]),
    QA: new Set(["team.view", "task.update.own", "task.status.change", "submission.review", "comment.create"]),
    DESIGNER: new Set(["team.view", "task.create", "task.update.own", "submission.create", "comment.create"]),
  };

  const ui = {
    profileAvatar: document.getElementById("profileAvatar"),
    studentName: document.getElementById("studentName"),
    studentEmail: document.getElementById("studentEmail"),

    crumbProject: document.getElementById("crumbProject"),
    title: document.getElementById("projectTitle"),
    statusBadge: document.getElementById("statusBadge"),
    visibilityLabel: document.getElementById("visibilityLabel"),
    projectID: document.getElementById("projectID"),
    repoLink: document.getElementById("repoLink"),

    pageNotice: document.getElementById("pageNotice"),
    globalSearchInput: document.getElementById("globalSearchInput"),

    tabButtons: Array.from(document.querySelectorAll(".tab-btn")),
    switchViewButtons: Array.from(document.querySelectorAll("[data-switch-view]")),

    viewOverview: document.getElementById("view-overview"),
    viewTeam: document.getElementById("view-team"),
    viewTasks: document.getElementById("view-tasks"),
    viewReview: document.getElementById("view-review"),
    viewEdit: document.getElementById("view-edit"),

    aboutContent: document.getElementById("aboutContent"),
    stackChips: document.getElementById("stackChips"),
    teamMiniList: document.getElementById("teamMiniList"),
    readinessList: document.getElementById("readinessList"),
    activityList: document.getElementById("activityList"),

    openRecruitmentBtn: document.getElementById("openRecruitmentBtn"),
    applyMemberBtn: document.getElementById("applyMemberBtn"),
    professorIDInput: document.getElementById("professorIDInput"),
    assignProfessorBtn: document.getElementById("assignProfessorBtn"),
    approveProjectBtn: document.getElementById("approveProjectBtn"),

    positionForm: document.getElementById("positionForm"),
    positionNameInput: document.getElementById("positionNameInput"),
    positionCodeInput: document.getElementById("positionCodeInput"),
    positionCapacityInput: document.getElementById("positionCapacityInput"),
    teamTableBody: document.getElementById("teamTableBody"),

    progressBadge: document.getElementById("progressBadge"),
    openTaskModalBtn: document.getElementById("openTaskModalBtn"),
    todoTasks: document.getElementById("todoTasks"),
    doingTasks: document.getElementById("doingTasks"),
    doneTasks: document.getElementById("doneTasks"),
    countTodo: document.getElementById("countTodo"),
    countDoing: document.getElementById("countDoing"),
    countDone: document.getElementById("countDone"),
    tasksTeamList: document.getElementById("tasksTeamList"),

    activeProgressWrap: document.getElementById("activeProgressWrap"),
    progressPercent: document.getElementById("progressPercent"),
    progressFill: document.getElementById("progressFill"),
    stackInfoConsole: document.getElementById("stackInfoConsole"),

    favoriteBtn: document.getElementById("favoriteBtn"),
    openEditViewBtn: document.getElementById("openEditViewBtn"),
    closeEditViewBtn: document.getElementById("closeEditViewBtn"),
    cancelEditBtn: document.getElementById("cancelEditBtn"),
    saveProjectBtn: document.getElementById("saveProjectBtn"),

    editTitleInput: document.getElementById("editTitleInput"),
    editShortDescriptionInput: document.getElementById("editShortDescriptionInput"),
    editReadmeInput: document.getElementById("editReadmeInput"),
    editRepoInput: document.getElementById("editRepoInput"),
    editStacksInput: document.getElementById("editStacksInput"),
    editVisibilityPublic: document.getElementById("editVisibilityPublic"),
    editVisibilityPrivate: document.getElementById("editVisibilityPrivate"),
    editStackChips: document.getElementById("editStackChips"),

    taskModal: document.getElementById("taskModal"),
    taskModalTitleInput: document.getElementById("taskModalTitleInput"),
    taskModalStatusSelect: document.getElementById("taskModalStatusSelect"),
    taskModalPrioritySelect: document.getElementById("taskModalPrioritySelect"),
    taskModalPositionSelect: document.getElementById("taskModalPositionSelect"),
    taskModalAssigneeSelect: document.getElementById("taskModalAssigneeSelect"),
    taskModalTagsInput: document.getElementById("taskModalTagsInput"),
    taskModalDescriptionInput: document.getElementById("taskModalDescriptionInput"),
    taskModalDueAtInput: document.getElementById("taskModalDueAtInput"),
    taskModalCreateBtn: document.getElementById("taskModalCreateBtn"),

    permissionsModal: document.getElementById("permissionsModal"),
    permMemberName: document.getElementById("permMemberName"),
    permRoleSelect: document.getElementById("permRoleSelect"),
    permChecklist: document.getElementById("permChecklist"),
    savePermissionsBtn: document.getElementById("savePermissionsBtn"),

    refreshBtn: document.getElementById("refreshBtn"),
    logoutBtn: document.getElementById("logoutBtn"),
  };

  const state = {
    projectID: "",
    project: null,
    stacks: [],
    positions: [],
    members: [],
    readiness: null,
    tasks: [],
    activeView: "overview",
    searchQuery: "",
    projectMeta: {},
    taskMeta: {},
    memberPerms: {},
    currentPermUserID: "",
    noticeTimer: null,
    favorite: false,
  };

  function escapeHTML(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function slugify(v) {
    return String(v || "")
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "") || "project";
  }

  function shortID(v) {
    const s = String(v || "");
    if (!s) return "-";
    if (s.length <= 8) return s;
    return `${s.slice(0, 8)}...${s.slice(-4)}`;
  }

  function initials(name, fallback) {
    const n = String(name || "").trim();
    if (n) {
      const parts = n.split(/\s+/).filter(Boolean);
      if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
      return n.slice(0, 2).toUpperCase();
    }
    const f = String(fallback || "").trim();
    return f ? f.slice(0, 2).toUpperCase() : "ST";
  }

  function formatDate(raw) {
    if (!raw) return "-";
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) return String(raw);
    return d.toLocaleString();
  }

  function parseJSON(raw, fallback) {
    try {
      return JSON.parse(raw);
    } catch (_) {
      return fallback;
    }
  }

  function loadJSON(key, fallback) {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    return parseJSON(raw, fallback);
  }

  function saveJSON(key, value) {
    localStorage.setItem(key, JSON.stringify(value));
  }

  function projectMetaKey() {
    return `${LS_PROJECT_META_PREFIX}${state.projectID}`;
  }

  function taskMetaKey() {
    return `${LS_TASK_META_PREFIX}${state.projectID}`;
  }

  function memberPermsKey() {
    return `${LS_MEMBER_PERMS_PREFIX}${state.projectID}`;
  }

  function decodePayload(token) {
    const parts = token.split(".");
    if (parts.length < 2) throw new Error("invalid JWT");
    let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const mod = payload.length % 4;
    if (mod > 0) payload += "=".repeat(4 - mod);
    return JSON.parse(atob(payload));
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

  async function loadOptional(name, method, url, fallback) {
    try {
      return await request(method, url);
    } catch (err) {
      console.warn(`[project-ui] ${name} failed`, err);
      return fallback;
    }
  }

  function clearSession() {
    localStorage.removeItem(LS_ACCESS);
    localStorage.removeItem(LS_REFRESH);
    localStorage.removeItem(LS_USER);
    localStorage.removeItem(LS_FACULTY);
    localStorage.removeItem(LS_STUDENT_NAME);
    localStorage.removeItem(LS_STUDENT_EMAIL);
    localStorage.removeItem(LS_SELECTED_PROJECT);
  }

  function setNotice(message, isError) {
    if (!message) {
      ui.pageNotice.hidden = true;
      ui.pageNotice.textContent = "";
      ui.pageNotice.classList.remove("err");
      return;
    }

    ui.pageNotice.hidden = false;
    ui.pageNotice.textContent = message;
    ui.pageNotice.classList.toggle("err", Boolean(isError));

    if (state.noticeTimer) {
      clearTimeout(state.noticeTimer);
    }
    state.noticeTimer = setTimeout(() => setNotice("", false), 4200);
  }

  function projectIDFromPath() {
    const parts = window.location.pathname.split("/").filter(Boolean);
    return parts.length >= 3 ? parts[2] : "";
  }

  function projectStatusCode() {
    return String(state.project?.status || "").toUpperCase();
  }

  function isProjectActive() {
    return projectStatusCode() === "ACTIVE";
  }

  function statusPresentation(status) {
    const s = String(status || "").toUpperCase();
    if (s === "ACTIVE") return { label: "В работе", cls: "active" };
    if (s === "REVIEW" || s === "GRADING") return { label: "На ревью", cls: "review" };
    if (s === "RECRUITMENT") return { label: "Набор", cls: "muted" };
    if (s === "ARCHIVE") return { label: "Архив", cls: "muted" };
    return { label: "Черновик", cls: "muted" };
  }

  function visibilityLabel(value) {
    return String(value || "").toUpperCase() === "PUBLIC" ? "Публичный" : "Приватный";
  }

  function getRepoURL() {
    const custom = String(state.projectMeta.repo || "").trim();
    if (custom) return custom;
    return `https://github.com/idsai-corp/${slugify(state.project?.title)}`;
  }

  function getReadmeText() {
    const custom = String(state.projectMeta.readme || "").trim();
    if (custom) return custom;
    return String(state.project?.description || "").trim();
  }

  function getDisplayName(userID) {
    const currentUser = localStorage.getItem(LS_USER) || "";
    if (String(userID) === String(currentUser)) {
      return localStorage.getItem(LS_STUDENT_NAME) || "Текущий студент";
    }
    return `Student ${shortID(userID)}`;
  }

  function getRoleLabel(member) {
    if (member.position_name) return member.position_name;
    if (member.position_code) return member.position_code;
    return "Без роли";
  }

  function activeMembers() {
    return state.members.filter((m) => String(m.status || "").toUpperCase() === "ACTIVE");
  }

  function membersByPosition(positionID) {
    return activeMembers().filter((m) => String(m.position_id || "") === String(positionID));
  }

  function toRFC3339(localDateTime) {
    if (!localDateTime) return null;
    const d = new Date(localDateTime);
    if (Number.isNaN(d.getTime())) return null;
    return d.toISOString();
  }

  function parseStacks(input) {
    const seen = new Set();
    const out = [];

    String(input || "")
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean)
      .forEach((s) => {
        const upper = s.toUpperCase();
        if (!seen.has(upper)) {
          seen.add(upper);
          out.push(upper);
        }
      });

    return out;
  }

  function projectPositionOptions(selectedID) {
    if (state.positions.length === 0) {
      return '<option value="">Нет ролей</option>';
    }

    const options = ['<option value="">Выберите роль</option>'];
    state.positions.forEach((p) => {
      const selected = String(p.id) === String(selectedID) ? "selected" : "";
      options.push(`<option value="${escapeHTML(p.id)}" ${selected}>${escapeHTML(p.name)} (${escapeHTML(p.code)})</option>`);
    });
    return options.join("");
  }

  function assigneeOptions(positionID, selectedUserID) {
    const members = membersByPosition(positionID);
    const options = ['<option value="">Без исполнителя</option>'];

    members.forEach((m) => {
      const selected = String(m.user_id) === String(selectedUserID) ? "selected" : "";
      options.push(`<option value="${escapeHTML(m.user_id)}" ${selected}>${escapeHTML(getDisplayName(m.user_id))}</option>`);
    });

    return options.join("");
  }

  function taskTags(task) {
    const meta = state.taskMeta[task.id] || {};
    const tags = [];
    if (task.position_code) tags.push(task.position_code);
    if (meta.priority) tags.push(meta.priority);
    if (Array.isArray(meta.tags)) {
      meta.tags.slice(0, 2).forEach((t) => tags.push(String(t).toUpperCase()));
    }
    return tags;
  }

  function permissionPresetForRole(roleCode) {
    const preset = ROLE_PRESETS[roleCode];
    if (!preset) return new Set();
    return new Set(preset);
  }

  function defaultPermissionState(member) {
    const role = String(member?.status || "").toUpperCase() === "ACTIVE" ? "DEVELOPER" : "QA";
    return {
      role,
      permissions: Array.from(permissionPresetForRole(role)),
    };
  }

  function ensureMemberPermissionState(userID, member) {
    const existing = state.memberPerms[userID];
    if (existing && Array.isArray(existing.permissions)) {
      return existing;
    }

    const def = defaultPermissionState(member);
    state.memberPerms[userID] = def;
    return def;
  }

  function bindProfile() {
    const name = localStorage.getItem(LS_STUDENT_NAME) || "Student";
    const email = localStorage.getItem(LS_STUDENT_EMAIL) || "student@university.edu";

    ui.studentName.textContent = name;
    ui.studentEmail.textContent = email;
    ui.profileAvatar.textContent = initials(name, email);
  }

  function bindProjectMetaToUI() {
    ui.editTitleInput.value = state.project?.title || "";
    ui.editShortDescriptionInput.value = state.project?.description || "";
    ui.editReadmeInput.value = getReadmeText();
    ui.editRepoInput.value = getRepoURL();
    ui.editStacksInput.value = state.stacks.map((s) => s.code).join(", ");

    const currentVisibility = String(state.project?.visibility || "").toUpperCase();
    const forced = String(state.projectMeta.visibility || "").toUpperCase();
    const visibility = forced || (currentVisibility === "PUBLIC" ? "PUBLIC" : "PRIVATE");
    ui.editVisibilityPublic.checked = visibility === "PUBLIC";
    ui.editVisibilityPrivate.checked = visibility !== "PUBLIC";

    renderEditStackChips();
  }

  function renderHero() {
    if (!state.project) return;

    const status = statusPresentation(state.project.status);
    document.title = `IDSAI Corp. · ${state.project.title || "project"}`;
    ui.title.textContent = state.project.title || "project";
    ui.crumbProject.textContent = state.project.title || "project";
    ui.statusBadge.className = `status-pill ${status.cls}`;
    ui.statusBadge.textContent = status.label;
    ui.visibilityLabel.textContent = visibilityLabel(state.project.visibility);
    ui.projectID.textContent = `id: ${state.project.id || "-"}`;

    const repo = getRepoURL();
    ui.repoLink.href = repo;
    ui.repoLink.textContent = repo.replace(/^https?:\/\//, "");

    ui.favoriteBtn.textContent = state.favorite ? "★ В избранном" : "★ В избранное";
  }

  function renderAbout() {
    const description = String(state.project?.description || "").trim();
    const readme = String(getReadmeText() || description || "Описание проекта пока не заполнено.").trim();

    const bullets = [
      "Ролевой доступ и управление командой",
      "Контроль задач через Kanban-колонки",
      "Критерии преподавателя и этапы запуска",
      "Прогресс проекта после перехода в ACTIVE",
    ];

    const html =
      `<p><strong>${escapeHTML(state.project?.title || "Project")}</strong> — ${escapeHTML(description || "Описание отсутствует.")}</p>` +
      `<p>${escapeHTML(readme).replaceAll("\n", "<br>")}</p>` +
      `<p><strong>Ключевые особенности</strong></p>` +
      `<ul>${bullets.map((b) => `<li>${escapeHTML(b)}</li>`).join("")}</ul>` +
      `<p class="muted-text">Последнее обновление: ${escapeHTML(formatDate(state.project?.updated_at))}</p>`;

    ui.aboutContent.innerHTML = html;
  }

  function renderStackChips() {
    ui.stackChips.innerHTML = "";
    if (state.stacks.length === 0) {
      ui.stackChips.innerHTML = '<div class="empty-state">Стек пока не заполнен.</div>';
      return;
    }

    state.stacks.forEach((stack) => {
      const chip = document.createElement("span");
      chip.className = "chip";
      chip.textContent = stack.code;
      ui.stackChips.appendChild(chip);
    });
  }

  function renderTeamMini() {
    const members = activeMembers().slice(0, 5);
    ui.teamMiniList.innerHTML = "";

    if (members.length === 0) {
      ui.teamMiniList.innerHTML = '<div class="empty-state">Активных участников пока нет.</div>';
      return;
    }

    members.forEach((m) => {
      const row = document.createElement("div");
      row.className = "mini-team-row";
      row.innerHTML =
        `<div class="member-avatar">${escapeHTML(initials(getDisplayName(m.user_id), m.user_id))}</div>` +
        `<div><p class="member-name">${escapeHTML(getDisplayName(m.user_id))}</p><p class="member-role">${escapeHTML(getRoleLabel(m))}</p></div>`;
      ui.teamMiniList.appendChild(row);
    });
  }

  function renderReadiness() {
    ui.readinessList.innerHTML = "";

    if (!state.readiness) {
      ui.readinessList.innerHTML = '<div class="empty-state">Данные о готовности не загружены.</div>';
      ui.approveProjectBtn.disabled = true;
      return;
    }

    const items = [
      ["Статус", state.readiness.status],
      ["Участники", `${state.readiness.active_members}/${state.readiness.required_members}`],
      ["Преподаватель", state.readiness.has_professor ? "Назначен" : "Не назначен"],
      ["Критерии", String(state.readiness.criteria_count)],
      ["Можно активировать", state.readiness.can_activate ? "Да" : "Нет"],
    ];

    items.forEach(([k, v]) => {
      const row = document.createElement("div");
      row.className = "readiness-item";
      row.innerHTML = `<span>${escapeHTML(k)}</span><strong>${escapeHTML(v)}</strong>`;
      ui.readinessList.appendChild(row);
    });

    ui.approveProjectBtn.disabled = !state.readiness.can_activate;
  }

  function renderActivity() {
    const stats = {
      todo: state.tasks.filter((t) => String(t.status || "").toUpperCase() === "OPEN").length,
      doing: state.tasks.filter((t) => String(t.status || "").toUpperCase() === "IN_PROGRESS").length,
      done: state.tasks.filter((t) => String(t.status || "").toUpperCase() === "DONE").length,
    };

    const feed = [
      { title: `Статус проекта: ${projectStatusCode() || "DRAFT"}`, when: formatDate(state.project?.updated_at) },
      { title: `Задачи: ${stats.todo} todo / ${stats.doing} in progress / ${stats.done} done`, when: "текущее состояние" },
      { title: `Участники: ${activeMembers().length} активных`, when: "команда" },
    ];

    ui.activityList.innerHTML = "";
    feed.forEach((item) => {
      const li = document.createElement("li");
      li.innerHTML = `<strong>${escapeHTML(item.title)}</strong><span>${escapeHTML(item.when)}</span>`;
      ui.activityList.appendChild(li);
    });
  }

  function renderTeamTable() {
    const query = state.searchQuery;
    const filtered = state.members.filter((m) => {
      if (!query) return true;
      const hay = `${getDisplayName(m.user_id)} ${m.user_id} ${m.position_name || ""} ${m.position_code || ""}`.toLowerCase();
      return hay.includes(query);
    });

    ui.teamTableBody.innerHTML = "";

    if (filtered.length === 0) {
      ui.teamTableBody.innerHTML = '<tr><td colspan="5"><div class="empty-state">Нет участников под текущий фильтр.</div></td></tr>';
      return;
    }

    filtered.forEach((m) => {
      const status = String(m.status || "").toUpperCase();
      const github = `https://github.com/${slugify(getDisplayName(m.user_id))}`;
      const roleOptions = projectPositionOptions(m.position_id || "");

      const row = document.createElement("tr");
      row.setAttribute("data-user-id", m.user_id || "");
      row.innerHTML =
        `<td>` +
          `<div class="user-cell">` +
            `<div class="member-avatar">${escapeHTML(initials(getDisplayName(m.user_id), m.user_id))}</div>` +
            `<div><strong>${escapeHTML(getDisplayName(m.user_id))}</strong><small>${escapeHTML(shortID(m.user_id))}</small></div>` +
          `</div>` +
        `</td>` +
        `<td><span class="status-badge ${status.toLowerCase()}">${escapeHTML(status)}</span></td>` +
        `<td><select class="member-role-select">${roleOptions}</select></td>` +
        `<td><a class="meta-link" href="${escapeHTML(github)}" target="_blank" rel="noreferrer">${escapeHTML(github.replace("https://", ""))}</a></td>` +
        `<td>` +
          `<div class="task-toolbar">` +
            `<button class="ghost-btn" data-member-action="approve" ${status === "APPLIED" && state.positions.length > 0 ? "" : "disabled"}>Одобрить</button>` +
            `<button class="ghost-btn" data-member-action="set-position" ${status === "ACTIVE" && state.positions.length > 0 ? "" : "disabled"}>Сменить роль</button>` +
            `<button class="ghost-btn" data-member-action="permissions">Права</button>` +
          `</div>` +
        `</td>`;

      ui.teamTableBody.appendChild(row);
    });
  }

  function renderProgress() {
    const isActive = isProjectActive();
    const doneCount = state.tasks.filter((t) => String(t.status || "").toUpperCase() === "DONE").length;
    const total = state.tasks.length;
    const percent = total > 0 ? Math.round((doneCount * 100) / total) : 0;

    ui.activeProgressWrap.hidden = !isActive;
    ui.progressBadge.className = `status-pill ${isActive ? "active" : "muted"}`;
    ui.progressBadge.textContent = isActive ? `Прогресс открыт · ${percent}%` : "Прогресс закрыт до ACTIVE";

    if (isActive) {
      ui.progressPercent.textContent = `${percent}%`;
      ui.progressFill.style.width = `${percent}%`;
    }

    ui.openTaskModalBtn.disabled = !isActive;
    ui.openTaskModalBtn.title = isActive ? "" : "Создание задач доступно только после ACTIVE";
  }

  function renderTasksTeam() {
    const members = activeMembers();
    ui.tasksTeamList.innerHTML = "";
    if (members.length === 0) {
      ui.tasksTeamList.innerHTML = '<div class="empty-state">Команда не набрана.</div>';
      return;
    }

    members.forEach((m) => {
      const row = document.createElement("div");
      row.className = "mini-team-row";
      row.innerHTML =
        `<div class="member-avatar">${escapeHTML(initials(getDisplayName(m.user_id), m.user_id))}</div>` +
        `<div><p class="member-name">${escapeHTML(getDisplayName(m.user_id))}</p><p class="member-role">${escapeHTML(getRoleLabel(m))}</p></div>`;
      ui.tasksTeamList.appendChild(row);
    });
  }

  function renderStackInfoConsole() {
    const data = {
      status: projectStatusCode() || "DRAFT",
      stacks: state.stacks.map((s) => s.code),
      members_active: activeMembers().length,
      tasks_total: state.tasks.length,
      criteria_count: state.readiness ? state.readiness.criteria_count : 0,
      readiness: state.readiness
        ? {
            can_activate: state.readiness.can_activate,
            active: state.readiness.active_members,
            required: state.readiness.required_members,
          }
        : null,
    };
    ui.stackInfoConsole.textContent = JSON.stringify(data, null, 2);
  }

  function taskColumn(status) {
    const s = String(status || "").toUpperCase();
    if (s === "DONE") return "done";
    if (s === "IN_PROGRESS") return "doing";
    return "todo";
  }

  function createTaskCard(task) {
    const card = document.createElement("article");
    card.className = "task-item";
    card.setAttribute("data-task-id", task.id || "");

    const status = String(task.status || "OPEN").toUpperCase();
    const tags = taskTags(task);

    card.innerHTML =
      `<h4>${escapeHTML(task.title || "Без названия")}</h4>` +
      `<p>${escapeHTML(task.description || "Описание отсутствует")}</p>` +
      `<p>Роль: ${escapeHTML(task.position_name || task.position_code || "-")}</p>` +
      `<p>Исполнитель: ${escapeHTML(task.assignee_user_id ? getDisplayName(task.assignee_user_id) : "не назначен")}</p>` +
      `<div class="task-tags">${tags.map((t) => `<span class="tag">${escapeHTML(t)}</span>`).join("")}</div>` +
      `<div class="task-controls">` +
        `<div class="task-control-row">` +
          `<select data-task-status>` +
            `<option value="OPEN" ${status === "OPEN" ? "selected" : ""}>OPEN</option>` +
            `<option value="IN_PROGRESS" ${status === "IN_PROGRESS" ? "selected" : ""}>IN_PROGRESS</option>` +
            `<option value="DONE" ${status === "DONE" ? "selected" : ""}>DONE</option>` +
          `</select>` +
          `<button class="ghost-btn" data-task-action="status">Статус</button>` +
        `</div>` +
        `<div class="task-control-row">` +
          `<select data-task-assignee>${assigneeOptions(task.position_id, task.assignee_user_id || "")}</select>` +
          `<button class="ghost-btn" data-task-action="assign">Назначить</button>` +
        `</div>` +
        `<button class="primary-btn" data-task-action="claim">Взять</button>` +
      `</div>`;

    return card;
  }

  function renderTasks() {
    const query = state.searchQuery;
    const filtered = state.tasks.filter((t) => {
      if (!query) return true;
      const hay = `${t.title || ""} ${t.description || ""} ${t.position_name || ""} ${t.position_code || ""}`.toLowerCase();
      return hay.includes(query);
    });

    const todo = filtered.filter((t) => taskColumn(t.status) === "todo");
    const doing = filtered.filter((t) => taskColumn(t.status) === "doing");
    const done = filtered.filter((t) => taskColumn(t.status) === "done");

    ui.countTodo.textContent = String(todo.length);
    ui.countDoing.textContent = String(doing.length);
    ui.countDone.textContent = String(done.length);

    ui.todoTasks.innerHTML = "";
    ui.doingTasks.innerHTML = "";
    ui.doneTasks.innerHTML = "";

    const renderList = (container, items) => {
      if (items.length === 0) {
        container.innerHTML = '<div class="empty-state">Пока пусто</div>';
        return;
      }
      items.forEach((task) => container.appendChild(createTaskCard(task)));
    };

    renderList(ui.todoTasks, todo);
    renderList(ui.doingTasks, doing);
    renderList(ui.doneTasks, done);

    renderProgress();
    renderTasksTeam();
    renderStackInfoConsole();
  }

  function renderOverview() {
    renderAbout();
    renderStackChips();
    renderTeamMini();
    renderReadiness();
    renderActivity();
  }

  function renderEditStackChips() {
    const stacks = parseStacks(ui.editStacksInput.value);
    ui.editStackChips.innerHTML = "";
    if (stacks.length === 0) {
      ui.editStackChips.innerHTML = '<div class="empty-state">Добавьте хотя бы один стек.</div>';
      return;
    }

    stacks.forEach((s) => {
      const chip = document.createElement("span");
      chip.className = "chip";
      chip.textContent = s;
      ui.editStackChips.appendChild(chip);
    });
  }

  function renderAll() {
    renderHero();
    renderOverview();
    renderTeamTable();
    renderTasks();
    bindProjectMetaToUI();
    renderTaskModalSelects();
  }

  function setView(viewName) {
    const target = ["overview", "team", "tasks", "review", "edit"].includes(viewName)
      ? viewName
      : "overview";

    state.activeView = target;

    const viewMap = {
      overview: ui.viewOverview,
      team: ui.viewTeam,
      tasks: ui.viewTasks,
      review: ui.viewReview,
      edit: ui.viewEdit,
    };

    Object.entries(viewMap).forEach(([key, el]) => {
      el.classList.toggle("active", key === target);
    });

    ui.tabButtons.forEach((btn) => {
      const view = btn.getAttribute("data-view");
      btn.classList.toggle("active", view === target);
    });

    if (target === "edit") {
      bindProjectMetaToUI();
    }
  }

  function openModal(modal) {
    modal.hidden = false;
    document.body.style.overflow = "hidden";
  }

  function closeModal(modal) {
    modal.hidden = true;
    if (ui.taskModal.hidden && ui.permissionsModal.hidden) {
      document.body.style.overflow = "";
    }
  }

  function renderTaskModalSelects() {
    ui.taskModalPositionSelect.innerHTML = projectPositionOptions("");
    syncTaskModalAssignees();
  }

  function syncTaskModalAssignees() {
    const positionID = ui.taskModalPositionSelect.value;
    ui.taskModalAssigneeSelect.innerHTML = assigneeOptions(positionID, "");
  }

  function openTaskModal() {
    if (!isProjectActive()) {
      setNotice("Создание задач доступно только после перевода проекта в ACTIVE.", true);
      return;
    }
    renderTaskModalSelects();
    openModal(ui.taskModal);
  }

  function collectTaskMeta() {
    const tags = String(ui.taskModalTagsInput.value || "")
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean)
      .slice(0, 5);

    return {
      priority: String(ui.taskModalPrioritySelect.value || "MEDIUM").toUpperCase(),
      tags,
    };
  }

  function clearTaskModalForm() {
    ui.taskModalTitleInput.value = "";
    ui.taskModalStatusSelect.value = "OPEN";
    ui.taskModalPrioritySelect.value = "MEDIUM";
    ui.taskModalDescriptionInput.value = "";
    ui.taskModalTagsInput.value = "";
    ui.taskModalDueAtInput.value = "";
    ui.taskModalPositionSelect.value = "";
    syncTaskModalAssignees();
  }

  async function createTaskFromModal() {
    const status = String(ui.taskModalStatusSelect.value || "OPEN").toUpperCase();
    const title = ui.taskModalTitleInput.value.trim();
    const description = ui.taskModalDescriptionInput.value.trim();
    const positionID = ui.taskModalPositionSelect.value;
    let assigneeUserID = ui.taskModalAssigneeSelect.value;

    if (!title || !positionID) {
      throw new Error("Заполните название задачи и роль.");
    }

    if (status === "IN_PROGRESS" && !assigneeUserID) {
      const possible = membersByPosition(positionID);
      if (possible.length > 0) {
        assigneeUserID = possible[0].user_id;
      } else {
        throw new Error("Для статуса IN_PROGRESS нужен исполнитель этой роли.");
      }
    }

    const payload = {
      title,
      description,
      position_id: positionID,
    };

    if (assigneeUserID) payload.assignee_user_id = assigneeUserID;

    const dueAt = toRFC3339(ui.taskModalDueAtInput.value);
    if (dueAt) payload.due_at = dueAt;

    const created = await request("POST", `/v2/projects/${state.projectID}/tasks`, payload);
    const meta = collectTaskMeta();

    if (created && created.id) {
      state.taskMeta[created.id] = meta;
      saveJSON(taskMetaKey(), state.taskMeta);
    }

    closeModal(ui.taskModal);
    clearTaskModalForm();
    setNotice("Задача создана.", false);
    await refreshData();
  }

  function openPermissionsModal(userID) {
    const member = state.members.find((m) => String(m.user_id) === String(userID));
    if (!member) return;

    state.currentPermUserID = userID;
    const st = ensureMemberPermissionState(userID, member);

    ui.permMemberName.textContent = getDisplayName(userID);
    ui.permRoleSelect.value = st.role;

    renderPermissionChecklist(new Set(st.permissions));
    openModal(ui.permissionsModal);
  }

  function renderPermissionChecklist(selected) {
    ui.permChecklist.innerHTML = "";

    PERMISSION_GROUPS.forEach((group) => {
      const wrap = document.createElement("section");
      wrap.className = "perm-group";
      wrap.innerHTML = `<h4>${escapeHTML(group.title)}</h4>`;

      const options = document.createElement("div");
      options.className = "perm-options";

      group.items.forEach((code) => {
        const label = document.createElement("label");
        const checked = selected.has(code) ? "checked" : "";
        label.innerHTML = `<input type="checkbox" data-perm-code="${escapeHTML(code)}" ${checked} />${escapeHTML(code)}`;
        options.appendChild(label);
      });

      wrap.appendChild(options);
      ui.permChecklist.appendChild(wrap);
    });
  }

  function permissionSelection() {
    const selected = [];
    ui.permChecklist.querySelectorAll("input[data-perm-code]").forEach((input) => {
      if (input.checked) {
        selected.push(input.getAttribute("data-perm-code"));
      }
    });
    return selected;
  }

  function syncPermissionsWithRole(roleCode) {
    renderPermissionChecklist(permissionPresetForRole(roleCode));
  }

  async function onSaveProject() {
    const title = ui.editTitleInput.value.trim();
    const shortDescription = ui.editShortDescriptionInput.value.trim();
    const readme = ui.editReadmeInput.value.trim();
    const repo = ui.editRepoInput.value.trim();
    const stacks = parseStacks(ui.editStacksInput.value);
    const visibility = ui.editVisibilityPublic.checked ? "PUBLIC" : "PRIVATE";

    if (!title) {
      throw new Error("Название проекта обязательно.");
    }

    await request("PATCH", `/v2/projects/${state.projectID}`, {
      title,
      description: shortDescription,
    });

    await request("PUT", `/v2/projects/${state.projectID}/stacks`, {
      stacks,
    });

    state.projectMeta = {
      ...state.projectMeta,
      readme,
      repo,
      visibility,
      local_stacks: stacks,
    };
    saveJSON(projectMetaKey(), state.projectMeta);

    if (visibilityLabel(state.project.visibility) !== (visibility === "PUBLIC" ? "Публичный" : "Приватный")) {
      setNotice("Изменение visibility в этом UI пока локальное (API update visibility не подключен).", false);
    } else {
      setNotice("Изменения проекта сохранены.", false);
    }

    await refreshData();
    setView("overview");
  }

  async function onCreatePosition(e) {
    e.preventDefault();

    const name = ui.positionNameInput.value.trim();
    const code = ui.positionCodeInput.value.trim();
    const capacity = parseInt(ui.positionCapacityInput.value, 10);

    if (!name) {
      throw new Error("Название роли обязательно.");
    }

    await request("POST", `/v2/projects/${state.projectID}/positions`, {
      name,
      code,
      capacity: Number.isNaN(capacity) || capacity <= 0 ? 1 : capacity,
    });

    ui.positionNameInput.value = "";
    ui.positionCodeInput.value = "";
    ui.positionCapacityInput.value = "1";

    setNotice("Роль добавлена.", false);
    await refreshData();
  }

  async function onMemberAction(actionBtn) {
    const row = actionBtn.closest("tr[data-user-id]");
    if (!row) return;

    const userID = row.getAttribute("data-user-id");
    const roleSelect = row.querySelector("select.member-role-select");
    const positionID = roleSelect ? roleSelect.value : "";
    const action = actionBtn.getAttribute("data-member-action");

    if (action === "permissions") {
      openPermissionsModal(userID);
      return;
    }

    if (!positionID) {
      throw new Error("Выберите роль участника.");
    }

    if (action === "approve") {
      await request("POST", `/v2/projects/${state.projectID}/members/${userID}/approve`, {
        position_id: positionID,
      });
      setNotice(`Участник ${shortID(userID)} одобрен.`, false);
      await refreshData();
      return;
    }

    if (action === "set-position") {
      await request("PATCH", `/v2/projects/${state.projectID}/members/${userID}/position`, {
        position_id: positionID,
      });
      setNotice(`Роль участника ${shortID(userID)} обновлена.`, false);
      await refreshData();
    }
  }

  async function onTaskAction(actionBtn) {
    const card = actionBtn.closest("[data-task-id]");
    if (!card) return;

    const taskID = card.getAttribute("data-task-id");
    const action = actionBtn.getAttribute("data-task-action");

    if (action === "claim") {
      await request("POST", `/v2/projects/${state.projectID}/tasks/${taskID}/claim`, {});
      setNotice(`Задача ${shortID(taskID)} взята в работу.`, false);
      await refreshData();
      return;
    }

    if (action === "status") {
      const statusSelect = card.querySelector("select[data-task-status]");
      const status = statusSelect ? statusSelect.value : "";
      if (!status) {
        throw new Error("Выберите статус.");
      }
      await request("PATCH", `/v2/projects/${state.projectID}/tasks/${taskID}/status`, {
        status,
      });
      setNotice(`Статус задачи ${shortID(taskID)} обновлен.`, false);
      await refreshData();
      return;
    }

    if (action === "assign") {
      const assigneeSelect = card.querySelector("select[data-task-assignee]");
      const assignee = assigneeSelect ? assigneeSelect.value : "";
      if (!assignee) {
        throw new Error("Выберите исполнителя.");
      }
      await request("PATCH", `/v2/projects/${state.projectID}/tasks/${taskID}/assignee`, {
        assignee_user_id: assignee,
      });
      setNotice(`Исполнитель задачи ${shortID(taskID)} обновлен.`, false);
      await refreshData();
    }
  }

  async function onOpenRecruitment() {
    await request("POST", `/v2/projects/${state.projectID}/recruitment/open`, {});
    setNotice("Набор в проект открыт.", false);
    await refreshData();
  }

  async function onApplyMember() {
    await request("POST", `/v2/projects/${state.projectID}/members/apply`, {});
    setNotice("Заявка на вступление отправлена.", false);
    await refreshData();
  }

  async function onAssignProfessor() {
    const professorID = ui.professorIDInput.value.trim();
    if (!professorID) {
      throw new Error("Введите UUID преподавателя.");
    }

    await request("POST", `/v2/projects/${state.projectID}/professor`, {
      professor_id: professorID,
    });
    ui.professorIDInput.value = "";
    setNotice("Преподаватель назначен.", false);
    await refreshData();
  }

  async function onApproveProject() {
    await request("POST", `/v2/projects/${state.projectID}/approve`, {});
    setNotice("Проект переведен в ACTIVE.", false);
    await refreshData();
  }

  function savePermissions() {
    if (!state.currentPermUserID) return;

    state.memberPerms[state.currentPermUserID] = {
      role: ui.permRoleSelect.value,
      permissions: permissionSelection(),
    };

    saveJSON(memberPermsKey(), state.memberPerms);
    closeModal(ui.permissionsModal);
    setNotice("Права участника обновлены (локально).", false);
  }

  async function refreshData() {
    state.project = await request("GET", `/v2/projects/${state.projectID}`);

    const [stacks, positions, members, readiness, tasks] = await Promise.all([
      loadOptional("stacks", "GET", `/v2/projects/${state.projectID}/stacks`, []),
      loadOptional("positions", "GET", `/v2/projects/${state.projectID}/positions`, []),
      loadOptional("members", "GET", `/v2/projects/${state.projectID}/members`, []),
      loadOptional("readiness", "GET", `/v2/projects/${state.projectID}/readiness`, null),
      loadOptional("tasks", "GET", `/v2/projects/${state.projectID}/tasks`, []),
    ]);

    state.stacks = Array.isArray(stacks) ? stacks : [];
    if (state.stacks.length === 0 && Array.isArray(state.projectMeta.local_stacks) && state.projectMeta.local_stacks.length > 0) {
      state.stacks = state.projectMeta.local_stacks.map((code) => ({ code: String(code).toUpperCase() }));
    } else if (state.stacks.length > 0) {
      state.projectMeta = {
        ...state.projectMeta,
        local_stacks: state.stacks.map((s) => s.code),
      };
      saveJSON(projectMetaKey(), state.projectMeta);
    }
    state.positions = Array.isArray(positions) ? positions : [];
    state.members = Array.isArray(members) ? members : [];
    state.readiness = readiness && typeof readiness === "object" ? readiness : null;
    state.tasks = Array.isArray(tasks) ? tasks : [];

    localStorage.setItem(LS_SELECTED_PROJECT, JSON.stringify(state.project));

    renderAll();
  }

  function wireTabSwitching() {
    ui.tabButtons.forEach((btn) => {
      btn.addEventListener("click", () => {
        const view = btn.getAttribute("data-view") || "overview";
        setView(view);
      });
    });

    ui.switchViewButtons.forEach((btn) => {
      btn.addEventListener("click", () => {
        setView(btn.getAttribute("data-switch-view") || "overview");
      });
    });
  }

  function wireModals() {
    document.querySelectorAll("[data-close-modal]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const id = btn.getAttribute("data-close-modal");
        const modal = document.getElementById(id);
        if (modal) closeModal(modal);
      });
    });

    [ui.taskModal, ui.permissionsModal].forEach((modal) => {
      modal.addEventListener("click", (e) => {
        if (e.target === modal) {
          closeModal(modal);
        }
      });
    });

    document.addEventListener("keydown", (e) => {
      if (e.key !== "Escape") return;
      if (!ui.taskModal.hidden) closeModal(ui.taskModal);
      if (!ui.permissionsModal.hidden) closeModal(ui.permissionsModal);
    });
  }

  function wireEvents() {
    ui.logoutBtn.addEventListener("click", () => {
      clearSession();
      window.location.href = "/dev/login";
    });

    ui.favoriteBtn.addEventListener("click", () => {
      state.favorite = !state.favorite;
      state.projectMeta = {
        ...state.projectMeta,
        favorite: state.favorite,
      };
      saveJSON(projectMetaKey(), state.projectMeta);
      renderHero();
    });

    ui.refreshBtn.addEventListener("click", async () => {
      try {
        await refreshData();
        setNotice("Данные обновлены.", false);
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.openEditViewBtn.addEventListener("click", () => setView("edit"));
    ui.closeEditViewBtn.addEventListener("click", () => setView("overview"));
    ui.cancelEditBtn.addEventListener("click", () => setView("overview"));

    ui.saveProjectBtn.addEventListener("click", async () => {
      try {
        await onSaveProject();
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.editStacksInput.addEventListener("input", () => {
      renderEditStackChips();
    });

    ui.positionForm.addEventListener("submit", async (e) => {
      try {
        await onCreatePosition(e);
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.teamTableBody.addEventListener("click", async (e) => {
      const btn = e.target.closest("button[data-member-action]");
      if (!btn) return;
      try {
        await onMemberAction(btn);
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.openRecruitmentBtn.addEventListener("click", async () => {
      try {
        await onOpenRecruitment();
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.applyMemberBtn.addEventListener("click", async () => {
      try {
        await onApplyMember();
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.assignProfessorBtn.addEventListener("click", async () => {
      try {
        await onAssignProfessor();
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    ui.approveProjectBtn.addEventListener("click", async () => {
      try {
        await onApproveProject();
      } catch (err) {
        if (err.data && typeof err.data === "object" && err.data.readiness) {
          setNotice(
            `Недостаточно условий: ${err.data.readiness.active_members}/${err.data.readiness.required_members} участников, критерии ${err.data.readiness.criteria_count}.`,
            true
          );
        } else {
          setNotice(err.message || String(err), true);
        }
      }
    });

    ui.openTaskModalBtn.addEventListener("click", openTaskModal);
    ui.taskModalPositionSelect.addEventListener("change", syncTaskModalAssignees);
    ui.taskModalCreateBtn.addEventListener("click", async () => {
      try {
        await createTaskFromModal();
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    [ui.todoTasks, ui.doingTasks, ui.doneTasks].forEach((container) => {
      container.addEventListener("click", async (e) => {
        const btn = e.target.closest("button[data-task-action]");
        if (!btn) return;
        try {
          await onTaskAction(btn);
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    });

    ui.globalSearchInput.addEventListener("input", () => {
      state.searchQuery = String(ui.globalSearchInput.value || "").trim().toLowerCase();
      renderTeamTable();
      renderTasks();
    });

    ui.permRoleSelect.addEventListener("change", () => {
      syncPermissionsWithRole(ui.permRoleSelect.value);
    });

    ui.savePermissionsBtn.addEventListener("click", savePermissions);
  }

  async function bootstrap() {
    const claims = ensureSession();
    if (!claims) return;

    bindProfile();

    state.projectID = projectIDFromPath();
    if (!state.projectID) {
      window.location.href = "/dev/projects";
      return;
    }

    state.projectMeta = loadJSON(projectMetaKey(), {});
    state.taskMeta = loadJSON(taskMetaKey(), {});
    state.memberPerms = loadJSON(memberPermsKey(), {});
    state.favorite = Boolean(state.projectMeta.favorite);

    wireTabSwitching();
    wireModals();
    wireEvents();

    try {
      await refreshData();
      setView("overview");
    } catch (err) {
      setNotice(`Не удалось загрузить проект: ${err.message || String(err)}`, true);
    }
  }

  bootstrap();
})();
