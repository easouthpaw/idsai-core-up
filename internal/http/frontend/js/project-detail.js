(() => {
  const auth = window.IDSAIAuth;
  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const LS_USER = "idsai_rbac_user_id";
  const LS_FACULTY = "idsai_rbac_faculty_id";
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";
  const LS_SELECTED_PROJECT = "idsai_selected_project";
  const LS_STUDENT_SECTION = "idsai_student_section";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

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
  const PRIMARY_LIFECYCLE_FLOW = ["DRAFT", "RECRUITMENT", "ACTIVE", "GRADING", "ARCHIVE"];
  const DEFAULT_PROJECT_COVERS = [
    "https://images.pexels.com/photos/16129724/pexels-photo-16129724.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/17323801/pexels-photo-17323801.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/4508751/pexels-photo-4508751.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/5257576/pexels-photo-5257576.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/10499056/pexels-photo-10499056.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/12899157/pexels-photo-12899157.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
  ];

  const ui = {
    profileAvatar: document.getElementById("profileAvatar"),
    studentName: document.getElementById("studentName"),
    studentEmail: document.getElementById("studentEmail"),
    sidebarNavLinks: Array.from(document.querySelectorAll(".side-nav [data-nav-section]")),

    crumbSectionLink: document.getElementById("crumbSectionLink"),
    crumbProject: document.getElementById("crumbProject"),
    title: document.getElementById("projectTitle"),
    statusBadge: document.getElementById("statusBadge"),
    visibilityLabel: document.getElementById("visibilityLabel"),
    projectID: document.getElementById("projectID"),
    repoLink: document.getElementById("repoLink"),
    heroCoverImage: document.getElementById("heroCoverImage"),

    pageNotice: document.getElementById("pageNotice"),
    globalSearchInput: document.getElementById("globalSearchInput"),
    lifecycleSummary: document.getElementById("lifecycleSummary"),
    lifecycleCurrentStage: document.getElementById("lifecycleCurrentStage"),
    lifecycleTimeline: document.getElementById("lifecycleTimeline"),

    tabButtons: Array.from(document.querySelectorAll(".tab-btn")),
    switchViewButtons: Array.from(document.querySelectorAll("[data-switch-view]")),

    viewOverview: document.getElementById("view-overview"),
    viewTeam: document.getElementById("view-team"),
    viewInvite: document.getElementById("view-invite"),
    viewTasks: document.getElementById("view-tasks"),
    viewCriteria: document.getElementById("view-criteria"),
    viewReview: document.getElementById("view-review"),
    viewEdit: document.getElementById("view-edit"),

    aboutCard: document.getElementById("aboutCard"),
    applyCard: document.getElementById("applyCard"),
    applyHint: document.getElementById("applyHint"),
    applyCommentInput: document.getElementById("applyCommentInput"),
    applyProjectBtn: document.getElementById("applyProjectBtn"),
    stackCard: document.getElementById("stackCard"),
    activityCard: document.getElementById("activityCard"),
    teamMiniCard: document.getElementById("teamMiniCard"),
    pipelineCard: document.getElementById("pipelineCard"),

    aboutContent: document.getElementById("aboutContent"),
    stackChips: document.getElementById("stackChips"),
    teamMiniList: document.getElementById("teamMiniList"),
    readinessList: document.getElementById("readinessList"),
    activityList: document.getElementById("activityList"),

    openRecruitmentBtn: document.getElementById("openRecruitmentBtn"),
    professorSearchInput: document.getElementById("professorSearchInput"),
    professorSearchResults: document.getElementById("professorSearchResults"),
    professorIdentity: document.getElementById("professorIdentity"),
    professorInviteHint: document.getElementById("professorInviteHint"),
    assignProfessorBtn: document.getElementById("assignProfessorBtn"),
    approveProjectBtn: document.getElementById("approveProjectBtn"),
    completeProjectBtn: document.getElementById("completeProjectBtn"),

    positionForm: document.getElementById("positionForm"),
    positionNameInput: document.getElementById("positionNameInput"),
    positionCodeInput: document.getElementById("positionCodeInput"),
    positionCapacityInput: document.getElementById("positionCapacityInput"),
    teamTableBody: document.getElementById("teamTableBody"),
    inviteSearchInput: document.getElementById("inviteSearchInput"),
    inviteRefreshBtn: document.getElementById("inviteRefreshBtn"),
    inviteCandidatesList: document.getElementById("inviteCandidatesList"),

    progressBadge: document.getElementById("progressBadge"),
    openTaskModalBtn: document.getElementById("openTaskModalBtn"),
    todoTasks: document.getElementById("todoTasks"),
    doingTasks: document.getElementById("doingTasks"),
    doneTasks: document.getElementById("doneTasks"),
    countTodo: document.getElementById("countTodo"),
    countDoing: document.getElementById("countDoing"),
    countDone: document.getElementById("countDone"),
    tasksTeamList: document.getElementById("tasksTeamList"),
    criteriaListView: document.getElementById("criteriaListView"),
    criteriaCountMeta: document.getElementById("criteriaCountMeta"),
    criteriaReviewHint: document.getElementById("criteriaReviewHint"),
    reviewStatusPill: document.getElementById("reviewStatusPill"),
    reviewIntro: document.getElementById("reviewIntro"),
    reviewCriteriaList: document.getElementById("reviewCriteriaList"),
    reviewGauge: document.getElementById("reviewGauge"),
    reviewGaugeValue: document.getElementById("reviewGaugeValue"),
    reviewSummaryScore: document.getElementById("reviewSummaryScore"),
    reviewSummaryMet: document.getElementById("reviewSummaryMet"),
    reviewSummaryMissed: document.getElementById("reviewSummaryMissed"),
    reviewSummaryDate: document.getElementById("reviewSummaryDate"),
    reviewSummaryReviewer: document.getElementById("reviewSummaryReviewer"),
    reviewOverallComment: document.getElementById("reviewOverallComment"),

    activeProgressWrap: document.getElementById("activeProgressWrap"),
    progressPercent: document.getElementById("progressPercent"),
    progressFill: document.getElementById("progressFill"),
    stackInfoConsole: document.getElementById("stackInfoConsole"),

    favoriteBtn: document.getElementById("favoriteBtn"),
    openEditViewBtn: document.getElementById("openEditViewBtn"),
    closeEditViewBtn: document.getElementById("closeEditViewBtn"),
    cancelEditBtn: document.getElementById("cancelEditBtn"),
    deleteProjectBtn: document.getElementById("deleteProjectBtn"),
    saveProjectBtn: document.getElementById("saveProjectBtn"),

    editTitleInput: document.getElementById("editTitleInput"),
    editShortDescriptionInput: document.getElementById("editShortDescriptionInput"),
    editReadmeInput: document.getElementById("editReadmeInput"),
    editRepoInput: document.getElementById("editRepoInput"),
    editStacksInput: document.getElementById("editStacksInput"),
    editVisibilityPublic: document.getElementById("editVisibilityPublic"),
    editVisibilityPrivate: document.getElementById("editVisibilityPrivate"),
    editStackChips: document.getElementById("editStackChips"),
    editCoverPreview: document.getElementById("editCoverPreview"),
    editCoverInput: document.getElementById("editCoverInput"),
    uploadCoverBtn: document.getElementById("uploadCoverBtn"),
    removeCoverBtn: document.getElementById("removeCoverBtn"),
    coverStatus: document.getElementById("coverStatus"),

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
    taskResultModal: document.getElementById("taskResultModal"),
    taskResultTarget: document.getElementById("taskResultTarget"),
    taskResultCommentInput: document.getElementById("taskResultCommentInput"),
    taskResultAttachmentsInput: document.getElementById("taskResultAttachmentsInput"),
    taskResultSubmitBtn: document.getElementById("taskResultSubmitBtn"),

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
    criteria: [],
    gradingItems: [],
    readiness: null,
    tasks: [],
    taskActivities: [],
    activeView: "overview",
    searchQuery: "",
    projectMeta: {},
    taskMeta: {},
    memberPerms: {},
    currentPermUserID: "",
    noticeTimer: null,
    favorite: false,
    studentCandidates: [],
    professorCandidates: [],
    selectedProfessorID: "",
    professorSummary: null,
    professorSearchTimer: null,
    studentSearchTimer: null,
    userDirectory: {},
    currentResultTaskID: "",
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

  function stableHash(value) {
    const s = String(value || "");
    let h = 0;
    for (let i = 0; i < s.length; i += 1) {
      h = (h * 31 + s.charCodeAt(i)) % 100000;
    }
    return h;
  }

  function defaultCoverIndex(project) {
    const variant = Number.parseInt(String(project && project.default_cover_variant !== undefined ? project.default_cover_variant : ""), 10);
    if (Number.isFinite(variant) && variant > 0) {
      return (variant - 1) % DEFAULT_PROJECT_COVERS.length;
    }
    return stableHash(project && (project.id || project.title || "")) % DEFAULT_PROJECT_COVERS.length;
  }

  function defaultCoverURL(project) {
    return DEFAULT_PROJECT_COVERS[defaultCoverIndex(project)] || DEFAULT_PROJECT_COVERS[0];
  }

  function projectCoverURL(project) {
    const custom = String(project && project.image_url ? project.image_url : "").trim();
    if (custom) return custom;
    return defaultCoverURL(project);
  }

  function bindCoverFallback(img, fallback) {
    if (!img) return;
    img.onerror = () => {
      if (fallback && img.src !== fallback) {
        img.src = fallback;
      }
    };
  }

  function renderProjectCovers() {
    const project = state.project;
    if (!project) return;

    const fallback = defaultCoverURL(project);
    const cover = projectCoverURL(project);

    if (ui.heroCoverImage) {
      ui.heroCoverImage.src = cover;
      bindCoverFallback(ui.heroCoverImage, fallback);
    }
    if (ui.editCoverPreview) {
      ui.editCoverPreview.src = cover;
      bindCoverFallback(ui.editCoverPreview, fallback);
    }
  }

  function setCoverStatus(message, isError) {
    if (!ui.coverStatus) return;
    ui.coverStatus.textContent = message || "";
    ui.coverStatus.classList.toggle("err", Boolean(isError));
    ui.coverStatus.classList.toggle("ok", Boolean(message) && !Boolean(isError));
  }

  function setButtonLoading(btn, loading, loadingText) {
    if (!btn) return;
    if (!btn.dataset.baseText) {
      btn.dataset.baseText = btn.textContent || "";
    }
    btn.disabled = Boolean(loading);
    btn.textContent = loading ? loadingText : btn.dataset.baseText;
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

  async function requestForm(method, url, formData) {
    const { resp, data } = await auth.requestJSON(url, {
      method,
      body: formData,
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

  async function loadOptional(name, method, url, fallback) {
    try {
      return await request(method, url);
    } catch (err) {
      console.warn(`[project-ui] ${name} failed`, err);
      return fallback;
    }
  }

  function clearSession() {
    auth.clearClientState();
    localStorage.removeItem(LS_SELECTED_PROJECT);
  }

  function currentStudentSection() {
    const params = new URLSearchParams(window.location.search || "");
    const fromURL = String(params.get("nav") || "").trim().toLowerCase();
    if (fromURL === "community" || fromURL === "invites" || fromURL === "mine") {
      return fromURL;
    }
    const cachedProject = loadJSON(LS_SELECTED_PROJECT, null);
    const fromProject = String(cachedProject?._nav_section || "").trim().toLowerCase();
    if (fromProject === "community" || fromProject === "invites" || fromProject === "mine") {
      return fromProject;
    }
    const fromStorage = String(localStorage.getItem(LS_STUDENT_SECTION) || "").trim().toLowerCase();
    if (fromStorage === "community" || fromStorage === "invites" || fromStorage === "mine") {
      return fromStorage;
    }
    return "mine";
  }

  function syncStudentSidebar() {
    const section = currentStudentSection();
    localStorage.setItem(LS_STUDENT_SECTION, section);

    ui.sidebarNavLinks.forEach((link) => {
      const value = String(link.getAttribute("data-nav-section") || "").trim().toLowerCase();
      link.classList.toggle("active", value === section);
    });

    if (!ui.crumbSectionLink) return;
    if (section === "community") {
      ui.crumbSectionLink.textContent = "Сообщество";
      ui.crumbSectionLink.href = "/dev/projects?tab=community";
      return;
    }
    if (section === "invites") {
      ui.crumbSectionLink.textContent = "Заявки";
      ui.crumbSectionLink.href = "/dev/invites";
      return;
    }
    ui.crumbSectionLink.textContent = "Мои проекты";
    ui.crumbSectionLink.href = "/dev/projects?tab=mine";
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

  function viewerAccess() {
    if (state.project && typeof state.project.viewer_access === "object" && state.project.viewer_access) {
      return state.project.viewer_access;
    }
    return {
      can_view_workspace: false,
      can_apply: false,
      can_view_final_grade: false,
    };
  }

  function canViewWorkspace() {
    return Boolean(viewerAccess().can_view_workspace);
  }

  function canApplyToProject() {
    return Boolean(viewerAccess().can_apply);
  }

  function canViewFinalGrade() {
    return Boolean(viewerAccess().can_view_final_grade);
  }

  function allowedViews() {
    if (canViewWorkspace()) {
      return ["overview", "team", "invite", "tasks", "criteria", "review", "edit"];
    }
    if (canViewFinalGrade()) {
      return ["overview", "review"];
    }
    return ["overview"];
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

  function lifecycleStageLabel(status) {
    const s = String(status || "").toUpperCase();
    if (s === "REVIEW") return "Модерация";
    if (s === "RECRUITMENT") return "Набор";
    if (s === "ACTIVE") return "В работе";
    if (s === "GRADING") return "Оценивание";
    if (s === "ARCHIVE") return "Архив";
    return "Черновик";
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

  function rememberUser(userID, fullName, email) {
    const id = String(userID || "").trim();
    if (!id) return;
    state.userDirectory[id] = {
      fullName: String(fullName || "").trim(),
      email: String(email || "").trim(),
    };
  }

  function getDisplayName(userID) {
    const id = String(userID || "").trim();
    const currentUser = localStorage.getItem(LS_USER) || "";
    if (id === String(currentUser)) {
      return localStorage.getItem(LS_STUDENT_NAME) || "Текущий студент";
    }
    const known = state.userDirectory[id];
    if (known && known.fullName) return known.fullName;
    if (known && known.email) return known.email;
    return `Student ${shortID(id)}`;
  }

  function getDisplaySubline(userID) {
    const id = String(userID || "").trim();
    const known = state.userDirectory[id];
    if (known && known.email) return known.email;
    return shortID(id);
  }

  function allMembers() {
    if (!canViewWorkspace()) {
      return [];
    }
    const members = Array.isArray(state.members) ? [...state.members] : [];
    const leadID = String(state.project?.created_by || "").trim();
    if (leadID && !members.some((m) => String(m.user_id) === leadID)) {
      members.unshift({
        id: `lead-${leadID}`,
        project_id: state.projectID,
        user_id: leadID,
        status: "ACTIVE",
        position_code: "TEAM_LEAD",
        position_name: "Тимлид",
      });
    }
    return members;
  }

  function getRoleLabel(member) {
    if (member.position_name) return member.position_name;
    if (member.position_code) return member.position_code;
    return "Без роли";
  }

  function activeMembers() {
    return allMembers().filter((m) => String(m.status || "").toUpperCase() === "ACTIVE");
  }

  function membersByPosition(positionID) {
    return activeMembers().filter((m) => String(m.position_id || "") === String(positionID));
  }

  function isCurrentUserLead() {
    const current = String(localStorage.getItem(LS_USER) || "");
    const creator = String(state.project?.created_by || "");
    if (current && creator && current === creator) {
      return true;
    }
    return allMembers().some((m) => String(m.user_id) === current && String(m.position_code || "").toUpperCase() === "TEAM_LEAD");
  }

  function isCurrentUserCreator() {
    const current = String(localStorage.getItem(LS_USER) || "");
    return current && current === String(state.project?.created_by || "");
  }

  function isCurrentUserActiveMember() {
    const current = String(localStorage.getItem(LS_USER) || "");
    if (!current) return false;
    if (current === String(state.project?.created_by || "")) return true;
    return allMembers().some((m) => String(m.user_id) === current && String(m.status || "").toUpperCase() === "ACTIVE");
  }

  function currentMemberStatus() {
    const current = String(localStorage.getItem(LS_USER) || "");
    if (!current) return "";
    if (current === String(state.project?.created_by || "")) return "ACTIVE";

    const raw = Array.isArray(state.members) ? state.members : [];
    const item = raw.find((m) => String(m.user_id) === current);
    if (!item) return "";
    return String(item.status || "").toUpperCase();
  }

  function professorReviewLabel(statusCode, hasProfessor) {
    const code = String(statusCode || "NONE").toUpperCase();
    if (code === "PENDING") return "Ожидаем ответа";
    if (code === "ACCEPTED") return "Подтвержден";
    if (code === "REJECTED") return "Отклонено";
    return hasProfessor ? "Назначен" : "Не назначен";
  }

  function doneTasksCount() {
    return Array.isArray(state.tasks)
      ? state.tasks.filter((task) => String(task.status || "").toUpperCase() === "DONE").length
      : 0;
  }

  function gradedCriteriaCount() {
    return Array.isArray(state.gradingItems)
      ? state.gradingItems.filter((item) => item && item.is_met !== null && item.is_met !== undefined).length
      : 0;
  }

  function lifecycleSnapshot() {
    const readiness = state.readiness || {};
    const publicSummary = state.project && typeof state.project.review_summary === "object"
      ? state.project.review_summary
      : null;
    const professorStatusCode = String(readiness.professor_status || state.project?.professor_review_status || "NONE").toUpperCase();
    const hasProfessor = Boolean(readiness.has_professor || state.project?.professor_id);
    const criteriaCount = Array.isArray(state.criteria) && state.criteria.length > 0
      ? state.criteria.length
      : Number(readiness.criteria_count || publicSummary?.total || 0);
    const tasksTotal = Array.isArray(state.tasks) ? state.tasks.length : 0;
    const tasksDone = doneTasksCount();

    return {
      statusCode: projectStatusCode() || "DRAFT",
      requiredMembers: Number(readiness.required_members || 0),
      activeMembers: Number(readiness.active_members || 0),
      criteriaCount,
      canActivate: Boolean(readiness && readiness.can_activate),
      hasProfessor,
      professorStatusCode,
      professorLabel: professorReviewLabel(professorStatusCode, hasProfessor),
      professorAccepted: professorStatusCode === "ACCEPTED",
      tasksTotal,
      tasksDone,
      gradedCriteria: Math.max(gradedCriteriaCount(), Number(publicSummary?.total || 0)),
    };
  }

  function lifecycleSummaryText(snapshot) {
    if (snapshot.statusCode === "REVIEW") {
      return snapshot.canActivate
        ? "Проект прошел подготовку и может быть запущен после модерации."
        : "Опциональный этап модерации: после него можно открыть набор или сразу запускать проект при полной готовности.";
    }

    if (snapshot.statusCode === "RECRUITMENT") {
      const blockers = [];
      if (snapshot.requiredMembers === 0) {
        blockers.push("создать роли");
      } else if (snapshot.activeMembers < snapshot.requiredMembers) {
        blockers.push(`добрать команду ${snapshot.activeMembers}/${snapshot.requiredMembers}`);
      }
      if (!snapshot.hasProfessor) {
        blockers.push("назначить преподавателя");
      } else if (!snapshot.professorAccepted) {
        blockers.push("получить подтверждение преподавателя");
      }
      if (snapshot.criteriaCount === 0) {
        blockers.push("добавить критерии");
      }
      return blockers.length > 0
        ? `До запуска осталось: ${blockers.join(", ")}.`
        : "Команда и критерии готовы: проект можно переводить в ACTIVE.";
    }

    if (snapshot.statusCode === "ACTIVE") {
      if (snapshot.tasksTotal === 0) {
        return "Проект запущен. Следующий шаг: создать и выполнить задачи перед отправкой на оценивание.";
      }
      if (!snapshot.professorAccepted) {
        return `Проект в работе. Завершено ${snapshot.tasksDone}/${snapshot.tasksTotal} задач, но преподаватель еще не подтвердил участие.`;
      }
      if (snapshot.tasksDone < snapshot.tasksTotal) {
        return `Проект в работе. Закройте все задачи: сейчас выполнено ${snapshot.tasksDone}/${snapshot.tasksTotal}.`;
      }
      return "Все задачи закрыты: проект можно отправлять преподавателю на оценивание.";
    }

    if (snapshot.statusCode === "GRADING") {
      if (snapshot.criteriaCount === 0) {
        return "Проект находится на оценивании. Для публикации итогов нужно добавить критерии.";
      }
      if (snapshot.gradedCriteria < snapshot.criteriaCount) {
        return `Проект на проверке: оценки выставлены по ${snapshot.gradedCriteria}/${snapshot.criteriaCount} критериям.`;
      }
      return "Все критерии оценены: можно публиковать итоговую оценку и переносить проект в архив.";
    }

    if (snapshot.statusCode === "ARCHIVE") {
      return "Проект завершен: итог опубликован, карточка остается в истории платформы.";
    }

    return "Стартовая стадия: заполните описание, роли и стек, затем откройте набор или отправьте проект на модерацию.";
  }

  function lifecycleStepState(stepCode, currentStatus) {
    const code = String(stepCode || "").toUpperCase();
    const status = String(currentStatus || "").toUpperCase();

    if (code === "REVIEW") {
      return status === "REVIEW" ? "is-current" : "is-optional";
    }

    const stepIndex = PRIMARY_LIFECYCLE_FLOW.indexOf(code);
    const currentIndex = PRIMARY_LIFECYCLE_FLOW.indexOf(status);

    if (stepIndex === -1) return "is-upcoming";
    if (status === "REVIEW") {
      return code === "DRAFT" ? "is-complete" : "is-upcoming";
    }
    if (currentIndex === -1) return stepIndex === 0 ? "is-current" : "is-upcoming";
    if (stepIndex < currentIndex) return "is-complete";
    if (stepIndex === currentIndex) return "is-current";
    return "is-upcoming";
  }

  function lifecycleStateLabel(stepState, optional) {
    if (stepState === "is-current") return "Сейчас";
    if (stepState === "is-complete") return "Пройдено";
    if (optional) return "Опционально";
    return "Впереди";
  }

  function isRecruitmentApplyMode() {
    const status = projectStatusCode();
    if (status !== "RECRUITMENT") return false;
    return canApplyToProject();
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

  function isCurrentUserAssignee(task) {
    const current = String(localStorage.getItem(LS_USER) || "");
    const assignee = String(task.assignee_user_id || "");
    return Boolean(current) && current === assignee;
  }

  function taskActivities(taskID) {
    return (Array.isArray(state.taskActivities) ? state.taskActivities : []).filter((item) => String(item.task_id) === String(taskID));
  }

  function gradingByCriterion() {
    const map = new Map();
    (Array.isArray(state.gradingItems) ? state.gradingItems : []).forEach((item) => {
      const id = String(item.criterion_id || "").trim();
      if (!id) return;
      map.set(id, {
        isMet: item.is_met === true ? true : item.is_met === false ? false : null,
        comment: String(item.comment || "").trim(),
        updatedAt: item.updated_at || null,
      });
    });
    return map;
  }

  function reviewSummaryData() {
    const criteria = Array.isArray(state.criteria) ? state.criteria : [];
    const grading = gradingByCriterion();
    const publicSummary = state.project && typeof state.project.review_summary === "object"
      ? state.project.review_summary
      : null;

    if ((!criteria.length || grading.size === 0) && publicSummary) {
      const reviewedAt = publicSummary.reviewed_at ? new Date(publicSummary.reviewed_at) : null;
      return {
        total: Number(publicSummary.total || 0),
        met: Number(publicSummary.met || 0),
        missed: Math.max(0, Number(publicSummary.total || 0) - Number(publicSummary.met || 0)),
        reviewed: Number(publicSummary.total || 0),
        passPercent: Number(publicSummary.pass_percent || 0),
        score: String(publicSummary.score || "0.0"),
        reviewer: String(publicSummary.reviewer || "Преподаватель"),
        reviewedAt: reviewedAt && !Number.isNaN(reviewedAt.getTime()) ? reviewedAt : null,
        overall: "Итоговая оценка опубликована. Детализация критериев доступна только участникам команды.",
        hasReview: Number(publicSummary.total || 0) > 0,
      };
    }

    let met = 0;
    let missed = 0;
    let reviewed = 0;
    let latest = null;
    const comments = [];

    criteria.forEach((criterion) => {
      const id = String(criterion.id || "");
      const item = grading.get(id);
      if (!item) return;
      if (item.isMet === true) {
        met += 1;
        reviewed += 1;
      } else if (item.isMet === false) {
        missed += 1;
        reviewed += 1;
      }
      if (item.comment) {
        comments.push({
          title: String(criterion.title || "Критерий").trim(),
          text: item.comment,
          isMet: item.isMet,
        });
      }
      if (item.updatedAt) {
        const dt = new Date(item.updatedAt);
        if (!Number.isNaN(dt.getTime()) && (!latest || dt > latest)) latest = dt;
      }
    });

    const total = criteria.length;
    const passPercent = total > 0 ? Math.round((met * 100) / total) : 0;
    const score = total > 0 ? ((met * 5) / total).toFixed(1) : "0.0";
    const reviewer = state.professorSummary?.full_name || state.professorSummary?.email || "Преподаватель";

    let overall = "Комментарий преподавателя пока не добавлен.";
    if (comments.length > 0) {
      const ranked = comments
        .slice()
        .sort((a, b) => {
          const aw = a.isMet === false ? 0 : 1;
          const bw = b.isMet === false ? 0 : 1;
          return aw - bw;
        })
        .slice(0, 3);
      overall = ranked
        .map((item) => `${item.title}: ${item.text}`)
        .join("\n\n");
    } else if (reviewed > 0 && missed === 0) {
      overall = "Все проверенные критерии отмечены как выполненные.";
    } else if (reviewed > 0) {
      overall = `Есть замечания по ${missed} критериям.`;
    }

    return {
      total,
      met,
      missed,
      reviewed,
      passPercent,
      score,
      reviewer,
      reviewedAt: latest,
      overall,
      hasReview: reviewed > 0 || comments.length > 0,
    };
  }

  function taskActivityLabel(item) {
    const event = String(item.event_type || "").toUpperCase();
    if (event === "CREATED") return "Создана задача";
    if (event === "ASSIGNED") return "Назначен исполнитель";
    if (event === "CLAIMED") return "Задача взята в работу";
    if (event === "COMPLETED") return "Задача завершена";
    const from = String(item.from_status || "").toUpperCase();
    const to = String(item.to_status || "").toUpperCase();
    if (event === "STATUS_CHANGED" && from && to) {
      return `Статус: ${from} -> ${to}`;
    }
    return "Изменение задачи";
  }

  function taskActivityTone(item) {
    const event = String(item.event_type || "").toUpperCase();
    if (event === "CREATED") return "created";
    if (event === "ASSIGNED") return "assigned";
    if (event === "CLAIMED") return "claimed";
    if (event === "COMPLETED") return "completed";
    return "changed";
  }

  function renderTaskActivityTimeline(taskID) {
    const items = taskActivities(taskID);
    if (!items.length) {
      return '<div class="task-timeline-empty">Лог активности появится после первых действий по задаче.</div>';
    }

    return (
      '<div class="task-timeline">' +
      items
        .map((item) => {
          const tone = taskActivityTone(item);
          const actor = item.actor_name || item.actor_email || "Система";
          const comment = String(item.comment || "").trim();
          const attachments = Array.isArray(item.attachments) ? item.attachments.filter(Boolean) : [];
          const attachmentHTML = attachments.length
            ? `<div class="timeline-attachments">` +
                attachments
                  .map((value) => {
                    const raw = String(value || "").trim();
                    const safe = (() => {
                      try {
                        const parsed = new URL(raw);
                        if (parsed.protocol === "http:" || parsed.protocol === "https:") {
                          return parsed.toString();
                        }
                      } catch (_) {}
                      return "";
                    })();
                    if (!safe) return "";
                    const href = escapeHTML(safe);
                    return `<a href="${href}" target="_blank" rel="noreferrer">${href}</a>`;
                  })
                  .join("") +
              `</div>`
            : "";
          return (
            `<article class="timeline-item ${tone}">` +
              `<div class="timeline-dot"></div>` +
              `<div class="timeline-body">` +
                `<p class="timeline-title">${escapeHTML(taskActivityLabel(item))}</p>` +
                `<p class="timeline-meta">${escapeHTML(actor)} · ${escapeHTML(formatDate(item.created_at))}</p>` +
                (comment ? `<p class="timeline-comment">${escapeHTML(comment)}</p>` : "") +
                attachmentHTML +
              `</div>` +
            `</article>`
          );
        })
        .join("") +
      `</div>`
    );
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
    const avatarURL = localStorage.getItem(LS_AVATAR_URL) || "";

    ui.studentName.textContent = name;
    ui.studentEmail.textContent = email;
    renderAvatar(ui.profileAvatar, initials(name, email), avatarURL);
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

    const applyMode = canApplyToProject();
    const canEdit = isCurrentUserActiveMember() && !applyMode;

    ui.favoriteBtn.textContent = state.favorite ? "★ В избранном" : "★ В избранное";
    ui.favoriteBtn.hidden = applyMode;
    ui.openEditViewBtn.hidden = !canEdit;
    ui.openEditViewBtn.disabled = !canEdit;
    if (ui.uploadCoverBtn) ui.uploadCoverBtn.disabled = !canEdit;
    if (ui.removeCoverBtn) ui.removeCoverBtn.disabled = !canEdit;
    if (ui.deleteProjectBtn) {
      const canDelete = isCurrentUserCreator();
      ui.deleteProjectBtn.hidden = !canDelete;
      ui.deleteProjectBtn.disabled = !canDelete;
    }

    renderProjectCovers();
  }

  function renderAbout() {
    const description = String(state.project?.description || "").trim();
    const readme = String(getReadmeText() || description).trim();
    const content = readme || "Описание проекта пока не заполнено.";

    const html =
      `<p><strong>${escapeHTML(state.project?.title || "Project")}</strong></p>` +
      `<p>${escapeHTML(content).replaceAll("\n", "<br>")}</p>` +
      `<p class="muted-text">Последнее обновление: ${escapeHTML(formatDate(state.project?.updated_at))}</p>`;

    ui.aboutContent.innerHTML = html;
  }

  function renderLifecycle() {
    if (!ui.lifecycleTimeline || !ui.lifecycleSummary || !ui.lifecycleCurrentStage || !state.project) return;

    const snapshot = lifecycleSnapshot();
    const currentStatus = snapshot.statusCode || "DRAFT";
    const headerStatus = statusPresentation(currentStatus);
    const steps = [
      {
        code: "DRAFT",
        title: "Черновик",
        copy: "Идея проекта, описание, README, стек и базовые роли команды.",
        meta: [visibilityLabel(state.project?.visibility), `Стек ${state.stacks.length}`, "README / описание"],
      },
      {
        code: "REVIEW",
        title: "Модерация",
        copy: "Опциональная проверка от куратора или администратора перед запуском.",
        meta: ["Admin override", snapshot.canActivate ? "Можно запускать" : "Промежуточный этап"],
        optional: true,
      },
      {
        code: "RECRUITMENT",
        title: "Набор команды",
        copy: "Формируются роли, команда, преподаватель и критерии готовности проекта.",
        meta: [`Роли ${snapshot.activeMembers}/${snapshot.requiredMembers}`, `Преподаватель ${snapshot.professorLabel}`, `Критерии ${snapshot.criteriaCount}`],
      },
      {
        code: "ACTIVE",
        title: "Работа",
        copy: "Команда выполняет задачи, двигает канбан и готовит проект к сдаче.",
        meta: [`Задачи ${snapshot.tasksDone}/${snapshot.tasksTotal}`, snapshot.professorAccepted ? "Ревьюер подтвержден" : "Ждем ревьюера"],
      },
      {
        code: "GRADING",
        title: "Оценивание",
        copy: "Преподаватель проверяет проект по критериям и выставляет итог.",
        meta: [`Оценено ${snapshot.gradedCriteria}/${snapshot.criteriaCount}`, `Профессор ${snapshot.professorLabel}`],
      },
      {
        code: "ARCHIVE",
        title: "Архив",
        copy: "Финальная стадия: оценка опубликована, проект завершен и сохранен в истории.",
        meta: [snapshot.statusCode === "ARCHIVE" ? "Проект завершен" : "Финальная стадия", `Критерии ${snapshot.criteriaCount}`],
      },
    ];

    ui.lifecycleSummary.textContent = lifecycleSummaryText(snapshot);
    ui.lifecycleCurrentStage.className = `status-pill ${headerStatus.cls}`;
    ui.lifecycleCurrentStage.textContent = lifecycleStageLabel(currentStatus);

    ui.lifecycleTimeline.innerHTML = steps.map((step, idx) => {
      const stepState = lifecycleStepState(step.code, currentStatus);
      const classes = ["lifecycle-step", stepState];
      if (step.optional && stepState !== "is-current") {
        classes.push("is-optional");
      }
      const meta = step.meta
        .filter(Boolean)
        .map((item) => `<span class="lifecycle-chip">${escapeHTML(item)}</span>`)
        .join("");

      return (
        `<article class="${classes.join(" ")}">` +
          `<div class="lifecycle-step-top">` +
            `<div class="lifecycle-step-index">${String(idx + 1).padStart(2, "0")}</div>` +
            `<span class="lifecycle-step-state">${escapeHTML(lifecycleStateLabel(stepState, Boolean(step.optional)))}</span>` +
          `</div>` +
          `<div class="lifecycle-step-main">` +
            `<p class="lifecycle-step-code">${escapeHTML(step.code)}</p>` +
            `<h3 class="lifecycle-step-title">${escapeHTML(step.title)}</h3>` +
            `<p class="lifecycle-step-copy">${escapeHTML(step.copy)}</p>` +
          `</div>` +
          `<div class="lifecycle-step-meta">${meta || '<span class="lifecycle-chip">Без данных</span>'}</div>` +
        `</article>`
      );
    }).join("");
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
    if (!ui.readinessList || !ui.approveProjectBtn) return;
    ui.readinessList.innerHTML = "";

    const statusCode = projectStatusCode();
    const canOpenRecruitment = isCurrentUserLead() && (statusCode === "DRAFT" || statusCode === "REVIEW");
    if (ui.openRecruitmentBtn) {
      ui.openRecruitmentBtn.hidden = !canOpenRecruitment;
      ui.openRecruitmentBtn.disabled = !canOpenRecruitment;
    }

    if (!state.readiness) {
      ui.readinessList.innerHTML = '<div class="empty-state">Данные о готовности не загружены.</div>';
      ui.approveProjectBtn.hidden = false;
      ui.approveProjectBtn.disabled = true;
      if (ui.completeProjectBtn) {
        ui.completeProjectBtn.hidden = true;
        ui.completeProjectBtn.disabled = true;
        ui.completeProjectBtn.title = "";
      }
      renderProfessorInviteArea(String(state.project?.professor_review_status || "NONE"));
      return;
    }

    const professorStatusCode = String(state.readiness.professor_status || state.project?.professor_review_status || "NONE").toUpperCase();
    const professorStatusLabel = professorReviewLabel(
      professorStatusCode,
      Boolean(state.readiness.has_professor || state.project?.professor_id)
    );

    const hasEnoughMembers = Number(state.readiness.active_members || 0) >= Number(state.readiness.required_members || 0) &&
      Number(state.readiness.required_members || 0) > 0;
    const professorAccepted = professorStatusCode === "ACCEPTED";
    const hasCriteria = Number(state.readiness.criteria_count || 0) > 0;

    const items = [
      {
        label: "Роли",
        value: `${state.readiness.active_members}/${state.readiness.required_members}`,
        stateClass: hasEnoughMembers ? "is-done" : Number(state.readiness.active_members || 0) > 0 ? "is-current" : "is-blocked",
      },
      {
        label: "Преподаватель",
        value: professorStatusLabel,
        stateClass: professorAccepted ? "is-done" : professorStatusCode === "PENDING" ? "is-current" : "is-blocked",
      },
      {
        label: "Критерии",
        value: String(state.readiness.criteria_count),
        stateClass: hasCriteria ? "is-done" : "is-blocked",
      },
      {
        label: "Запуск",
        value: state.readiness.can_activate ? "Готово" : "Не готово",
        stateClass: state.readiness.can_activate ? "is-done" : hasEnoughMembers && professorAccepted && hasCriteria ? "is-current" : "is-blocked",
      },
    ];

    items.forEach((item) => {
      const row = document.createElement("div");
      row.className = `readiness-item ${item.stateClass}`;
      row.innerHTML = `<span>${escapeHTML(item.label)}</span><strong>${escapeHTML(item.value)}</strong>`;
      ui.readinessList.appendChild(row);
    });

    const canShowApprove = statusCode !== "ACTIVE" && statusCode !== "GRADING" && statusCode !== "ARCHIVE";
    ui.approveProjectBtn.hidden = !canShowApprove;
    ui.approveProjectBtn.disabled = !state.readiness.can_activate;

    if (ui.completeProjectBtn) {
      const isMember = isCurrentUserActiveMember();
      const tasksTotal = Array.isArray(state.tasks) ? state.tasks.length : 0;
      const tasksDone = Array.isArray(state.tasks)
        ? state.tasks.filter((t) => String(t.status || "").toUpperCase() === "DONE").length
        : 0;
      const allTasksDone = tasksTotal > 0 && tasksDone === tasksTotal;

      const visible = statusCode === "ACTIVE" && isMember;
      ui.completeProjectBtn.hidden = !visible;
      if (!visible) {
        ui.completeProjectBtn.disabled = true;
        ui.completeProjectBtn.title = "";
      } else {
        const readyForSubmit = professorAccepted && allTasksDone;
        ui.completeProjectBtn.disabled = !readyForSubmit;
        const reasons = [];
        if (!professorAccepted) reasons.push("преподаватель еще не подтвердил участие");
        if (tasksTotal === 0) reasons.push("нет задач для проверки");
        if (tasksTotal > 0 && tasksDone < tasksTotal) reasons.push(`выполнено задач ${tasksDone}/${tasksTotal}`);
        ui.completeProjectBtn.title = readyForSubmit ? "" : `Нельзя завершить: ${reasons.join(", ")}`;
      }
    }

    renderProfessorInviteArea(professorStatusCode);
  }

  function renderProfessorInviteArea(professorStatusCode) {
    const status = String(professorStatusCode || state.project?.professor_review_status || "NONE").toUpperCase();
    const currentUser = String(localStorage.getItem(LS_USER) || "");
    const isInvitedProfessor = String(state.project?.professor_id || "") === currentUser;

    if (ui.professorIdentity) {
      if (state.professorSummary && state.professorSummary.user_id) {
        const fullName = state.professorSummary.full_name || state.professorSummary.email || "Преподаватель";
        const dep = state.professorSummary.department_code ? ` · ${state.professorSummary.department_code}` : "";
        ui.professorIdentity.hidden = false;
        ui.professorIdentity.innerHTML =
          `<div class="member-avatar">${escapeHTML(initials(fullName, state.professorSummary.email))}</div>` +
          `<div><strong>${escapeHTML(fullName)}</strong><small>${escapeHTML(state.professorSummary.email || "")}${escapeHTML(dep)}</small></div>`;
      } else {
        ui.professorIdentity.hidden = true;
        ui.professorIdentity.innerHTML = "";
      }
    }

    if (!ui.professorInviteHint) return;

    if (status === "PENDING") {
      if (isInvitedProfessor) {
        ui.professorInviteHint.innerHTML = 'У вас есть приглашение на ревью. Откройте страницу <a href="/dev/professor/reviews">/dev/professor/reviews</a> и примите его.';
      } else {
        ui.professorInviteHint.innerHTML = 'Ожидаем подтверждения преподавателя в его кабинете ревью.';
      }
    } else if (status === "ACCEPTED") {
      ui.professorInviteHint.innerHTML = "Преподаватель подтвердил участие в ревью.";
    } else if (status === "REJECTED") {
      ui.professorInviteHint.innerHTML = "Преподаватель отклонил приглашение. Выберите другого преподавателя.";
    } else {
      ui.professorInviteHint.innerHTML = "Преподаватель пока не приглашён.";
    }
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
    const members = allMembers();
    const filtered = members.filter((m) => {
      if (!query) return true;
      const hay = `${getDisplayName(m.user_id)} ${m.user_id} ${m.position_name || ""} ${m.position_code || ""}`.toLowerCase();
      return hay.includes(query);
    });

    ui.teamTableBody.innerHTML = "";

    if (filtered.length === 0) {
      ui.teamTableBody.innerHTML = '<tr><td colspan="5"><div class="empty-state">Нет участников под текущий фильтр.</div></td></tr>';
      return;
    }

    const currentUser = String(localStorage.getItem(LS_USER) || "");
    const canManageTeam = isCurrentUserLead();

    filtered.forEach((m) => {
      const status = String(m.status || "").toUpperCase();
      const statusClass = status === "APPLIED" ? "invited" : status.toLowerCase();
      const statusLabel = status === "APPLIED" ? "INVITED" : status;
      const github = `https://github.com/${slugify(getDisplayName(m.user_id))}`;
      const isLeadRow = String(m.user_id) === String(state.project?.created_by || "");
      const roleOptions = isLeadRow
        ? `<option value="">Тимлид</option>${projectPositionOptions("").replace('<option value="">Выберите роль</option>', "")}`
        : projectPositionOptions(m.position_id || "");
      const roleSelectDisabled = isLeadRow || status !== "ACTIVE";
      const canApprove = canManageTeam && status === "APPLIED";
      const canRejectApplication = canManageTeam && status === "APPLIED";
      const canSetPosition = canManageTeam && status === "ACTIVE" && !isLeadRow && state.positions.length > 0;
      const canRespondInvite = status === "INVITED" && String(m.user_id) === currentUser;
      const canManagePerms = status === "ACTIVE" && !isLeadRow;

      const row = document.createElement("tr");
      row.setAttribute("data-user-id", m.user_id || "");
      row.innerHTML =
        `<td>` +
          `<div class="user-cell">` +
            `<div class="member-avatar">${escapeHTML(initials(getDisplayName(m.user_id), m.user_id))}</div>` +
            `<div><strong>${escapeHTML(getDisplayName(m.user_id))}</strong><small>${escapeHTML(getDisplaySubline(m.user_id))}</small></div>` +
          `</div>` +
        `</td>` +
        `<td><span class="status-badge ${statusClass}">${escapeHTML(statusLabel)}</span></td>` +
        `<td><select class="member-role-select" ${roleSelectDisabled ? "disabled" : ""}>${roleOptions}</select></td>` +
        `<td><a class="meta-link" href="${escapeHTML(github)}" target="_blank" rel="noreferrer">${escapeHTML(github.replace("https://", ""))}</a></td>` +
        `<td>` +
          `<div class="task-toolbar">` +
            `<button class="ghost-btn" data-member-action="approve" ${canApprove ? "" : "disabled"}>Одобрить</button>` +
            `<button class="ghost-btn" data-member-action="reject-application" ${canRejectApplication ? "" : "disabled"}>Отклонить</button>` +
            `<button class="ghost-btn" data-member-action="set-position" ${canSetPosition ? "" : "disabled"}>Сменить роль</button>` +
            `<button class="ghost-btn" data-member-action="accept-invite" ${canRespondInvite ? "" : "disabled"}>Принять</button>` +
            `<button class="ghost-btn" data-member-action="reject-invite" ${canRespondInvite ? "" : "disabled"}>Отклонить</button>` +
            `<button class="ghost-btn" data-member-action="permissions" ${canManagePerms ? "" : "disabled"}>Права</button>` +
          `</div>` +
        `</td>`;

      ui.teamTableBody.appendChild(row);
    });
  }

  function renderInviteCandidates() {
    if (!ui.inviteCandidatesList) return;
    ui.inviteCandidatesList.innerHTML = "";

    if (!Array.isArray(state.studentCandidates) || state.studentCandidates.length === 0) {
      ui.inviteCandidatesList.innerHTML = '<div class="empty-state">Подходящие студенты не найдены.</div>';
      return;
    }

    state.studentCandidates.forEach((item) => {
      rememberUser(item.user_id, item.full_name, item.email);
      const row = document.createElement("div");
      row.className = "invite-row";
      row.innerHTML =
        `<div class="invite-user">` +
          `<div class="member-avatar">${escapeHTML(initials(item.full_name, item.email))}</div>` +
          `<div><strong>${escapeHTML(item.full_name || item.email)}</strong><small>${escapeHTML(item.email || "")} · ${escapeHTML(item.department_code || "-")}</small></div>` +
        `</div>` +
        `<input class="invite-comment-input" type="text" placeholder="Комментарий к приглашению" />` +
        `<button class="ghost-btn" data-invite-user="${escapeHTML(item.user_id)}">Пригласить</button>`;
      ui.inviteCandidatesList.appendChild(row);
    });
  }

  function renderProfessorSearchResults() {
    if (!ui.professorSearchResults) return;
    ui.professorSearchResults.innerHTML = "";
    const raw = String(ui.professorSearchInput ? ui.professorSearchInput.value : "").trim();
    const canAssignByRawID = /^[0-9a-f-]{36}$/i.test(raw);

    if (!Array.isArray(state.professorCandidates) || state.professorCandidates.length === 0) {
      ui.professorSearchResults.hidden = true;
      ui.assignProfessorBtn.disabled = !(state.selectedProfessorID || canAssignByRawID);
      return;
    }

    state.professorCandidates.forEach((item) => {
      rememberUser(item.user_id, item.full_name, item.email);
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "search-result";
      btn.setAttribute("data-prof-id", item.user_id);
      btn.innerHTML = `<strong>${escapeHTML(item.full_name || item.email)}</strong><span>${escapeHTML(item.email || "")} · ${escapeHTML(item.department_code || "-")}</span>`;
      ui.professorSearchResults.appendChild(btn);
    });
    ui.professorSearchResults.hidden = false;
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
    const isLead = isCurrentUserLead();
    const isAssignee = isCurrentUserAssignee(task);
    let controlsHTML = "";

    if (isLead) {
      controlsHTML =
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
          (isAssignee && status === "OPEN" ? `<button class="primary-btn" data-task-action="claim">Взять</button>` : "") +
          (isAssignee && status === "IN_PROGRESS" ? `<button class="primary-btn" data-task-action="complete-open">Выполнено</button>` : "") +
        `</div>`;
    } else if (isAssignee) {
      if (status === "OPEN") {
        controlsHTML =
          `<div class="task-controls student-flow">` +
            `<button class="primary-btn" data-task-action="claim">Взять</button>` +
          `</div>`;
      } else if (status === "IN_PROGRESS") {
        controlsHTML =
          `<div class="task-controls student-flow">` +
            `<button class="primary-btn" data-task-action="complete-open">Выполнено</button>` +
          `</div>`;
      } else {
        controlsHTML = `<div class="task-controls student-flow"><span class="task-note done">Задача закрыта</span></div>`;
      }
    } else {
      controlsHTML = `<div class="task-controls student-flow"><span class="task-note">Ожидайте назначения или обновлений.</span></div>`;
    }

    card.innerHTML =
      `<h4>${escapeHTML(task.title || "Без названия")}</h4>` +
      `<p>${escapeHTML(task.description || "Описание отсутствует")}</p>` +
      `<p>Роль: ${escapeHTML(task.position_name || task.position_code || "-")}</p>` +
      `<p>Исполнитель: ${escapeHTML(task.assignee_user_id ? getDisplayName(task.assignee_user_id) : "не назначен")}</p>` +
      `<div class="task-tags">${tags.map((t) => `<span class="tag">${escapeHTML(t)}</span>`).join("")}</div>` +
      controlsHTML +
      `<div class="task-timeline-wrap">` +
        `<p class="task-timeline-head">Лента задачи</p>` +
        renderTaskActivityTimeline(task.id || "") +
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

  function renderCriteriaView() {
    if (!ui.criteriaListView) return;
    const criteria = Array.isArray(state.criteria) ? state.criteria : [];
    const count = criteria.length;
    if (ui.criteriaCountMeta) {
      ui.criteriaCountMeta.textContent = `${count} ${count === 1 ? "критерий" : count < 5 ? "критерия" : "критериев"}`;
    }

    if (!criteria.length) {
      ui.criteriaListView.innerHTML = '<div class="empty-state">Преподаватель еще не добавил критерии.</div>';
    } else {
      ui.criteriaListView.innerHTML = criteria
        .map((item, idx) => {
          const weight = Number(item.weight || 0);
          return (
            `<article class="criteria-item">` +
              `<div class="criteria-idx">${idx + 1}</div>` +
              `<div>` +
                `<strong>${escapeHTML(item.title || "Без названия")}</strong>` +
                `<p>${escapeHTML(item.description || "Описание отсутствует")}</p>` +
              `</div>` +
              `<span class="criteria-weight">Вес ${escapeHTML(weight > 0 ? weight : 1)}</span>` +
            `</article>`
          );
        })
        .join("");
    }

    if (ui.criteriaReviewHint) {
      const status = projectStatusCode();
      const summary = reviewSummaryData();
      if (summary.hasReview) {
        ui.criteriaReviewHint.textContent = `Проверено: ${summary.met}/${summary.total}. Итоговый балл: ${summary.score}/5.0 (${summary.passPercent}%).`;
      } else if (status === "REVIEW" || status === "GRADING" || status === "ARCHIVE") {
        ui.criteriaReviewHint.textContent = "Проект на этапе ревью. Итоговая оценка будет доступна после проверки преподавателем.";
      } else {
        ui.criteriaReviewHint.textContent = "Оценивание появится после завершения проекта и начала ревью.";
      }
    }
  }

  function renderReviewView() {
    if (!ui.reviewCriteriaList) return;
    const criteria = Array.isArray(state.criteria) ? state.criteria : [];
    const grading = gradingByCriterion();
    const summary = reviewSummaryData();
    const status = projectStatusCode();

    if (!canViewWorkspace() && canViewFinalGrade()) {
      ui.reviewCriteriaList.innerHTML = '<div class="empty-state">Итоговая оценка опубликована. Детали по критериям видны только участникам команды.</div>';
    } else if (!criteria.length) {
      ui.reviewCriteriaList.innerHTML = '<div class="empty-state">Критерии пока не настроены преподавателем.</div>';
    } else {
      ui.reviewCriteriaList.innerHTML = criteria
        .map((criterion) => {
          const id = String(criterion.id || "");
          const item = grading.get(id) || { isMet: null, comment: "" };
          const yesActive = item.isMet === true ? "active" : "";
          const noActive = item.isMet === false ? "active" : "";
          const comment = String(item.comment || "").trim();
          const commentClass = comment
            ? item.isMet === false
              ? "review-comment-box negative"
              : "review-comment-box"
            : "review-comment-box empty";
          const commentText = comment || "Комментарий не оставлен.";

          return (
            `<article class="review-criterion">` +
              `<div class="review-criterion-head">` +
                `<strong>${escapeHTML(criterion.title || "Без названия")}</strong>` +
                `<div class="review-bool">` +
                  `<span class="yes ${yesActive}">Да</span>` +
                  `<span class="no ${noActive}">Нет</span>` +
                `</div>` +
              `</div>` +
              `<p>${escapeHTML(criterion.description || "Описание отсутствует")}</p>` +
              `<div class="${commentClass}">${escapeHTML(commentText)}</div>` +
            `</article>`
          );
        })
        .join("");
    }

    if (ui.reviewStatusPill) {
      if (status === "ARCHIVE" && summary.hasReview) {
        ui.reviewStatusPill.className = "status-pill active";
        ui.reviewStatusPill.textContent = "Проверено преподавателем";
      } else if (status === "GRADING") {
        ui.reviewStatusPill.className = "status-pill review";
        ui.reviewStatusPill.textContent = summary.hasReview ? "Идет оценивание" : "На оценивании";
      } else {
        ui.reviewStatusPill.className = "status-pill muted";
        ui.reviewStatusPill.textContent = "Оценка не опубликована";
      }
    }

    if (ui.reviewIntro) {
      if (summary.hasReview) {
        ui.reviewIntro.textContent = `Ревью по критериям сохранено. Выполнено: ${summary.met}/${summary.total}.`;
      } else if (status === "GRADING") {
        ui.reviewIntro.textContent = "Проект отправлен преподавателю на оценивание. Результаты появятся после проверки.";
      } else {
        ui.reviewIntro.textContent = "После завершения проекта преподаватель выставляет оценки по критериям. Здесь отображаются результаты ревью.";
      }
    }

    if (ui.reviewGauge) {
      ui.reviewGauge.style.setProperty("--review-percent", String(summary.passPercent));
    }
    if (ui.reviewGaugeValue) {
      ui.reviewGaugeValue.textContent = `${summary.passPercent}%`;
    }
    if (ui.reviewSummaryScore) {
      ui.reviewSummaryScore.innerHTML = `${escapeHTML(summary.score)} <span>/ 5.0</span>`;
    }
    if (ui.reviewSummaryMet) {
      ui.reviewSummaryMet.textContent = `${summary.met} / ${summary.total}`;
    }
    if (ui.reviewSummaryMissed) {
      ui.reviewSummaryMissed.textContent = `${summary.missed}`;
    }
    if (ui.reviewSummaryDate) {
      ui.reviewSummaryDate.textContent = summary.reviewedAt ? formatDate(summary.reviewedAt.toISOString()) : "-";
    }
    if (ui.reviewSummaryReviewer) {
      ui.reviewSummaryReviewer.textContent = summary.reviewer;
    }
    if (ui.reviewOverallComment) {
      ui.reviewOverallComment.className = summary.hasReview ? "review-comment-box" : "empty-state";
      ui.reviewOverallComment.textContent = summary.overall;
    }
  }

  function renderOverview() {
    renderLifecycle();
    renderAbout();
    renderStackChips();
    renderTeamMini();
    renderReadiness();
    renderActivity();
  }

  function renderAccessMode() {
    const applyMode = canApplyToProject();
    const workspaceMode = canViewWorkspace();
    const visibleViews = new Set(allowedViews());

    if (ui.applyCard) {
      ui.applyCard.hidden = !applyMode;
    }
    if (ui.stackCard) {
      ui.stackCard.hidden = !workspaceMode;
    }
    if (ui.activityCard) {
      ui.activityCard.hidden = !workspaceMode;
    }
    if (ui.teamMiniCard) {
      ui.teamMiniCard.hidden = !workspaceMode;
    }
    if (ui.pipelineCard) {
      ui.pipelineCard.hidden = !workspaceMode;
    }

    if (ui.applyHint) {
      ui.applyHint.textContent = applyMode
        ? "Оставьте короткий комментарий и отправьте заявку в команду проекта."
        : "Оставьте короткий комментарий и отправьте заявку в проект.";
    }
    if (ui.applyProjectBtn) {
      ui.applyProjectBtn.disabled = !applyMode;
    }

    ui.tabButtons.forEach((btn) => {
      const view = btn.getAttribute("data-view") || "overview";
      btn.hidden = !visibleViews.has(view);
    });
    ui.switchViewButtons.forEach((btn) => {
      btn.hidden = !workspaceMode;
    });

    if (!visibleViews.has(state.activeView)) {
      state.activeView = "overview";
    }

    ui.tabButtons.forEach((btn) => {
      const view = btn.getAttribute("data-view") || "overview";
      btn.classList.toggle("active", visibleViews.has(view) && view === state.activeView);
    });

    const viewMap = {
      overview: ui.viewOverview,
      team: ui.viewTeam,
      invite: ui.viewInvite,
      tasks: ui.viewTasks,
      criteria: ui.viewCriteria,
      review: ui.viewReview,
      edit: ui.viewEdit,
    };
    Object.entries(viewMap).forEach(([key, el]) => {
      if (!el) return;
      el.classList.toggle("active", visibleViews.has(key) && key === state.activeView);
      el.hidden = !visibleViews.has(key);
    });
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
    renderInviteCandidates();
    renderTasks();
    renderCriteriaView();
    renderReviewView();
    bindProjectMetaToUI();
    renderTaskModalSelects();
    renderAccessMode();
  }

  function setView(viewName) {
    const visibleViews = allowedViews();
    const target = visibleViews.includes(viewName)
      ? viewName
      : "overview";

    state.activeView = target;

    const viewMap = {
      overview: ui.viewOverview,
      team: ui.viewTeam,
      invite: ui.viewInvite,
      tasks: ui.viewTasks,
      criteria: ui.viewCriteria,
      review: ui.viewReview,
      edit: ui.viewEdit,
    };

    Object.entries(viewMap).forEach(([key, el]) => {
      const allowed = visibleViews.includes(key);
      el.classList.toggle("active", allowed && key === target);
      el.hidden = !allowed;
    });

    ui.tabButtons.forEach((btn) => {
      const view = btn.getAttribute("data-view");
      btn.classList.toggle("active", view === target);
    });

    if (target === "edit") {
      bindProjectMetaToUI();
    }
    if (target === "invite") {
      loadStudentCandidates(ui.inviteSearchInput ? ui.inviteSearchInput.value : "")
        .catch((err) => setNotice(err.message || String(err), true));
    }
  }

  function initialViewFromURL() {
    const params = new URLSearchParams(window.location.search || "");
    const raw = String(params.get("view") || "").trim().toLowerCase();
    if (!raw) return "overview";
    return allowedViews().includes(raw)
      ? raw
      : "overview";
  }

  function openModal(modal) {
    if (!modal) return;
    modal.hidden = false;
    document.body.style.overflow = "hidden";
  }

  function closeModal(modal) {
    if (!modal) return;
    modal.hidden = true;
    if (modal === ui.taskResultModal) {
      clearTaskResultForm();
    }
    if (
      ui.taskModal.hidden &&
      ui.permissionsModal.hidden &&
      (!ui.taskResultModal || ui.taskResultModal.hidden)
    ) {
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

  function clearTaskResultForm() {
    state.currentResultTaskID = "";
    if (ui.taskResultTarget) {
      ui.taskResultTarget.textContent = "Задача";
    }
    if (ui.taskResultCommentInput) {
      ui.taskResultCommentInput.value = "";
    }
    if (ui.taskResultAttachmentsInput) {
      ui.taskResultAttachmentsInput.value = "";
    }
    if (ui.taskResultSubmitBtn) {
      ui.taskResultSubmitBtn.disabled = false;
      ui.taskResultSubmitBtn.textContent = "Готово";
    }
  }

  function openTaskResultModal(taskID) {
    const task = state.tasks.find((item) => String(item.id) === String(taskID));
    if (!task) {
      throw new Error("Задача не найдена.");
    }
    if (!isCurrentUserAssignee(task)) {
      throw new Error("Фиксировать результат может только назначенный исполнитель.");
    }
    const status = String(task.status || "").toUpperCase();
    if (status !== "IN_PROGRESS") {
      throw new Error("Задача должна быть в статусе IN_PROGRESS.");
    }

    state.currentResultTaskID = String(taskID);
    if (ui.taskResultTarget) {
      ui.taskResultTarget.textContent = `${task.title || "Задача"} · ${task.position_name || task.position_code || "без роли"}`;
    }
    if (ui.taskResultCommentInput) {
      ui.taskResultCommentInput.value = "";
    }
    if (ui.taskResultAttachmentsInput) {
      ui.taskResultAttachmentsInput.value = "";
    }
    openModal(ui.taskResultModal);
  }

  function collectTaskResultAttachments() {
    const raw = String(ui.taskResultAttachmentsInput ? ui.taskResultAttachmentsInput.value : "");
    const parts = raw
      .split(/\n|,/g)
      .map((value) => value.trim())
      .filter(Boolean);
    const uniq = [];
    const seen = new Set();
    parts.forEach((item) => {
      if (seen.has(item)) return;
      seen.add(item);
      uniq.push(item);
    });
    return uniq.slice(0, 10);
  }

  async function submitTaskResultFromModal() {
    const taskID = String(state.currentResultTaskID || "");
    if (!taskID) {
      throw new Error("Не выбрана задача для фиксации результата.");
    }

    const comment = String(ui.taskResultCommentInput ? ui.taskResultCommentInput.value : "").trim();
    const attachments = collectTaskResultAttachments();
    const submitBtn = ui.taskResultSubmitBtn;
    const prevTitle = submitBtn ? submitBtn.textContent : "Готово";
    if (submitBtn) {
      submitBtn.disabled = true;
      submitBtn.textContent = "Сохраняем...";
    }

    try {
      await request("POST", `/v2/projects/${state.projectID}/tasks/${taskID}/complete`, {
        comment,
        attachments,
      });
    } finally {
      if (submitBtn) {
        submitBtn.disabled = false;
        submitBtn.textContent = prevTitle;
      }
    }

    closeModal(ui.taskResultModal);
    setNotice("Результат зафиксирован. Задача переведена в DONE.", false);
    await refreshData();
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
      if (status === "IN_PROGRESS") {
        await request("PATCH", `/v2/projects/${state.projectID}/tasks/${created.id}/status`, {
          status: "IN_PROGRESS",
        });
      }
    }

    closeModal(ui.taskModal);
    clearTaskModalForm();
    setNotice("Задача создана.", false);
    await refreshData();
  }

  function openPermissionsModal(userID) {
    const member = allMembers().find((m) => String(m.user_id) === String(userID));
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

  async function onUploadCover(file) {
    if (!file) return;
    const allowed = new Set(["image/jpeg", "image/png", "image/webp"]);
    if (!allowed.has(String(file.type || "").toLowerCase())) {
      throw new Error("Поддерживаются JPG/PNG/WEBP.");
    }
    if (Number(file.size || 0) > 12 * 1024 * 1024) {
      throw new Error("Файл слишком большой (макс. 12MB).");
    }

    const form = new FormData();
    form.append("image", file);
    const project = await requestForm("POST", `/v2/projects/${state.projectID}/image`, form);
    state.project = project;
    renderProjectCovers();
    setCoverStatus("Обложка проекта обновлена.", false);
  }

  async function onRemoveCover() {
    const project = await request("DELETE", `/v2/projects/${state.projectID}/image`);
    state.project = project;
    renderProjectCovers();
    setCoverStatus("Обложка проекта удалена. Используется вариант по умолчанию.", false);
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

  async function onDeleteProject() {
    if (!isCurrentUserCreator()) {
      throw new Error("Удалять проект может только его создатель.");
    }
    const confirmed = window.confirm("Удалить проект без возможности восстановления?");
    if (!confirmed) return;

    await request("DELETE", `/v2/projects/${state.projectID}`);
    localStorage.removeItem(LS_SELECTED_PROJECT);
    window.location.href = "/dev/projects";
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
    const statusBadge = row.querySelector(".status-badge");
    const status = String(statusBadge ? statusBadge.textContent : "").trim().toUpperCase();

    if (action === "permissions") {
      if (status !== "ACTIVE") {
        throw new Error("Права можно настраивать только для ACTIVE участников.");
      }
      openPermissionsModal(userID);
      return;
    }

    if (action === "accept-invite" || action === "reject-invite") {
      await request("POST", `/v2/projects/${state.projectID}/members/respond`, {
        accept: action === "accept-invite",
      });
      setNotice(action === "accept-invite" ? "Приглашение принято." : "Приглашение отклонено.", false);
      await refreshData();
      return;
    }

    if (action === "approve") {
      const payload = positionID ? { position_id: positionID } : {};
      await request("POST", `/v2/projects/${state.projectID}/members/${userID}/approve`, payload);
      setNotice(`Участник ${shortID(userID)} принят в команду.`, false);
      await refreshData();
      return;
    }

    if (action === "reject-application") {
      await request("POST", `/v2/projects/${state.projectID}/members/${userID}/reject`, {});
      setNotice(`Заявка участника ${shortID(userID)} отклонена.`, false);
      await refreshData();
      return;
    }

    if (action === "set-position") {
      if (!positionID) {
        throw new Error("Выберите роль участника.");
      }
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
      return;
    }

    if (action === "complete-open") {
      openTaskResultModal(taskID);
    }
  }

  async function onOpenRecruitment() {
    await request("POST", `/v2/projects/${state.projectID}/recruitment/open`, {});
    setNotice("Набор в проект открыт.", false);
    await refreshData();
  }

  async function loadStudentCandidates(query) {
    const q = String(query || "").trim();
    const encoded = encodeURIComponent(q);
    const items = await request("GET", `/v2/projects/${state.projectID}/candidates/students?q=${encoded}&limit=60`);
    state.studentCandidates = Array.isArray(items) ? items : [];
    renderInviteCandidates();
  }

  async function inviteCandidate(userID, comment) {
    await request("POST", `/v2/projects/${state.projectID}/members/invite`, {
      user_id: userID,
      comment: String(comment || "").trim(),
    });
    setNotice("Приглашение отправлено.", false);
    await refreshData();
    if (state.activeView === "invite") {
      await loadStudentCandidates(ui.inviteSearchInput ? ui.inviteSearchInput.value : "");
    }
  }

  async function searchProfessors(query) {
    const q = String(query || "").trim();
    if (q.length < 2) {
      state.professorCandidates = [];
      renderProfessorSearchResults();
      return;
    }
    const encoded = encodeURIComponent(q);
    const items = await request("GET", `/v2/projects/${state.projectID}/candidates/professors?q=${encoded}&limit=20`);
    state.professorCandidates = Array.isArray(items) ? items : [];
    renderProfessorSearchResults();
  }

  async function onAssignProfessor() {
    const fallbackUUID = String(ui.professorSearchInput.value || "").trim();
    const professorID = state.selectedProfessorID || fallbackUUID;
    if (!professorID) {
      throw new Error("Выберите преподавателя из подсказок.");
    }

    await request("POST", `/v2/projects/${state.projectID}/professor`, {
      professor_id: professorID,
    });
    state.selectedProfessorID = "";
    state.professorCandidates = [];
    ui.professorSearchInput.value = "";
    ui.professorSearchResults.hidden = true;
    ui.professorSearchResults.innerHTML = "";
    ui.assignProfessorBtn.disabled = true;
    setNotice("Приглашение преподавателю отправлено.", false);
    await refreshData();
  }

  async function onApproveProject() {
    await request("POST", `/v2/projects/${state.projectID}/approve`, {});
    setNotice("Проект переведен в ACTIVE.", false);
    await refreshData();
  }

  async function onSubmitProjectForGrading() {
    const confirmed = window.confirm("Отправить проект преподавателю на оценивание? После этого статус станет GRADING.");
    if (!confirmed) return;

    await request("POST", `/v2/projects/${state.projectID}/grading/submit`, {});
    setNotice("Проект отправлен на оценивание преподавателю.", false);
    await refreshData();
  }

  async function onApplyToProject() {
    if (projectStatusCode() !== "RECRUITMENT") {
      throw new Error("Подать заявку можно только на этапе набора (RECRUITMENT).");
    }
    if (!canApplyToProject()) {
      throw new Error("Вы уже связаны с этим проектом.");
    }

    const comment = String(ui.applyCommentInput ? ui.applyCommentInput.value : "").trim();
    await request("POST", `/v2/projects/${state.projectID}/members/apply`, { comment });

    if (ui.applyCommentInput) {
      ui.applyCommentInput.value = "";
    }
    if (ui.applyHint) {
      ui.applyHint.textContent = "Заявка отправлена. Ожидайте решения тимлида.";
    }

    setNotice("Заявка отправлена.", false);
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
    rememberUser(localStorage.getItem(LS_USER), localStorage.getItem(LS_STUDENT_NAME), localStorage.getItem(LS_STUDENT_EMAIL));
    rememberUser(state.project?.created_by, state.project?.created_by_name, state.project?.created_by_email);

    if (!canViewWorkspace()) {
      state.stacks = [];
      state.positions = [];
      state.members = [];
      state.readiness = null;
      state.criteria = [];
      state.tasks = [];
      state.gradingItems = [];
      state.taskActivities = [];
      state.professorSummary = null;
      localStorage.setItem(LS_SELECTED_PROJECT, JSON.stringify({
        ...state.project,
        _nav_section: currentStudentSection(),
      }));
      renderAll();
      return;
    }

    const [stacks, positions, members, readiness, criteria, tasks, gradingResp, taskActivityResp, professorResp] = await Promise.all([
      loadOptional("stacks", "GET", `/v2/projects/${state.projectID}/stacks`, []),
      loadOptional("positions", "GET", `/v2/projects/${state.projectID}/positions`, []),
      loadOptional("members", "GET", `/v2/projects/${state.projectID}/members`, []),
      loadOptional("readiness", "GET", `/v2/projects/${state.projectID}/readiness`, null),
      loadOptional("criteria", "GET", `/v2/projects/${state.projectID}/criteria`, []),
      loadOptional("tasks", "GET", `/v2/projects/${state.projectID}/tasks`, []),
      loadOptional("grading", "GET", `/v2/projects/${state.projectID}/grading`, { items: [] }),
      loadOptional("task activity", "GET", `/v2/projects/${state.projectID}/tasks/activity`, { items: [] }),
      loadOptional("assigned professor", "GET", `/v2/projects/${state.projectID}/professor`, { professor: null }),
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
    state.members.forEach((member) => {
      rememberUser(member.user_id, member.full_name, member.email);
    });
    state.readiness = readiness && typeof readiness === "object" ? readiness : null;
    state.criteria = Array.isArray(criteria) ? criteria : [];
    state.tasks = Array.isArray(tasks) ? tasks : [];
    if (Array.isArray(gradingResp)) {
      state.gradingItems = gradingResp;
    } else {
      state.gradingItems = gradingResp && Array.isArray(gradingResp.items) ? gradingResp.items : [];
    }
    if (Array.isArray(taskActivityResp)) {
      state.taskActivities = taskActivityResp;
    } else {
      state.taskActivities = taskActivityResp && Array.isArray(taskActivityResp.items) ? taskActivityResp.items : [];
    }
    state.professorSummary = professorResp && typeof professorResp === "object" ? professorResp.professor || null : null;
    if (state.professorSummary && state.professorSummary.user_id) {
      rememberUser(state.professorSummary.user_id, state.professorSummary.full_name, state.professorSummary.email);
    }

    localStorage.setItem(LS_SELECTED_PROJECT, JSON.stringify({
      ...state.project,
      _nav_section: currentStudentSection(),
    }));

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

    [ui.taskModal, ui.permissionsModal, ui.taskResultModal].forEach((modal) => {
      if (!modal) return;
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
      if (ui.taskResultModal && !ui.taskResultModal.hidden) closeModal(ui.taskResultModal);
    });
  }

  function wireEvents() {
    ui.logoutBtn.addEventListener("click", () => {
      auth.logout();
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

    if (ui.uploadCoverBtn && ui.editCoverInput) {
      ui.uploadCoverBtn.addEventListener("click", () => {
        ui.editCoverInput.click();
      });
      ui.editCoverInput.addEventListener("change", async () => {
        const file = ui.editCoverInput.files && ui.editCoverInput.files[0] ? ui.editCoverInput.files[0] : null;
        if (!file) return;
        setButtonLoading(ui.uploadCoverBtn, true, "Загрузка...");
        try {
          await onUploadCover(file);
          setNotice("Обложка проекта обновлена.", false);
        } catch (err) {
          setCoverStatus(err.message || String(err), true);
          setNotice(err.message || String(err), true);
        } finally {
          ui.editCoverInput.value = "";
          setButtonLoading(ui.uploadCoverBtn, false, "Загрузить");
        }
      });
    }

    if (ui.removeCoverBtn) {
      ui.removeCoverBtn.addEventListener("click", async () => {
        setButtonLoading(ui.removeCoverBtn, true, "Удаление...");
        try {
          await onRemoveCover();
          setNotice("Обложка удалена.", false);
        } catch (err) {
          setCoverStatus(err.message || String(err), true);
          setNotice(err.message || String(err), true);
        } finally {
          setButtonLoading(ui.removeCoverBtn, false, "Удалить");
        }
      });
    }

    if (ui.deleteProjectBtn) {
      ui.deleteProjectBtn.addEventListener("click", async () => {
        try {
          await onDeleteProject();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

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

    if (ui.openRecruitmentBtn) {
      ui.openRecruitmentBtn.addEventListener("click", async () => {
        try {
          await onOpenRecruitment();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

    if (ui.assignProfessorBtn && ui.professorSearchInput && ui.professorSearchResults) {
      ui.assignProfessorBtn.addEventListener("click", async () => {
        try {
          await onAssignProfessor();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });

      ui.professorSearchInput.addEventListener("input", () => {
        const raw = String(ui.professorSearchInput.value || "").trim();
        state.selectedProfessorID = "";
        ui.assignProfessorBtn.disabled = !/^[0-9a-f-]{36}$/i.test(raw);
        if (state.professorSearchTimer) clearTimeout(state.professorSearchTimer);
        state.professorSearchTimer = setTimeout(() => {
          searchProfessors(ui.professorSearchInput.value).catch((err) => setNotice(err.message || String(err), true));
        }, 250);
      });

      ui.professorSearchResults.addEventListener("click", (e) => {
        const btn = e.target.closest("button[data-prof-id]");
        if (!btn) return;
        const profID = btn.getAttribute("data-prof-id") || "";
        const item = state.professorCandidates.find((x) => String(x.user_id) === String(profID));
        if (!item) return;
        state.selectedProfessorID = profID;
        ui.professorSearchInput.value = `${item.full_name || item.email} <${item.email}>`;
        ui.professorSearchResults.hidden = true;
        ui.assignProfessorBtn.disabled = false;
      });

      document.addEventListener("click", (e) => {
        if (!ui.professorSearchResults || ui.professorSearchResults.hidden) return;
        if (e.target === ui.professorSearchInput || ui.professorSearchResults.contains(e.target)) return;
        ui.professorSearchResults.hidden = true;
      });
    }

    if (ui.inviteRefreshBtn) {
      ui.inviteRefreshBtn.addEventListener("click", async () => {
        try {
          await loadStudentCandidates(ui.inviteSearchInput ? ui.inviteSearchInput.value : "");
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

    if (ui.inviteSearchInput) {
      ui.inviteSearchInput.addEventListener("input", () => {
        if (state.studentSearchTimer) clearTimeout(state.studentSearchTimer);
        state.studentSearchTimer = setTimeout(() => {
          loadStudentCandidates(ui.inviteSearchInput.value).catch((err) => setNotice(err.message || String(err), true));
        }, 250);
      });
    }

    if (ui.inviteCandidatesList) {
      ui.inviteCandidatesList.addEventListener("click", async (e) => {
        const btn = e.target.closest("button[data-invite-user]");
        if (!btn) return;
        const userID = btn.getAttribute("data-invite-user") || "";
        const row = btn.closest(".invite-row");
        const commentInput = row ? row.querySelector(".invite-comment-input") : null;
        const comment = commentInput ? commentInput.value : "";
        try {
          await inviteCandidate(userID, comment);
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

    if (ui.approveProjectBtn) {
      ui.approveProjectBtn.addEventListener("click", async () => {
        try {
          await onApproveProject();
        } catch (err) {
          if (err.data && typeof err.data === "object" && err.data.readiness) {
            setNotice(
              `Недостаточно условий: роли ${err.data.readiness.active_members}/${err.data.readiness.required_members}, критерии ${err.data.readiness.criteria_count}.`,
              true
            );
          } else {
            setNotice(err.message || String(err), true);
          }
        }
      });
    }
    if (ui.completeProjectBtn) {
      ui.completeProjectBtn.addEventListener("click", async () => {
        try {
          await onSubmitProjectForGrading();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }
    if (ui.applyProjectBtn) {
      ui.applyProjectBtn.addEventListener("click", async () => {
        try {
          await onApplyToProject();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

    ui.openTaskModalBtn.addEventListener("click", openTaskModal);
    ui.taskModalPositionSelect.addEventListener("change", syncTaskModalAssignees);
    ui.taskModalCreateBtn.addEventListener("click", async () => {
      try {
        await createTaskFromModal();
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });
    if (ui.taskResultSubmitBtn) {
      ui.taskResultSubmitBtn.addEventListener("click", async () => {
        try {
          await submitTaskResultFromModal();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

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
    const claims = await auth.ensureSession("student");
    if (!claims) return;

    syncStudentSidebar();
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
      setView(initialViewFromURL());
    } catch (err) {
      setNotice(`Не удалось загрузить проект: ${err.message || String(err)}`, true);
    }
  }

  bootstrap();
})();
