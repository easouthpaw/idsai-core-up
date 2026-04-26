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
  const LS_SELECTED_PROJECT = "idsai_selected_project";
  const LS_STUDENT_SECTION = "idsai_student_section";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

  const LS_PROJECT_META_PREFIX = "idsai_project_meta:";
  const LS_TASK_META_PREFIX = "idsai_task_meta:";
  const LS_TASK_DISPLAY_LIMIT_PREFIX = "idsai_task_display_limit:";
  const LS_STAGE_HINTS_HIDDEN_PREFIX = "idsai_stage_hints_hidden:";
  const TASK_DISPLAY_LIMITS = [5, 10, 25, 50];
  const DEFAULT_TASK_DISPLAY_LIMIT = 5;

  const SYSTEM_LIFECYCLE_ROLES = new Set(["TEAM_LEAD", "MEMBER", "INVITED_MEMBER", "PROJECT_PROFESSOR"]);
  const ROLE_ASSETS = {
    CO_LEAD: "/dev/static/assets/role-co-lead.svg",
    RECRUITER: "/dev/static/assets/role-recruiter.svg",
    TASK_MANAGER: "/dev/static/assets/role-task-manager.svg",
  };
  const PERMISSION_LABELS = {
    "project.view": "Просмотр проекта",
    "project.edit": "Редактирование проекта",
    "project.invite_professor": "Приглашение преподавателя",
    "project.submit_for_review": "Отправка на проверку",
    "position.create": "Создание проектных ролей",
    "member.approve": "Управление заявками",
    "member.access.manage": "Управление ролями доступа",
    "task.view": "Просмотр задач",
    "task.create": "Создание задач",
    "task.assign": "Назначение задач",
    "task.update": "Изменение задач",
    "task.delete": "Удаление задач",
    "task.claim": "Взять задачу в работу",
    "grading.view": "Просмотр критериев и оценок",
  };
  const SYSTEM_TASK_POSITION_CODES = new Set(["TEAM_LEAD"]);
  const PRIMARY_LIFECYCLE_FLOW = ["DRAFT", "RECRUITMENT", "ACTIVE", "GRADING", "COMPLETED"];
  const DEFAULT_PROJECT_COVERS = [
    "https://images.pexels.com/photos/16129724/pexels-photo-16129724.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/17323801/pexels-photo-17323801.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/4508751/pexels-photo-4508751.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/5257576/pexels-photo-5257576.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/10499056/pexels-photo-10499056.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/12899157/pexels-photo-12899157.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
  ];
  const HINT_BADGE_HTML =
    `<span class="hint-badge">` +
      `<span class="hint-badge-icon" aria-hidden="true">` +
        `<svg viewBox="0 0 20 20" fill="none" focusable="false" aria-hidden="true">` +
          `<path d="M7.2 12.5h5.6M7.9 15.2h4.2M10 2.7a4.9 4.9 0 0 0-2.9 8.9c.5.4.8 1 .9 1.6h4c.1-.6.4-1.2.9-1.6A4.9 4.9 0 0 0 10 2.7Z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>` +
        `</svg>` +
      `</span>` +
      `<span>Подсказка</span>` +
    `</span>`;
  const ALERT_ICON_HTML =
    `<svg viewBox="0 0 24 24" fill="none" focusable="false" aria-hidden="true">` +
      `<path d="M13 16h-1v-4h1m0-4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"></path>` +
    `</svg>`;

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
    stageSpotlight: document.getElementById("stageSpotlight"),
    stageSpotlightSide: document.querySelector(".stage-spotlight-side"),
    stageSpotlightTitle: document.getElementById("stageSpotlightTitle"),
    stageSpotlightCopy: document.getElementById("stageSpotlightCopy"),
    stageCurrentBadge: document.getElementById("stageCurrentBadge"),
    stageNextBadge: document.getElementById("stageNextBadge"),
    hideStageHintsBtn: document.getElementById("hideStageHintsBtn"),
    stageHintsCollapsed: document.getElementById("stageHintsCollapsed"),
    showStageHintsBtn: document.getElementById("showStageHintsBtn"),

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
    teamStageHint: document.getElementById("teamStageHint"),
    tasksStageHint: document.getElementById("tasksStageHint"),
    criteriaStageHint: document.getElementById("criteriaStageHint"),

    openRecruitmentBtn: document.getElementById("openRecruitmentBtn"),
    professorSearchInput: document.getElementById("professorSearchInput"),
    professorSearchResults: document.getElementById("professorSearchResults"),
    professorIdentity: document.getElementById("professorIdentity"),
    professorInviteHint: document.getElementById("professorInviteHint"),
    assignProfessorBtn: document.getElementById("assignProfessorBtn"),
    pipelineStatusNote: document.getElementById("pipelineStatusNote"),
    approveProjectBtn: document.getElementById("approveProjectBtn"),
    completeProjectBtn: document.getElementById("completeProjectBtn"),

    positionForm: document.getElementById("positionForm"),
    positionNameInput: document.getElementById("positionNameInput"),
    positionCodeInput: document.getElementById("positionCodeInput"),
    positionCapacityInput: document.getElementById("positionCapacityInput"),
    openAccessRoleModalBtn: document.getElementById("openAccessRoleModalBtn"),
    teamTableBody: document.getElementById("teamTableBody"),
    inviteSearchInput: document.getElementById("inviteSearchInput"),
    inviteRefreshBtn: document.getElementById("inviteRefreshBtn"),
    inviteCandidatesList: document.getElementById("inviteCandidatesList"),

    progressBadge: document.getElementById("progressBadge"),
    taskDisplayLimitSwitch: document.getElementById("taskDisplayLimitSwitch"),
    taskDisplayLimitButtons: Array.from(document.querySelectorAll("[data-task-limit]")),
    openTaskModalBtn: document.getElementById("openTaskModalBtn"),
    todoTasks: document.getElementById("todoTasks"),
    doingTasks: document.getElementById("doingTasks"),
    doneTasks: document.getElementById("doneTasks"),
    countTodo: document.getElementById("countTodo"),
    countDoing: document.getElementById("countDoing"),
    countDone: document.getElementById("countDone"),
    taskListBody: document.getElementById("taskListBody"),
    taskListTabs: document.querySelectorAll(".task-list-tab"),
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
    previewFinalReportBtn: document.getElementById("previewFinalReportBtn"),
    downloadFinalReportBtn: document.getElementById("downloadFinalReportBtn"),
    downloadFinalReportModalBtn: document.getElementById("downloadFinalReportModalBtn"),

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
    accessRoleModal: document.getElementById("accessRoleModal"),
    finalReportPreviewModal: document.getElementById("finalReportPreviewModal"),
    finalReportPreviewFrame: document.getElementById("finalReportPreviewFrame"),
    permMemberName: document.getElementById("permMemberName"),
    permRoleDots: document.getElementById("permRoleDots"),
    permLoading: document.getElementById("permLoading"),
    permContent: document.getElementById("permContent"),
    permSystemRoles: document.getElementById("permSystemRoles"),
    permAssignableRoles: document.getElementById("permAssignableRoles"),
    permEffectivePermissions: document.getElementById("permEffectivePermissions"),
    permEffectiveCount: document.getElementById("permEffectiveCount"),
    savePermissionsBtn: document.getElementById("savePermissionsBtn"),
    accessRoleNameInput: document.getElementById("accessRoleNameInput"),
    accessRoleCodeInput: document.getElementById("accessRoleCodeInput"),
    accessRoleDescriptionInput: document.getElementById("accessRoleDescriptionInput"),
    accessRolePermissionCount: document.getElementById("accessRolePermissionCount"),
    accessRoleLoading: document.getElementById("accessRoleLoading"),
    accessRolePermissionsList: document.getElementById("accessRolePermissionsList"),
    saveAccessRoleBtn: document.getElementById("saveAccessRoleBtn"),
    teamHelperCard: document.querySelector("#view-team .helper-card"),
    professorAssignWrap: document.querySelector(".professor-assign"),

    refreshBtn: document.getElementById("refreshBtn"),
    logoutBtn: document.getElementById("logoutBtn"),
  };

  const state = {
    profile: null,
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
    myPermissions: [],
    taskListTab: "all",
    taskDisplayLimit: DEFAULT_TASK_DISPLAY_LIMIT,
    currentPermUserID: "",
    currentPermCanManageAccess: false,
    accessCatalog: [],
    accessPermissionCatalog: [],
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

  function permissionLabel(code) {
    const key = String(code || "").trim();
    return PERMISSION_LABELS[key] || key;
  }

  function roleSaveErrorMessage(err) {
    const msg = String(err?.message || err || "");
    if (msg.includes("position capacity reached")) {
      return "В выбранной роли уже заняты все места. Увеличьте лимит роли или выберите другую роль.";
    }
    if (msg.includes("forbidden")) {
      return "У вас нет права менять роли доступа. Нужно право member.access.manage.";
    }
    if (msg.includes("cannot modify system-managed access")) {
      return "Системную роль тимлида нельзя менять через эту модалку.";
    }
    if (msg.includes("only one managed role")) {
      return "У участника может быть только одна роль доступа.";
    }
    if (msg.includes("reserved position_id")) {
      return "TEAM_LEAD - системная роль, ее нельзя назначить как обычную роль проекта.";
    }
    if (msg.includes("unknown position_id")) {
      return "Выбранная роль проекта больше не найдена. Обновите страницу и попробуйте еще раз.";
    }
    return msg || "Не удалось сохранить роль.";
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

  function profileURL(userID) {
    const id = String(userID || "").trim();
    return id ? `/dev/profile?user_id=${encodeURIComponent(id)}` : "/dev/profile";
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
    return i18n ? i18n.formatDateTime(d) : d.toLocaleString();
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

  function taskDisplayLimitKey() {
    return `${LS_TASK_DISPLAY_LIMIT_PREFIX}${state.projectID}`;
  }

  function normalizeTaskDisplayLimit(value) {
    const parsed = Number.parseInt(String(value || ""), 10);
    return TASK_DISPLAY_LIMITS.includes(parsed) ? parsed : DEFAULT_TASK_DISPLAY_LIMIT;
  }

  function loadTaskDisplayLimit() {
    if (!state.projectID) return DEFAULT_TASK_DISPLAY_LIMIT;
    try {
      return normalizeTaskDisplayLimit(localStorage.getItem(taskDisplayLimitKey()));
    } catch (_) {
      return DEFAULT_TASK_DISPLAY_LIMIT;
    }
  }

  function setTaskDisplayLimit(limit) {
    const next = normalizeTaskDisplayLimit(limit);
    state.taskDisplayLimit = next;
    if (!state.projectID) return;
    try {
      localStorage.setItem(taskDisplayLimitKey(), String(next));
    } catch (_) {}
  }

  function hasProjectPermission(code) {
    const wanted = String(code || "").trim();
    return Boolean(wanted) && Array.isArray(state.myPermissions) && state.myPermissions.includes(wanted);
  }

  function canManageAccess() {
    return hasProjectPermission("member.access.manage");
  }

  function canApproveProjectLaunch() {
    return hasProjectPermission("project.approve");
  }

  function canInviteProfessorToProject() {
    return hasProjectPermission("project.invite_professor");
  }

  function canCreateTasksInProject() {
    return canViewWorkspace() && hasProjectPermission("task.create");
  }

  function canAssignTasksInProject() {
    return canViewWorkspace() && hasProjectPermission("task.assign");
  }

  function canUpdateTasksInProject() {
    return canViewWorkspace() && hasProjectPermission("task.update");
  }

  function canDeleteTasksInProject() {
    return canViewWorkspace() && hasProjectPermission("task.delete");
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

  async function request(method, url, body, extra = {}) {
    const { resp, data } = await auth.requestJSON(url, {
      method,
      headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      ...extra,
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
      return await request(method, url, undefined, { skipAccessAlert: true });
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

  function syncStudentSidebar(profile) {
    const section = currentStudentSection();
    localStorage.setItem(LS_STUDENT_SECTION, section);

    const host = document.querySelector("[data-role-sidebar]");
    if (host && roleSidebar && typeof roleSidebar.renderSidebar === "function") {
      host.dataset.sidebarActive = section;
      roleSidebar.renderSidebar(host, {
        role: "student",
        active: section,
        profile,
        scope: typeof auth.getDefaultScope === "function" ? auth.getDefaultScope() : null,
      });

      const logoutBtn = document.getElementById("logoutBtn");
      if (logoutBtn && logoutBtn.dataset.bound !== "1") {
        logoutBtn.dataset.bound = "1";
        logoutBtn.addEventListener("click", () => {
          auth.logout();
        });
      }
    }

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

  function normalizedLifecycleStatus(status) {
    const code = String(status || "").toUpperCase();
    if (code === "REVIEW") return "DRAFT";
    if (code === "ARCHIVE") return "COMPLETED";
    return code || "DRAFT";
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

  function stageHintsStorageKey() {
    return `${LS_STAGE_HINTS_HIDDEN_PREFIX}${state.projectID || projectIDFromPath() || "default"}`;
  }

  function areStageHintsHidden() {
    try {
      const stored = localStorage.getItem(stageHintsStorageKey());
      if (stored === null) return false;
      return stored === "1";
    } catch (_) {
      return false;
    }
  }

  function applyStageHintsVisibility() {
    const hidden = areStageHintsHidden();
    if (ui.stageSpotlight) {
      ui.stageSpotlight.hidden = hidden;
    }
    if (ui.stageHintsCollapsed) {
      ui.stageHintsCollapsed.hidden = !hidden;
    }
    if (ui.hideStageHintsBtn) {
      ui.hideStageHintsBtn.hidden = hidden;
      ui.hideStageHintsBtn.setAttribute("aria-expanded", hidden ? "false" : "true");
    }
    if (ui.showStageHintsBtn) {
      ui.showStageHintsBtn.hidden = !hidden;
      ui.showStageHintsBtn.setAttribute("aria-expanded", hidden ? "false" : "true");
    }
  }

  function setStageHintsHidden(hidden) {
    try {
      localStorage.setItem(stageHintsStorageKey(), hidden ? "1" : "0");
    } catch (_) {}
    applyStageHintsVisibility();
    renderTabHints();
  }

  function revealElement(el) {
    if (!el) return;
    el.scrollIntoView({ behavior: "smooth", block: "start" });
    if (typeof el.focus === "function") {
      try {
        el.focus({ preventScroll: true });
      } catch (_) {}
    }
  }

  function switchToViewAndReveal(viewName, target) {
    setView(viewName);
    requestAnimationFrame(() => {
      if (!target) return;
      revealElement(target);
    });
  }

  function runStageAction(actionName) {
    const action = String(actionName || "").trim();
    if (!action) return;

    if (action === "edit-project") {
      switchToViewAndReveal("edit", ui.viewEdit || ui.aboutCard);
      if (ui.editTitleInput && !ui.viewEdit.hidden) {
        requestAnimationFrame(() => revealElement(ui.editTitleInput));
      }
      return;
    }
    if (action === "overview-about") {
      switchToViewAndReveal("overview", ui.aboutCard);
      return;
    }
    if (action === "open-team") {
      switchToViewAndReveal("team", ui.teamTableBody);
      return;
    }
    if (action === "manage-professor") {
      switchToViewAndReveal("overview", ui.pipelineCard || ui.professorIdentity || ui.professorSearchInput);
      if (ui.professorSearchInput && !ui.professorSearchInput.disabled) {
        requestAnimationFrame(() => revealElement(ui.professorSearchInput));
      }
      return;
    }
    if (action === "open-criteria") {
      switchToViewAndReveal("criteria", ui.viewCriteria);
      return;
    }
    if (action === "open-tasks") {
      switchToViewAndReveal("tasks", ui.viewTasks);
      return;
    }
    if (action === "open-review") {
      switchToViewAndReveal("review", ui.viewReview);
      return;
    }
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
      can_view_project_details: false,
      can_apply: false,
      can_view_final_grade: false,
    };
  }

  function canViewWorkspace() {
    return Boolean(viewerAccess().can_view_workspace);
  }

  function canViewProjectDetails() {
    return canViewWorkspace() || Boolean(viewerAccess().can_view_project_details);
  }

  function canApplyToProject() {
    return Boolean(viewerAccess().can_apply);
  }

  function canViewFinalGrade() {
    return Boolean(viewerAccess().can_view_final_grade);
  }

  function finalReportURL(disposition) {
    if (!state.projectID) return "";
    const url = new URL(`/v2/projects/${encodeURIComponent(state.projectID)}/final-report.pdf`, window.location.origin);
    const lang = i18n && typeof i18n.getLanguage === "function"
      ? String(i18n.getLanguage() || "").trim().toLowerCase()
      : "";
    if (lang) {
      url.searchParams.set("lang", lang);
    }
    if (String(disposition || "").trim().toLowerCase() === "inline") {
      url.searchParams.set("disposition", "inline");
    }
    return url.toString();
  }

  function taskLimitI18n() {
    const lang = i18n && typeof i18n.getLanguage === "function"
      ? String(i18n.getLanguage() || "").trim().toLowerCase()
      : "kk";
    switch (lang) {
      case "en":
        return {
          switchLabel: "Task card limit",
          buttonLabel: (value) => `Show up to ${value} tasks per column`,
          summary: (shown, total) => `Showing ${shown} of ${total} tasks in this column. You can raise the limit above.`,
        };
      case "ru":
        return {
          switchLabel: "Лимит карточек задач",
          buttonLabel: (value) => `Показывать до ${value} задач в колонке`,
          summary: (shown, total) => `Показано ${shown} из ${total} задач в этой колонке. Лимит можно увеличить выше.`,
        };
      default:
        return {
          switchLabel: "Тапсырма карточкаларының лимиті",
          buttonLabel: (value) => `Бағанда ${value} тапсырмаға дейін көрсету`,
          summary: (shown, total) => `Бұл бағанда ${total}-тың ${shown}-ы ғана көрсетілді. Қажет болса, жоғарыдағы лимитті үлкейтіңіз.`,
        };
    }
  }

  function renderTaskDisplayLimitControl() {
    if (!ui.taskDisplayLimitSwitch || !ui.taskDisplayLimitButtons.length) return;
    const copy = taskLimitI18n();
    ui.taskDisplayLimitSwitch.setAttribute("aria-label", copy.switchLabel);
    ui.taskDisplayLimitSwitch.title = copy.switchLabel;
    ui.taskDisplayLimitButtons.forEach((btn) => {
      const limit = normalizeTaskDisplayLimit(btn.dataset.taskLimit);
      const active = limit === state.taskDisplayLimit;
      btn.classList.toggle("active", active);
      btn.setAttribute("aria-pressed", active ? "true" : "false");
      btn.setAttribute("aria-label", copy.buttonLabel(limit));
      btn.title = copy.buttonLabel(limit);
    });
  }

  function allowedViews() {
    if (canViewWorkspace()) {
      return ["overview", "team", "invite", "tasks", "criteria", "review", "edit"];
    }
    if (canViewProjectDetails()) {
      return ["overview", "team", "tasks", "criteria", "review"];
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
    if (s === "REVIEW") return { label: "Подготовка", cls: "muted" };
    if (s === "GRADING") return { label: "Оценивание", cls: "review" };
    if (s === "RECRUITMENT") return { label: "Набор", cls: "muted" };
    if (s === "COMPLETED") return { label: "Завершен", cls: "active" };
    if (s === "ARCHIVE") return { label: "Завершен", cls: "active" };
    return { label: "Черновик", cls: "muted" };
  }

  function lifecycleStageLabel(status) {
    const s = String(status || "").toUpperCase();
    if (s === "REVIEW") return "Подготовка";
    if (s === "RECRUITMENT") return "Набор";
    if (s === "ACTIVE") return "В работе";
    if (s === "GRADING") return "Оценивание";
    if (s === "COMPLETED") return "Завершен";
    if (s === "ARCHIVE") return "Завершен";
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
    if (!canViewProjectDetails()) {
      return [];
    }
    const leadID = String(state.project?.created_by || "").trim();
    const members = (Array.isArray(state.members) ? state.members : []).map((member) => {
      if (String(member?.user_id || "") !== leadID) {
        return member;
      }
      if (String(member?.position_name || "").trim() || String(member?.position_code || "").trim()) {
        return member;
      }
      return {
        ...member,
        position_code: "TEAM_LEAD",
        position_name: "Тимлид",
      };
    });
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
    if (member.access_role_name) return member.access_role_name;
    if (member.access_role_code) return member.access_role_code;
    if (member.position_name) return member.position_name;
    if (member.position_code) return member.position_code;
    const status = String(member?.status || "").toUpperCase();
    if (status === "ACTIVE") return "Участник";
    if (status === "INVITED") return "Приглашён";
    if (status === "APPLIED") return "Заявка";
    return "Без роли";
  }

  function memberLifecycleRoles(member) {
    const roles = [];
    const userID = String(member?.user_id || "");
    const status = String(member?.status || "").toUpperCase();
    if (userID && userID === String(state.project?.created_by || "")) {
      roles.push("TEAM_LEAD");
    } else if (status === "ACTIVE") {
      roles.push("MEMBER");
    } else if (status === "INVITED") {
      roles.push("INVITED_MEMBER");
    } else if (status === "APPLIED") {
      roles.push("APPLIED");
    }
    return roles;
  }

  function activeMembers() {
    return allMembers().filter((m) => String(m.status || "").toUpperCase() === "ACTIVE");
  }

  function normalizePositionCode(code) {
    return String(code || "").trim().toUpperCase();
  }

  function isSystemTaskPositionCode(code) {
    return SYSTEM_TASK_POSITION_CODES.has(normalizePositionCode(code));
  }

  function projectPositionByID(positionID) {
    const id = String(positionID || "");
    return (Array.isArray(state.positions) ? state.positions : []).find((p) => String(p.id || "") === id) || null;
  }

  function isTeamLeadTaskPositionID(positionID) {
    const position = projectPositionByID(positionID);
    return Boolean(position) && isSystemTaskPositionCode(position.code);
  }

  function memberAssignablePositions() {
    return (Array.isArray(state.positions) ? state.positions : []).filter((p) => !isSystemTaskPositionCode(p.code));
  }

  function teamLeadMembers() {
    const leadID = String(state.project?.created_by || "");
    return activeMembers().filter((m) => {
      const memberID = String(m.user_id || "");
      return Boolean(memberID) && (memberID === leadID || normalizePositionCode(m.position_code) === "TEAM_LEAD");
    });
  }

  function membersByPosition(positionID) {
    if (isTeamLeadTaskPositionID(positionID)) {
      return teamLeadMembers();
    }
    return activeMembers().filter((m) => String(m.position_id || "") === String(positionID));
  }

  function positionOccupancy(positionID, excludeUserID = "") {
    const excluded = String(excludeUserID || "");
    const members = activeMembers().filter((m) => String(m.position_id || "") === String(positionID || ""));
    return {
      total: members.length,
      withoutTarget: members.filter((m) => String(m.user_id || "") !== excluded).length,
    };
  }

  function positionCapacity(position) {
    const raw = Number(position?.capacity || 0);
    return Number.isFinite(raw) && raw > 0 ? raw : 1;
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

  function taskDueDate(task) {
    if (!task || !task.due_at) return null;
    const due = new Date(task.due_at);
    return Number.isNaN(due.getTime()) ? null : due;
  }

  function isTaskDone(task) {
    return String(task?.status || "").toUpperCase() === "DONE";
  }

  function isTaskOverdue(task) {
    const due = taskDueDate(task);
    return Boolean(due) && !isTaskDone(task) && due.getTime() < Date.now();
  }

  function overdueTasksCount() {
    return Array.isArray(state.tasks)
      ? state.tasks.filter((task) => isTaskOverdue(task)).length
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
      overdueTasks: overdueTasksCount(),
      gradedCriteria: Math.max(gradedCriteriaCount(), Number(publicSummary?.total || 0)),
    };
  }

  function lifecycleSummaryText(snapshot) {
    if (snapshot.statusCode === "REVIEW") {
      return snapshot.canActivate
        ? "Проект прошел подготовку и готов к запуску."
        : "Проект находится на финальной подготовке перед запуском команды.";
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
        blockers.push("дождаться подтверждения преподавателя");
      }
      if (snapshot.criteriaCount === 0) {
        blockers.push("добавить критерии");
      }
      return blockers.length > 0
        ? `До запуска осталось: ${blockers.join(", ")}.`
        : "Команда и критерии готовы: проект можно переводить в активную фазу.";
    }

    if (snapshot.statusCode === "ACTIVE") {
      if (snapshot.tasksTotal === 0) {
        return "Проект запущен. Следующий шаг: создать и выполнить задачи перед отправкой на оценивание.";
      }
      if (!snapshot.professorAccepted) {
        return `Проект в работе. Завершено ${snapshot.tasksDone}/${snapshot.tasksTotal} задач, но преподаватель еще не подтвердил участие.${snapshot.overdueTasks > 0 ? ` Просрочено: ${snapshot.overdueTasks}.` : ""}`;
      }
      if (snapshot.tasksDone < snapshot.tasksTotal) {
        return `Проект в работе. Закройте все задачи: сейчас выполнено ${snapshot.tasksDone}/${snapshot.tasksTotal}.${snapshot.overdueTasks > 0 ? ` Просрочено: ${snapshot.overdueTasks}.` : ""}`;
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
      return "Все критерии оценены: можно публиковать итоговую оценку и завершать проект.";
    }

    if (snapshot.statusCode === "COMPLETED") {
      return "Проект завершен: итоговая оценка опубликована и доступна в карточке проекта.";
    }

    if (snapshot.statusCode === "ARCHIVE") {
      return "Проект завершен: итоговая оценка опубликована и доступна в карточке проекта.";
    }

    return "Стартовая стадия: заполните описание, роли и стек, затем откройте набор проекта.";
  }

  function lifecycleStepState(stepCode, currentStatus) {
    const code = String(stepCode || "").toUpperCase();
    const status = normalizedLifecycleStatus(currentStatus);

    const stepIndex = PRIMARY_LIFECYCLE_FLOW.indexOf(code);
    const currentIndex = PRIMARY_LIFECYCLE_FLOW.indexOf(status);

    if (stepIndex === -1) return "is-upcoming";
    if (currentIndex === -1) return stepIndex === 0 ? "is-current" : "is-upcoming";
    if (stepIndex < currentIndex) return "is-complete";
    if (stepIndex === currentIndex) {
      if (code === "COMPLETED" && status === "COMPLETED") return "is-completed-current";
      return "is-current";
    }
    return "is-upcoming";
  }

  function lifecycleStateLabel(stepState, optional) {
    if (stepState === "is-completed-current") return "Завершен";
    if (stepState === "is-current") return "Сейчас";
    if (stepState === "is-complete") return "Пройдено";
    if (optional) return "Опционально";
    return "Впереди";
  }

  function lifecycleStepTitle(code) {
    const value = String(code || "").toUpperCase();
    if (value === "RECRUITMENT") return "Набор команды";
    if (value === "ACTIVE") return "Работа";
    if (value === "GRADING") return "Оценивание";
    if (value === "COMPLETED") return "Завершение";
    return "Подготовка";
  }

  function stageGuideData() {
    const snapshot = lifecycleSnapshot();
    const status = normalizedLifecycleStatus(snapshot.statusCode);
    const descriptionText = String(getReadmeText() || state.project?.description || "").trim();
    const hasDescription = descriptionText.length >= 20;
    const hasRoles = snapshot.requiredMembers > 0;
    const hasCriteria = snapshot.criteriaCount > 0;
    const tasksTotal = snapshot.tasksTotal;
    const tasksDone = snapshot.tasksDone;
    const allTasksDone = tasksTotal > 0 && tasksDone === tasksTotal;
    const canOpenEdit = canViewWorkspace() && isCurrentUserActiveMember() && !canApplyToProject();

    if (status === "RECRUITMENT") {
      return {
        tone: "recruitment",
        currentLabel: lifecycleStepTitle("RECRUITMENT"),
        nextLabel: lifecycleStepTitle("ACTIVE"),
        title: "Что нужно, чтобы перейти в работу",
        copy: "Видны только требования ближайшего перехода. Выполненные пункты сразу зачеркиваются.",
        items: [
          {
            label: "Набрать команду по всем ролям",
            hint: `Сейчас занято ${snapshot.activeMembers} из ${snapshot.requiredMembers || 0} мест.`,
            done: snapshot.requiredMembers > 0 && snapshot.activeMembers >= snapshot.requiredMembers,
            action: { name: "open-team", label: "Открыть команду" },
          },
          {
            label: "Подтвердить преподавателя",
            hint: snapshot.hasProfessor ? `Статус приглашения: ${snapshot.professorLabel}.` : "Назначьте преподавателя на проект.",
            done: snapshot.professorAccepted,
            action: canViewWorkspace() ? { name: "manage-professor", label: "К преподавателю" } : null,
          },
          {
            label: "Преподаватель должен добавить критерии оценки",
            hint: hasCriteria
              ? `Преподаватель уже настроил ${snapshot.criteriaCount} критериев.`
              : "Критерии добавляет преподаватель на вкладке «Критерии». Команда здесь только сверяется с ними.",
            done: hasCriteria,
            action: { name: "open-criteria", label: "Открыть критерии" },
          },
        ],
      };
    }

    if (status === "ACTIVE") {
      return {
        tone: "active",
        currentLabel: lifecycleStepTitle("ACTIVE"),
        nextLabel: lifecycleStepTitle("GRADING"),
        title: "Что нужно, чтобы отправить проект на оценивание",
        copy: "Команда видит только то, что блокирует ближайшую сдачу проекта.",
        items: [
          {
            label: "Подтверждение преподавателя получено",
            hint: snapshot.hasProfessor ? `Текущий статус: ${snapshot.professorLabel}.` : "Сначала назначьте преподавателя.",
            done: snapshot.professorAccepted,
            action: canViewWorkspace() ? { name: "manage-professor", label: "Проверить преподавателя" } : null,
          },
          {
            label: "Создать хотя бы одну задачу",
            hint: tasksTotal > 0
              ? `В проекте уже ${tasksTotal} задач.`
              : "Первую задачу обычно создает тимлид или task manager, иначе проект нельзя отправить на оценивание.",
            done: tasksTotal > 0,
            action: { name: "open-tasks", label: "Открыть задачи" },
          },
          {
            label: "Закрыть все задачи",
            hint: tasksTotal > 0
              ? `Готово ${tasksDone} из ${tasksTotal}.`
              : "Сначала добавьте задачи в канбан. Это делает тимлид или task manager.",
            done: allTasksDone,
            action: { name: "open-tasks", label: "Перейти в канбан" },
          },
        ],
      };
    }

    if (status === "GRADING") {
      return {
        tone: "grading",
        currentLabel: lifecycleStepTitle("GRADING"),
        nextLabel: lifecycleStepTitle("COMPLETED"),
        title: "Что нужно, чтобы завершить проект",
        copy: "На этом этапе команда видит только ближайшие требования до публикации итоговой оценки.",
        items: [
          {
            label: "Критерии оценки настроены",
            hint: `Всего критериев: ${snapshot.criteriaCount}.`,
            done: hasCriteria,
            action: { name: "open-criteria", label: "Открыть критерии" },
          },
          {
            label: "Оценки выставлены по всем критериям",
            hint: `Сейчас проверено ${snapshot.gradedCriteria} из ${snapshot.criteriaCount}.`,
            done: hasCriteria && snapshot.gradedCriteria >= snapshot.criteriaCount,
            action: { name: "open-review", label: "Открыть ревью" },
          },
        ],
      };
    }

    if (status === "COMPLETED") {
      return {
        tone: "completed",
        currentLabel: lifecycleStepTitle("COMPLETED"),
        nextLabel: "Финал достигнут",
        title: "Проект завершен",
        copy: "Все обязательные шаги выполнены. Здесь остается только итоговый статус проекта.",
        items: [
          {
            label: "Итоговая оценка опубликована",
            hint: "Карточка проекта зафиксирована как завершенный кейс.",
            done: true,
          },
        ],
      };
    }

    return {
      tone: "draft",
      currentLabel: lifecycleStepTitle("DRAFT"),
      nextLabel: lifecycleStepTitle("RECRUITMENT"),
      title: "Что подготовить перед открытием набора",
      copy: "Блок показывает только ближайший шаг, чтобы команде было понятнее, что делать прямо сейчас.",
      items: [
        {
          label: "Заполнить описание или README проекта",
          hint: hasDescription ? "Базовое описание проекта уже есть." : "Кратко опишите идею, стек и ожидаемый результат.",
          done: hasDescription,
          action: { name: canOpenEdit ? "edit-project" : "overview-about", label: canOpenEdit ? "Открыть редактирование" : "Открыть описание" },
        },
        {
          label: "Добавить хотя бы одну роль в команду",
          hint: hasRoles ? `Сейчас в проекте ${snapshot.requiredMembers} ролей.` : "Создайте роли, чтобы открыть набор осознанно.",
          done: hasRoles,
          action: { name: "open-team", label: "Открыть команду" },
        },
      ],
    };
  }

  function stageHintCopy(guide) {
    const items = Array.isArray(guide?.items) ? guide.items : [];
    if (!items.length) {
      return guide?.copy || "";
    }

    const next = guide?.nextLabel || "следующий этап";
    const pending = items.filter((item) => item && !item.done);
    if (!pending.length) {
      return `Все требования для перехода на этап «${next}» выполнены. Готовые пункты ниже отмечены и вычеркнуты.`;
    }

    const pendingLabels = pending
      .slice(0, 3)
      .map((item) => String(item.label || "").trim().toLowerCase())
      .filter(Boolean)
      .join(", ");
    if (!pendingLabels) {
      return guide?.copy || "";
    }

    return `Чтобы перейти на этап «${next}», нужно: ${pendingLabels}. Готовые требования ниже отмечаются и вычеркиваются.`;
  }

  function renderStageChecklist(items) {
    if (!ui.readinessList) return;
    const list = Array.isArray(items) ? items : [];
    if (!list.length) {
      ui.readinessList.innerHTML = '<div class="empty-state">Ближайшие требования пока не определены.</div>';
      return;
    }

    const doneCount = list.filter((item) => item && item.done).length;
    const listHTML = list
      .map((item, idx) => {
        const id = `stage-check-${idx}`;
        const action = item && typeof item.action === "object" ? item.action : null;
        const actionHTML = !item.done && action && action.name && action.label
          ? (
            `<div class="stage-checklist-actions">` +
              `<button class="stage-checklist-action" type="button" data-stage-action="${escapeHTML(action.name)}">${escapeHTML(action.label)}</button>` +
            `</div>`
          )
          : "";
        return (
          `<div class="stage-checklist-item ${item.done ? "is-done" : "is-pending"}">` +
            `<input id="${id}" type="checkbox" ${item.done ? "checked" : ""} disabled tabindex="-1" aria-hidden="true" />` +
            `<label class="stage-checklist-label" for="${id}">` +
              `<span class="stage-checklist-title">${escapeHTML(item.label)}</span>` +
              `<small class="stage-checklist-hint">${escapeHTML(item.hint || "")}</small>` +
            `</label>` +
            actionHTML +
          `</div>`
        );
      })
      .join("");

    ui.readinessList.innerHTML =
      `<div class="stage-checklist-head">` +
        `<span>Требования перехода</span>` +
        `<strong>${doneCount}/${list.length} готово</strong>` +
      `</div>` +
      `<div class="stage-checklist-items">${listHTML}</div>`;
  }

  function tabStageHints() {
    const snapshot = lifecycleSnapshot();
    const status = normalizedLifecycleStatus(snapshot.statusCode);
    const tasksRemaining = Math.max(0, snapshot.tasksTotal - snapshot.tasksDone);
    const criteriaReady = snapshot.criteriaCount > 0;

    const team = (() => {
      if (status === "RECRUITMENT") {
        return {
          tone: "recruitment",
          title: "Сейчас главное закрыть все роли",
          copy: `Активных участников ${snapshot.activeMembers} из ${snapshot.requiredMembers || 0}. Составом обычно управляют тимлид, ко-лид и рекрутер. Когда каждое место будет занято и преподаватель подтвердит участие, проект можно будет запускать.`,
        };
      }
      if (status === "ACTIVE") {
        return {
          tone: "active",
          title: "Команда уже в рабочем составе",
          copy: "Проверьте, что у каждого участника есть своя роль и доступы. Если нужно, здесь же можно управлять составом и приглашениями.",
        };
      }
      if (status === "GRADING") {
        return {
          tone: "grading",
          title: "Состав команды должен быть понятен для оценки",
          copy: "Во время оценивания обычно уже не меняют участников. Здесь важно, чтобы роли и вклад команды были отражены корректно.",
        };
      }
      if (status === "COMPLETED") {
        return {
          tone: "completed",
          title: "Это финальный состав команды",
          copy: "Именно этот состав будет виден в завершенном кейсе проекта и в истории работы.",
        };
      }
      return {
        tone: "draft",
        title: "Сначала подготовьте каркас команды",
        copy: "Добавьте роли и количество мест заранее. Так при открытии набора студентам будет сразу понятно, кого именно вы ищете.",
      };
    })();

    const tasks = (() => {
      if (status === "ACTIVE") {
        if (snapshot.tasksTotal === 0) {
          return {
            tone: "active",
            title: "Создайте первую задачу",
            copy: "Без задач проект нельзя отправить на оценивание. Первые карточки обычно создает тимлид или task manager, а затем распределяет их по ролям.",
          };
        }
        if (tasksRemaining > 0) {
          return {
            tone: "active",
            title: "Доведите канбан до конца",
            copy: `Сейчас осталось закрыть ${tasksRemaining} задач. Перед отправкой на оценивание все карточки должны быть в колонке "Готово". Лишние задачи при необходимости может удалить тимлид или task manager.`,
          };
        }
        return {
          tone: "active",
          title: "Все задачи уже закрыты",
          copy: "Канбан готов. Если преподаватель уже подтвержден, проект можно отправлять на оценивание.",
        };
      }
      if (status === "GRADING") {
        return {
          tone: "grading",
          title: "Задачи здесь работают как история выполнения",
          copy: "На этапе оценивания канбан нужен как подтверждение сделанной работы. Новые задачи обычно уже не добавляют.",
        };
      }
      if (status === "COMPLETED") {
        return {
          tone: "completed",
          title: "Канбан зафиксирован как история проекта",
          copy: "Здесь можно посмотреть, как команда прошла путь от первых задач до завершения кейса.",
        };
      }
      return {
        tone: "draft",
        title: "До запуска задачи можно не расписывать подробно",
        copy: "Сначала соберите команду, преподавателя и критерии. Полноценный канбан удобнее вести, когда проект уже перешел в рабочую фазу.",
      };
    })();

    const criteria = (() => {
      if (status === "GRADING") {
        return {
          tone: "grading",
          title: "По этим критериям идет финальная оценка",
          copy: `Сейчас оценено ${snapshot.gradedCriteria} из ${snapshot.criteriaCount}. Критерии добавляет и заполняет преподаватель, а команда здесь видит итоговое состояние проверки.`,
        };
      }
      if (status === "ACTIVE") {
        return {
          tone: criteriaReady ? "active" : "draft",
          title: criteriaReady ? "Команда должна сверяться с критериями" : "Критерии еще не готовы",
          copy: criteriaReady
            ? `Сейчас в проекте ${snapshot.criteriaCount} критериев. Их добавляет преподаватель, а команда использует этот список как ориентир перед сдачей проекта.`
            : "Без критериев проект упрется в блокер. Их должен добавить преподаватель до момента отправки проекта на проверку.",
        };
      }
      if (status === "COMPLETED") {
        return {
          tone: "completed",
          title: "Критерии уже отработали свою роль",
          copy: "Здесь остается финальный набор требований, по которым преподаватель оценивал проект.",
        };
      }
      return {
        tone: criteriaReady ? "recruitment" : "draft",
        title: criteriaReady ? "Критерии уже подготовлены" : "Подготовьте критерии заранее",
        copy: criteriaReady
          ? `В проекте уже есть ${snapshot.criteriaCount} критериев. Их подготовил преподаватель, поэтому команда заранее понимает ожидания.`
          : "Без критериев проект не сможет перейти дальше. Их заранее добавляет преподаватель, а студенты здесь только видят готовый список.",
      };
    })();

    return { team, tasks, criteria };
  }

  function renderTabStageHint(target, data) {
    if (!target || !data) return;
    const isCompact = target.classList.contains("tab-stage-hint--compact");
    const hidden = areStageHintsHidden();
    if (hidden) {
      target.className = `tab-stage-hint tab-stage-hint--restore${isCompact ? " tab-stage-hint--compact" : ""}`;
      target.innerHTML =
        `<div class="tab-stage-hint-head">` +
          `<strong>Подсказки по этапам скрыты</strong>` +
          `<button class="hint-toggle-btn hint-toggle-btn--tab" type="button" data-stage-hints-visibility="show">Открыть</button>` +
        `</div>` +
        `<p>Откройте их в любой момент, чтобы быстро свериться со следующим этапом проекта.</p>`;
      return;
    }
    target.className = `tab-stage-hint tab-stage-hint--${data.tone || "draft"}${isCompact ? " tab-stage-hint--compact" : ""}`;
    target.innerHTML =
      `<div class="tab-stage-hint-head">` +
        HINT_BADGE_HTML +
        `<button class="hint-toggle-btn hint-toggle-btn--tab" type="button" data-stage-hints-visibility="hide">Скрыть</button>` +
      `</div>` +
      `<strong>${escapeHTML(data.title || "")}</strong>` +
      `<p>${escapeHTML(data.copy || "")}</p>`;
  }

  function renderTabHints() {
    const hints = tabStageHints();
    renderTabStageHint(ui.teamStageHint, hints.team);
    renderTabStageHint(ui.tasksStageHint, hints.tasks);
    renderTabStageHint(ui.criteriaStageHint, hints.criteria);
  }

  function pipelineAlertHTML(message) {
    return (
      `<span class="pipeline-status-note__icon" aria-hidden="true">${ALERT_ICON_HTML}</span>` +
      `<span class="pipeline-status-note__copy">${escapeHTML(message || "")}</span>`
    );
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

  function projectPositionOptions(selectedID, options = {}) {
    const includeSystemTaskPositions = Boolean(options.includeSystemTaskPositions);
    const placeholder = Object.prototype.hasOwnProperty.call(options, "placeholder")
      ? String(options.placeholder || "")
      : "Выберите роль";
    const positions = includeSystemTaskPositions
      ? (Array.isArray(state.positions) ? state.positions : [])
      : memberAssignablePositions();

    if (positions.length === 0) {
      return '<option value="">Нет ролей</option>';
    }

    const html = [];
    if (placeholder) {
      html.push(`<option value="">${escapeHTML(placeholder)}</option>`);
    }
    positions.forEach((p) => {
      const selected = String(p.id) === String(selectedID) ? "selected" : "";
      html.push(`<option value="${escapeHTML(p.id)}" ${selected}>${escapeHTML(p.name)} (${escapeHTML(p.code)})</option>`);
    });
    return html.join("");
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
    if (Array.isArray(meta.tags)) {
      meta.tags.slice(0, 2).forEach((t) => tags.push(String(t).toUpperCase()));
    }
    return tags;
  }

  function taskPriorityMeta(task) {
    const meta = state.taskMeta[task.id] || {};
    const code = String(meta.priority || "").trim().toUpperCase();
    if (!code) return null;
    if (code === "HIGH") return { code, tone: "high", label: "Сложная" };
    if (code === "LOW") return { code, tone: "low", label: "Легкая" };
    return { code: "MEDIUM", tone: "medium", label: "Средняя" };
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

  function criterionWeightValue(criterion) {
    const weight = Number(criterion && criterion.weight ? criterion.weight : 0);
    return weight > 0 ? weight : 1;
  }

  function projectRetakeCount() {
    const count = Number(state.project && state.project.retake_count ? state.project.retake_count : 0);
    return Number.isFinite(count) && count > 0 ? Math.round(count) : 0;
  }

  function projectRetakePenaltyPercent() {
    const explicit = Number(state.project && state.project.retake_penalty_percent ? state.project.retake_penalty_percent : 0);
    return Number.isFinite(explicit) && explicit >= 0 ? Math.round(explicit) : 0;
  }

  function reviewSummaryData() {
    const criteria = Array.isArray(state.criteria) ? state.criteria : [];
    const grading = gradingByCriterion();
    const publicSummary = state.project && typeof state.project.review_summary === "object"
      ? state.project.review_summary
      : null;
    const retakeCount = projectRetakeCount();
    const penaltyPercent = projectRetakePenaltyPercent();

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
        retakeCount,
        penaltyPercent,
      };
    }

    let met = 0;
    let missed = 0;
    let reviewed = 0;
    let weightTotal = 0;
    let weightMet = 0;
    let latest = null;
    const comments = [];

    criteria.forEach((criterion) => {
      const weight = criterionWeightValue(criterion);
      weightTotal += weight;
      const id = String(criterion.id || "");
      const item = grading.get(id);
      if (!item) return;
      if (item.isMet === true) {
        met += 1;
        reviewed += 1;
        weightMet += weight;
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
    const rawPassPercent = weightTotal > 0 ? Math.round((weightMet * 100) / weightTotal) : 0;
    const passPercent = Math.max(0, rawPassPercent - penaltyPercent);
    const score = (passPercent * 5 / 100).toFixed(1);
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
      retakeCount,
      penaltyPercent,
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

  // permissionPresetForRole, defaultPermissionState, ensureMemberPermissionState — removed in v1 (backend-driven).

  function bindProfile(profile) {
    const current = profile || auth.getCachedProfile();
    state.profile = current || null;

    const host = document.querySelector("[data-role-sidebar]");
    if (current?.is_professor && host && roleSidebar && typeof roleSidebar.renderSidebar === "function") {
      host.dataset.sidebarRole = "teacher";
      host.dataset.sidebarActive = "projects";
      roleSidebar.renderSidebar(host, {
        role: "teacher",
        active: "projects",
        profile: current,
        scope: typeof auth.getDefaultScope === "function" ? auth.getDefaultScope() : null,
      });

      const logoutBtn = document.getElementById("logoutBtn");
      if (logoutBtn && logoutBtn.dataset.bound !== "1") {
        logoutBtn.dataset.bound = "1";
        logoutBtn.addEventListener("click", () => {
          auth.logout();
        });
      }

      if (ui.crumbSectionLink) {
        ui.crumbSectionLink.textContent = "Проекты";
        ui.crumbSectionLink.href = "/dev/professor#projects";
      }
      return;
    }

    syncStudentSidebar(current);
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
    const canEdit = canViewWorkspace() && isCurrentUserActiveMember() && !applyMode;

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
        title: "Подготовка",
        copy: "Идея проекта, описание, README, стек и базовые роли команды.",
        meta: [visibilityLabel(state.project?.visibility), `Стек ${state.stacks.length}`, "README / описание"],
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
        code: "COMPLETED",
        title: "Завершение",
        copy: "Итоговая оценка опубликована, проект завершен и доступен как завершенный кейс.",
        meta: [snapshot.statusCode === "COMPLETED" || snapshot.statusCode === "ARCHIVE" ? "Итог опубликован" : "Финальная стадия", `Критерии ${snapshot.criteriaCount}`],
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

    const statusCode = projectStatusCode();
    const normalizedStatus = normalizedLifecycleStatus(statusCode);
    const canManagePipeline = canViewWorkspace() && isCurrentUserLead();
    const canGrantLaunch = canViewWorkspace() && canApproveProjectLaunch();
    const isActivationStatus = statusCode !== "ACTIVE" && statusCode !== "GRADING" && statusCode !== "COMPLETED" && statusCode !== "ARCHIVE";
    const canOpenRecruitment = canManagePipeline && normalizedStatus === "DRAFT";
    const snapshot = lifecycleSnapshot();
    const guide = stageGuideData();

    if (ui.stageSpotlight) {
      ui.stageSpotlight.className = `stage-spotlight stage-spotlight--${guide.tone}`;
    }
    if (ui.stageSpotlightTitle) {
      ui.stageSpotlightTitle.textContent = guide.title;
    }
    if (ui.stageSpotlightCopy) {
      ui.stageSpotlightCopy.textContent = stageHintCopy(guide);
    }
    if (ui.stageCurrentBadge) {
      ui.stageCurrentBadge.textContent = `Сейчас: ${guide.currentLabel}`;
    }
    if (ui.stageNextBadge) {
      ui.stageNextBadge.textContent = `Дальше: ${guide.nextLabel}`;
    }
    renderStageChecklist(guide.items);

    if (ui.openRecruitmentBtn) {
      ui.openRecruitmentBtn.hidden = !canOpenRecruitment;
      ui.openRecruitmentBtn.disabled = !canOpenRecruitment;
      ui.openRecruitmentBtn.textContent = "Открыть набор";
    }

    if (ui.pipelineStatusNote) {
      ui.pipelineStatusNote.hidden = true;
      ui.pipelineStatusNote.textContent = "";
      ui.pipelineStatusNote.className = "pipeline-status-note";
    }

    const professorStatusCode = String(
      state.readiness?.professor_status || state.project?.professor_review_status || "NONE"
    ).toUpperCase();

    if (!state.readiness) {
      ui.approveProjectBtn.hidden = !canGrantLaunch;
      ui.approveProjectBtn.disabled = true;
      ui.approveProjectBtn.textContent = "Дать разрешение на запуск";
      ui.approveProjectBtn.title = "";
      if (ui.completeProjectBtn) {
        ui.completeProjectBtn.hidden = true;
        ui.completeProjectBtn.disabled = true;
        ui.completeProjectBtn.title = "";
      }
      applyStageHintsVisibility();
      renderProfessorInviteArea(professorStatusCode);
      return;
    }

    const professorAccepted = professorStatusCode === "ACCEPTED";
    const canShowApprove = canGrantLaunch && isActivationStatus;
    ui.approveProjectBtn.hidden = !canShowApprove;
    ui.approveProjectBtn.textContent = "Дать разрешение на запуск";
    ui.approveProjectBtn.disabled = !state.readiness.can_activate;
    ui.approveProjectBtn.title = !canShowApprove
      ? ""
      : state.readiness.can_activate
        ? "Перевести проект в ACTIVE"
        : "Для запуска должны быть готовы команда, подтвержден преподаватель и настроены критерии.";

    if (ui.completeProjectBtn) {
      const isMember = isCurrentUserActiveMember();
      const tasksTotal = Array.isArray(state.tasks) ? state.tasks.length : 0;
      const tasksDone = Array.isArray(state.tasks)
        ? state.tasks.filter((t) => String(t.status || "").toUpperCase() === "DONE").length
        : 0;
      const allTasksDone = tasksTotal > 0 && tasksDone === tasksTotal;

      const visible = statusCode === "ACTIVE" && isMember;
      ui.completeProjectBtn.hidden = !visible;
      ui.completeProjectBtn.textContent = "Отправить на оценивание";
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
        ui.completeProjectBtn.title = readyForSubmit ? "" : `Нельзя отправить на оценивание: ${reasons.join(", ")}`;
      }
    }

    if (ui.pipelineStatusNote) {
      let note = "";
      let noteClass = "pipeline-status-note pipeline-status-note--info";

      if (normalizedStatus === "DRAFT") {
        if (canOpenRecruitment) {
          note = "Когда базовая структура готова, откройте набор и начните собирать команду.";
        } else if (canViewWorkspace()) {
          note = "Открыть набор может тимлид проекта. Подготовьте описание и роли, чтобы следующий шаг был очевиден для команды.";
        }
      } else if (normalizedStatus === "RECRUITMENT") {
        if (canGrantLaunch && state.readiness.can_activate) {
          note = "Все условия собраны. Преподаватель может дать разрешение и перевести проект в рабочую фазу.";
          noteClass = "pipeline-status-note pipeline-status-note--success";
        } else if (canGrantLaunch) {
          note = "Для запуска еще не хватает обязательных условий. Проверьте чеклист слева и доберите недостающие пункты.";
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        } else if (canManagePipeline && state.readiness.can_activate) {
          note = "Команда готова. Следующий шаг за преподавателем: дать разрешение на запуск.";
          noteClass = "pipeline-status-note pipeline-status-note--success";
        } else if (canManagePipeline) {
          note = "Доведите набор до готовности, и после этого преподаватель сможет запустить проект.";
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        }
      } else if (normalizedStatus === "ACTIVE") {
        if (projectRetakeCount() > 0) {
          note = "Проект возвращен на пересдачу. Закройте замечания преподавателя и отправьте его на оценивание повторно.";
          noteClass = "pipeline-status-note pipeline-status-note--error";
        } else if (snapshot.tasksTotal === 0) {
          note = "Сначала создайте задачи в канбане, иначе проект нельзя будет отправить на оценивание.";
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        } else if (snapshot.tasksDone < snapshot.tasksTotal) {
          note = `До сдачи осталось закрыть ${snapshot.tasksTotal - snapshot.tasksDone} задач.`;
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        } else if (!snapshot.professorAccepted) {
          note = "Все задачи готовы, но нужно дождаться подтверждения преподавателя.";
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        } else {
          note = "Проект готов к передаче на оценивание. Кнопка отправки вынесена рядом.";
          noteClass = "pipeline-status-note pipeline-status-note--success";
        }
      } else if (normalizedStatus === "GRADING") {
        if (snapshot.criteriaCount === 0) {
          note = "Преподавателю нужно сначала добавить критерии, иначе финальная оценка не будет опубликована.";
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        } else if (snapshot.gradedCriteria < snapshot.criteriaCount) {
          note = `Пока проверено ${snapshot.gradedCriteria} из ${snapshot.criteriaCount} критериев.`;
          noteClass = "pipeline-status-note pipeline-status-note--warning";
        } else {
          note = "Все критерии заполнены. Осталось опубликовать итоговую оценку.";
          noteClass = "pipeline-status-note pipeline-status-note--success";
        }
      } else if (normalizedStatus === "COMPLETED") {
        note = "Финальная оценка уже опубликована. Проект завершен.";
        noteClass = "pipeline-status-note pipeline-status-note--success";
      }

      ui.pipelineStatusNote.hidden = !note;
      ui.pipelineStatusNote.className = noteClass;
      ui.pipelineStatusNote.innerHTML = note ? pipelineAlertHTML(note) : "";
    }

    if (ui.stageSpotlight) {
      const hasSideContent = Boolean(
        (ui.pipelineStatusNote && !ui.pipelineStatusNote.hidden) ||
        (ui.openRecruitmentBtn && !ui.openRecruitmentBtn.hidden) ||
        (ui.approveProjectBtn && !ui.approveProjectBtn.hidden) ||
        (ui.completeProjectBtn && !ui.completeProjectBtn.hidden)
      );
      ui.stageSpotlight.classList.toggle("is-full-width", !hasSideContent);
      if (ui.stageSpotlightSide) {
        ui.stageSpotlightSide.hidden = !hasSideContent;
      }
    }

    applyStageHintsVisibility();
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
          `<div><strong><a href="${escapeHTML(profileURL(state.professorSummary.user_id))}">${escapeHTML(fullName)}</a></strong><small>${escapeHTML(state.professorSummary.email || "")}${escapeHTML(dep)}</small></div>`;
      } else {
        ui.professorIdentity.hidden = true;
        ui.professorIdentity.innerHTML = "";
      }
    }

    if (!ui.professorInviteHint) return;

    if (i18n && typeof i18n.clearRich === "function") {
      i18n.clearRich(ui.professorInviteHint);
    }

    if (status === "PENDING") {
      if (isInvitedProfessor) {
        if (i18n && typeof i18n.setRich === "function") {
          i18n.setRich(ui.professorInviteHint, "project.professorInvite.pending");
        } else {
          ui.professorInviteHint.innerHTML = 'У вас есть приглашение на ревью. Откройте страницу <a href="/dev/professor/reviews">/dev/professor/reviews</a> и примите его.';
        }
      } else {
        ui.professorInviteHint.textContent = "Ожидаем подтверждения преподавателя в его кабинете ревью.";
      }
    } else if (status === "ACCEPTED") {
      ui.professorInviteHint.textContent = "Преподаватель подтвердил участие в ревью.";
    } else if (status === "REJECTED") {
      ui.professorInviteHint.textContent = "Преподаватель отклонил приглашение. Выберите другого преподавателя.";
    } else {
      ui.professorInviteHint.textContent = "Преподаватель пока не приглашён.";
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
    const canManageTeam = hasProjectPermission("member.approve");

    filtered.forEach((m) => {
      const status = String(m.status || "").toUpperCase();
      const statusClass = status === "APPLIED" ? "invited" : status.toLowerCase();
      const statusLabel = status === "APPLIED" ? "INVITED" : status;
      const github = `https://github.com/${slugify(getDisplayName(m.user_id))}`;
      const isLeadRow = String(m.user_id) === String(state.project?.created_by || "");
      const roleLabel = isLeadRow ? "Тимлид" : getRoleLabel(m);
      const roleIsEmpty = !isLeadRow && !String(m.position_name || m.position_code || "").trim();
      const canApprove = canManageTeam && status === "APPLIED";
      const canRejectApplication = canManageTeam && status === "APPLIED";
      const canRemoveMember = canManageTeam && !isLeadRow && (status === "ACTIVE" || status === "INVITED");
      const canRespondInvite = status === "INVITED" && String(m.user_id) === currentUser;
      const canOpenRoleModal = status === "ACTIVE" && !isLeadRow && canManageAccess();
      const actions = [];

      if (canApprove) {
        actions.push('<button class="ghost-btn" data-member-action="approve">Одобрить</button>');
      }
      if (canRejectApplication) {
        actions.push('<button class="ghost-btn" data-member-action="reject-application">Отклонить</button>');
      }
      if (canRespondInvite) {
        actions.push('<button class="ghost-btn" data-member-action="accept-invite">Принять</button>');
        actions.push('<button class="ghost-btn" data-member-action="reject-invite">Отклонить</button>');
      }
      if (canOpenRoleModal) {
        actions.push('<button class="ghost-btn" data-member-action="role">Роль</button>');
      }
      if (canRemoveMember) {
        actions.push('<button class="ghost-btn danger-btn" data-member-action="remove">Удалить</button>');
      }

      const row = document.createElement("tr");
      row.setAttribute("data-user-id", m.user_id || "");
      row.setAttribute("data-member-status", status);
      row.innerHTML =
        `<td>` +
          `<div class="user-cell">` +
            `<div class="member-avatar">${escapeHTML(initials(getDisplayName(m.user_id), m.user_id))}</div>` +
            `<div><strong><a href="${escapeHTML(profileURL(m.user_id))}">${escapeHTML(getDisplayName(m.user_id))}</a></strong><small>${escapeHTML(getDisplaySubline(m.user_id))}</small></div>` +
          `</div>` +
        `</td>` +
        `<td><span class="status-badge ${statusClass}">${escapeHTML(statusLabel)}</span></td>` +
        `<td><span class="member-role-pill${roleIsEmpty ? " muted" : ""}">${escapeHTML(roleLabel)}</span></td>` +
        `<td><a class="meta-link" href="${escapeHTML(github)}" target="_blank" rel="noreferrer">${escapeHTML(github.replace("https://", ""))}</a></td>` +
        `<td>` +
          `<div class="task-toolbar">` +
            `${actions.length ? actions.join("") : '<span class="member-action-empty">—</span>'}` +
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
          `<div><strong><a href="${escapeHTML(profileURL(item.user_id))}">${escapeHTML(item.full_name || item.email)}</a></strong><small>${escapeHTML(item.email || "")} · ${escapeHTML(item.department_code || "-")}</small></div>` +
        `</div>` +
        `<input class="invite-comment-input" type="text" placeholder="Комментарий к приглашению" />` +
        `<button class="ghost-btn" data-invite-user="${escapeHTML(item.user_id)}">Пригласить</button>`;
      ui.inviteCandidatesList.appendChild(row);
    });
  }

  function renderProfessorSearchResults() {
    if (!ui.professorSearchResults) return;
    if (!canViewWorkspace()) {
      ui.professorSearchResults.hidden = true;
      ui.professorSearchResults.innerHTML = "";
      return;
    }
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
    const canCreateTasks = canCreateTasksInProject() && isActive;
    const doneCount = state.tasks.filter((t) => String(t.status || "").toUpperCase() === "DONE").length;
    const total = state.tasks.length;
    const percent = total > 0 ? Math.round((doneCount * 100) / total) : 0;
    const overdueCount = overdueTasksCount();

    ui.activeProgressWrap.hidden = !isActive;
    ui.progressBadge.className = `status-pill ${isActive ? "active" : "muted"}`;
    ui.progressBadge.textContent = isActive
      ? `Прогресс открыт · ${percent}%${overdueCount > 0 ? ` · просрочено ${overdueCount}` : ""}`
      : "Прогресс закрыт до ACTIVE";

    if (isActive) {
      ui.progressPercent.textContent = `${percent}%`;
      ui.progressFill.style.width = `${percent}%`;
    }

    ui.openTaskModalBtn.disabled = !canCreateTasks;
    if (!canCreateTasksInProject()) {
      ui.openTaskModalBtn.title = "Создание задач доступно только участникам проекта с правом управления";
    } else {
      ui.openTaskModalBtn.title = isActive ? "" : "Создание задач доступно только после ACTIVE";
    }
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
      tasks_overdue: overdueTasksCount(),
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

  function taskStatusMeta(task) {
    const status = String(task && task.status || "OPEN").toUpperCase();
    if (isTaskOverdue(task)) {
      return { tone: "overdue", label: "Просрочено" };
    }
    if (status === "DONE") {
      return { tone: "done", label: "Готово" };
    }
    if (status === "IN_PROGRESS") {
      return { tone: "in-progress", label: "В работе" };
    }
    return { tone: "open", label: "Открыта" };
  }

  function createTaskCard(task) {
    const card = document.createElement("article");
    const overdue = isTaskOverdue(task);
    const priority = taskPriorityMeta(task);
    const statusMeta = taskStatusMeta(task);
    const cardClasses = ["task-item", `task-item--state-${statusMeta.tone}`];
    card.className = cardClasses.join(" ");
    card.setAttribute("data-task-id", task.id || "");

    const status = String(task.status || "OPEN").toUpperCase();
    const description = String(task.description || "").trim();
    const tags = taskTags(task);
    const canAssignTasks = canAssignTasksInProject();
    const canUpdateTasks = canUpdateTasksInProject();
    const canDeleteTasks = canDeleteTasksInProject();
    const isAssignee = isCurrentUserAssignee(task);
    const dueLabel = task.due_at ? formatDate(task.due_at) : "не задан";
    let controlsHTML = "";

    if (canAssignTasks || canUpdateTasks || canDeleteTasks) {
      controlsHTML = `<div class="task-controls">`;
      if (canUpdateTasks) {
        controlsHTML +=
          `<div class="task-control-row">` +
            `<select data-task-status>` +
              `<option value="OPEN" ${status === "OPEN" ? "selected" : ""}>OPEN</option>` +
              `<option value="IN_PROGRESS" ${status === "IN_PROGRESS" ? "selected" : ""}>IN_PROGRESS</option>` +
              `<option value="DONE" ${status === "DONE" ? "selected" : ""}>DONE</option>` +
            `</select>` +
            `<button class="ghost-btn" data-task-action="status">Статус</button>` +
          `</div>`;
      }
      if (canAssignTasks) {
        controlsHTML +=
          `<div class="task-control-row">` +
            `<select data-task-assignee>${assigneeOptions(task.position_id, task.assignee_user_id || "")}</select>` +
            `<button class="ghost-btn" data-task-action="assign">Назначить</button>` +
          `</div>`;
      }
      if (isAssignee && status === "OPEN") {
        controlsHTML += `<button class="primary-btn" data-task-action="claim">Взять</button>`;
      }
      if (isAssignee && status === "IN_PROGRESS") {
        controlsHTML += `<button class="primary-btn" data-task-action="complete-open">Выполнено</button>`;
      }
      if (canDeleteTasks) {
        controlsHTML += `<button class="ghost-btn danger-btn full" data-task-action="delete">Удалить задачу</button>`;
      }
      controlsHTML += `</div>`;
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

    let tagHTML = "";
    if (priority) {
      tagHTML += `<span class="tag tag-priority tag-priority--${priority.tone}">${escapeHTML(priority.label)}</span>`;
    }
    if (tags.length) {
      tagHTML += tags.map((t) => `<span class="tag">${escapeHTML(t)}</span>`).join("");
    }

    card.innerHTML =
      `<div class="task-summary-head">` +
        `<h4>${escapeHTML(task.title || "Без названия")}</h4>` +
        `<span class="task-state">${escapeHTML(statusMeta.label)}</span>` +
      `</div>` +
      `<p class="task-preview">${escapeHTML(description || "Описание появится после добавления деталей.")}</p>` +
      `<div class="task-summary-meta">` +
        `<div class="task-meta-item">` +
          `<span class="task-meta-label">Роль</span>` +
          `<strong class="task-meta-value">${escapeHTML(task.position_name || task.position_code || "-")}</strong>` +
        `</div>` +
        `<div class="task-meta-item">` +
          `<span class="task-meta-label">Исполнитель</span>` +
          `<strong class="task-meta-value">${escapeHTML(task.assignee_user_id ? getDisplayName(task.assignee_user_id) : "не назначен")}</strong>` +
        `</div>` +
        `<div class="task-meta-item">` +
          `<span class="task-meta-label">Срок</span>` +
          `<strong class="task-meta-value task-deadline ${overdue ? "is-overdue" : ""}">${escapeHTML(dueLabel)}</strong>` +
        `</div>` +
      `</div>` +
      `<div class="task-tags">${tagHTML}</div>` +
      `<details class="task-details">` +
        `<summary class="task-toggle">` +
          `<span class="task-toggle-text task-toggle-text--show">Показать</span>` +
          `<span class="task-toggle-text task-toggle-text--hide">Скрыть</span>` +
        `</summary>` +
        `<div class="task-expanded">` +
          `<div class="task-detail-block">` +
            `<p class="task-detail-label">Подробная информация</p>` +
            `<p class="task-detail-copy">${escapeHTML(description || "Описание отсутствует")}</p>` +
          `</div>` +
          (overdue ? `<div class="task-note overdue">Срок истек, пока задача не завершена.</div>` : "") +
          controlsHTML +
          `<div class="task-timeline-wrap">` +
            `<p class="task-timeline-head">Лента задачи</p>` +
            renderTaskActivityTimeline(task.id || "") +
          `</div>` +
        `</div>` +
      `</details>`;

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

    renderTaskDisplayLimitControl();
    const renderList = (container, items) => {
      if (items.length === 0) {
        container.innerHTML = '<div class="empty-state">Пока пусто</div>';
        return;
      }
      const visible = items.slice(0, state.taskDisplayLimit);
      visible.forEach((task) => container.appendChild(createTaskCard(task)));
      if (items.length > visible.length) {
        const copy = taskLimitI18n();
        const note = document.createElement("div");
        note.className = "kanban-limit-note";
        note.textContent = copy.summary(visible.length, items.length);
        container.appendChild(note);
      }
    };

    renderList(ui.todoTasks, todo);
    renderList(ui.doingTasks, doing);
    renderList(ui.doneTasks, done);

    renderProgress();
    renderTasksTeam();
    renderStackInfoConsole();
    renderTaskList();
  }

  function renderTaskList() {
    if (!ui.taskListBody) return;
    const query = state.searchQuery;
    const filtered = state.tasks.filter((t) => {
      if (!query) return true;
      const hay = `${t.title || ""} ${t.description || ""} ${t.position_name || ""}`.toLowerCase();
      return hay.includes(query);
    });

    const todo  = filtered.filter((t) => taskColumn(t.status) === "todo");
    const doing = filtered.filter((t) => taskColumn(t.status) === "doing");
    const done  = filtered.filter((t) => taskColumn(t.status) === "done");

    const counts = { all: filtered.length, todo: todo.length, doing: doing.length, done: done.length };
    const labels = { all: "Все задачи", todo: "Очередь", doing: "В работе", done: "Завершённые" };

    ui.taskListTabs.forEach((btn) => {
      const tab = btn.dataset.taskTab;
      btn.setAttribute("aria-selected", tab === state.taskListTab ? "true" : "false");
      btn.classList.toggle("active", tab === state.taskListTab);
      btn.innerHTML =
        `${labels[tab]}<span class="tlt-count">${counts[tab]}</span>`;
    });

    const map = { all: filtered, todo, doing, done };
    const rows = map[state.taskListTab] || filtered;

    if (rows.length === 0) {
      ui.taskListBody.innerHTML = '<div class="empty-state">Нет задач в этой категории</div>';
      return;
    }

    ui.taskListBody.innerHTML = rows.map((task) => {
      const col = taskColumn(task.status);
      const statusLabel = col === "todo" ? "Очередь" : col === "doing" ? "В работе" : "Завершено";
      const assignee = task.assignee_user_id ? getDisplayName(task.assignee_user_id) : "не назначен";
      const due = task.due_at ? formatDate(task.due_at) : "";
      const overdue = task.due_at && col !== "done" && new Date(task.due_at) < new Date();
      return (
        `<div class="task-list-row" data-status="${col}">` +
          `<span class="task-list-status-dot task-list-status-dot--${col}"></span>` +
          `<span class="task-list-title">${escapeHTML(task.title || "Без названия")}</span>` +
          `<span class="task-list-assignee">${escapeHTML(assignee)}</span>` +
          (due ? `<span class="task-list-due${overdue ? " is-overdue" : ""}">${escapeHTML(due)}</span>` : `<span></span>`) +
          `<span class="task-list-pill task-list-pill--${col}">${statusLabel}</span>` +
        `</div>`
      );
    }).join("");
  }

  function renderCriteriaView() {
    if (!ui.criteriaListView) return;
    const criteria = Array.isArray(state.criteria) ? state.criteria : [];
    const count = criteria.length;
    if (ui.criteriaCountMeta) {
      ui.criteriaCountMeta.textContent = `${count} ${count === 1 ? "критерий" : count < 5 ? "критерия" : "критериев"}`;
    }

    if (!criteria.length) {
      ui.criteriaListView.innerHTML = '<div class="empty-state">Преподаватель еще не добавил критерии. Студенты здесь только видят этот список и сверяются с ним.</div>';
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
        const penaltyNote = summary.penaltyPercent > 0 ? ` С учетом пересдачи: -${summary.penaltyPercent}%.` : "";
        ui.criteriaReviewHint.textContent = `Проверено: ${summary.met}/${summary.total}. Итоговый балл: ${summary.score}/5.0 (${summary.passPercent}%).${penaltyNote}`;
      } else if (status === "ACTIVE" && projectRetakeCount() > 0) {
        ui.criteriaReviewHint.textContent = "Проект возвращен на пересдачу. После доработки команда сможет снова отправить его преподавателю.";
      } else if (status === "REVIEW" || status === "GRADING" || status === "COMPLETED" || status === "ARCHIVE") {
        ui.criteriaReviewHint.textContent = "Проект ожидает преподавательскую проверку. Итоговая оценка появится после завершения оценки.";
      } else {
        ui.criteriaReviewHint.textContent = "Критерии настраивает преподаватель, а итоговое оценивание появится после завершения проекта и запуска проверки.";
      }
    }
  }

  function renderReviewView() {
    if (!ui.reviewCriteriaList) return;
    const criteria = Array.isArray(state.criteria) ? state.criteria : [];
    const grading = gradingByCriterion();
    const summary = reviewSummaryData();
    const status = projectStatusCode();
    const canDownloadReport = canViewFinalGrade() && (status === "COMPLETED" || status === "ARCHIVE");
    const reportDownloadURL = canDownloadReport ? finalReportURL() : "";
    const reportPreviewURL = canDownloadReport ? finalReportURL("inline") : "";

    if (ui.downloadFinalReportBtn) {
      ui.downloadFinalReportBtn.hidden = !canDownloadReport;
      if (reportDownloadURL) {
        ui.downloadFinalReportBtn.href = reportDownloadURL;
      } else {
        ui.downloadFinalReportBtn.removeAttribute("href");
      }
    }
    if (ui.previewFinalReportBtn) {
      ui.previewFinalReportBtn.hidden = !canDownloadReport;
    }
    if (ui.downloadFinalReportModalBtn) {
      if (reportDownloadURL) {
        ui.downloadFinalReportModalBtn.href = reportDownloadURL;
      } else {
        ui.downloadFinalReportModalBtn.removeAttribute("href");
      }
    }
    if (!canDownloadReport && ui.finalReportPreviewModal && !ui.finalReportPreviewModal.hidden) {
      closeModal(ui.finalReportPreviewModal);
    } else if (reportPreviewURL && ui.finalReportPreviewModal && !ui.finalReportPreviewModal.hidden && ui.finalReportPreviewFrame) {
      ui.finalReportPreviewFrame.src = reportPreviewURL;
    }

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
      if ((status === "COMPLETED" || status === "ARCHIVE") && summary.hasReview) {
        ui.reviewStatusPill.className = "status-pill active";
        ui.reviewStatusPill.textContent = "Проверено преподавателем";
      } else if (status === "ACTIVE" && projectRetakeCount() > 0) {
        ui.reviewStatusPill.className = "status-pill review";
        ui.reviewStatusPill.textContent = "На пересдаче";
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
        if (status === "ACTIVE" && projectRetakeCount() > 0) {
          ui.reviewIntro.textContent = `Проект возвращен на пересдачу. Предыдущая проверка сохранена как ориентир для команды${summary.penaltyPercent > 0 ? `, текущий штраф: ${summary.penaltyPercent}%.` : "."}`;
        } else {
          ui.reviewIntro.textContent = `Ревью по критериям сохранено. Выполнено: ${summary.met}/${summary.total}.${summary.penaltyPercent > 0 ? ` Итог учитывает штраф ${summary.penaltyPercent}% за пересдачу.` : ""}`;
        }
      } else if (status === "GRADING") {
        ui.reviewIntro.textContent = "Проект отправлен преподавателю на оценивание. Результаты появятся после проверки.";
      } else if (status === "ACTIVE" && projectRetakeCount() > 0) {
        ui.reviewIntro.textContent = "Преподаватель вернул проект на доработку. После повторной сдачи итоговая оценка будет немного снижена.";
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
    const detailsMode = canViewProjectDetails();
    const visibleViews = new Set(allowedViews());

    if (ui.applyCard) {
      ui.applyCard.hidden = !applyMode;
    }
    if (ui.stackCard) {
      ui.stackCard.hidden = !detailsMode;
    }
    if (ui.activityCard) {
      ui.activityCard.hidden = !detailsMode;
    }
    if (ui.teamMiniCard) {
      ui.teamMiniCard.hidden = !detailsMode;
    }
    if (ui.pipelineCard) {
      ui.pipelineCard.hidden = !detailsMode;
    }
    if (ui.teamHelperCard) {
      ui.teamHelperCard.hidden = !workspaceMode;
    }
    if (ui.professorAssignWrap) {
      ui.professorAssignWrap.hidden = !workspaceMode || !canInviteProfessorToProject();
    }
    if (ui.positionForm) {
      ui.positionForm.hidden = !workspaceMode || !isCurrentUserLead();
    }
    if (ui.openAccessRoleModalBtn) {
      ui.openAccessRoleModalBtn.hidden = !workspaceMode || !canManageAccess();
    }
    if (ui.openTaskModalBtn) {
      ui.openTaskModalBtn.hidden = !workspaceMode || !hasProjectPermission("task.create");
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
    renderTabHints();
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

  function managedModals() {
    return [
      ui.taskModal,
      ui.permissionsModal,
      ui.accessRoleModal,
      ui.taskResultModal,
      ui.finalReportPreviewModal,
    ].filter(Boolean);
  }

  function closeModal(modal) {
    if (!modal) return;
    modal.hidden = true;
    if (modal === ui.taskResultModal) {
      clearTaskResultForm();
    }
    if (modal === ui.finalReportPreviewModal && ui.finalReportPreviewFrame) {
      ui.finalReportPreviewFrame.removeAttribute("src");
    }
    if (managedModals().every((item) => item.hidden)) {
      document.body.style.overflow = "";
    }
  }

  function renderTaskModalSelects() {
    syncTaskModalStatusOptions();
    ui.taskModalPositionSelect.innerHTML = projectPositionOptions("", { includeSystemTaskPositions: true });
    syncTaskModalAssignees();
  }

  function syncTaskModalAssignees() {
    const positionID = ui.taskModalPositionSelect.value;
    ui.taskModalAssigneeSelect.innerHTML = assigneeOptions(positionID, "");
  }

  function syncTaskModalStatusOptions() {
    if (!ui.taskModalStatusSelect) return;
    const canUpdateTasks = canUpdateTasksInProject();
    const current = String(ui.taskModalStatusSelect.value || "OPEN").toUpperCase();
    ui.taskModalStatusSelect.innerHTML = [
      '<option value="OPEN">Бэклог</option>',
      canUpdateTasks ? '<option value="IN_PROGRESS">В работе</option>' : "",
    ].join("");
    ui.taskModalStatusSelect.value = current === "IN_PROGRESS" && canUpdateTasks ? "IN_PROGRESS" : "OPEN";
  }

  function openTaskModal() {
    if (!canCreateTasksInProject()) {
      setNotice("У вас нет права создавать задачи в этом проекте.", true);
      return;
    }
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

    if (!canCreateTasksInProject()) {
      throw new Error("У вас нет права создавать задачи в этом проекте.");
    }
    if (!title || !positionID) {
      throw new Error("Заполните название задачи и роль.");
    }
    if (status === "IN_PROGRESS" && !canUpdateTasksInProject()) {
      throw new Error("Переводить новую задачу сразу в IN_PROGRESS может только участник с правом изменения статуса.");
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

  async function openPermissionsModal(userID) {
    const member = allMembers().find((m) => String(m.user_id) === String(userID));
    if (!member) return;

    const status = String(member.status || "").toUpperCase();
    const isLeadRow = String(member.user_id || "") === String(state.project?.created_by || "");
    state.currentPermCanManageAccess = canManageAccess() && status === "ACTIVE" && !isLeadRow;
    state.currentPermUserID = userID;
    ui.permMemberName.textContent = getDisplayName(userID);
    ui.permLoading.textContent = "Загрузка данных...";
    ui.permLoading.hidden = false;
    ui.permContent.hidden = true;
    ui.permSystemRoles.innerHTML = "";
    ui.permAssignableRoles.innerHTML = "";
    ui.permEffectivePermissions.innerHTML = "";
    if (ui.permRoleDots) ui.permRoleDots.innerHTML = "";
    if (ui.permEffectiveCount) ui.permEffectiveCount.textContent = "0";
    if (ui.savePermissionsBtn) ui.savePermissionsBtn.disabled = !state.currentPermCanManageAccess;
    openModal(ui.permissionsModal);

    try {
      let accessResp = {
        user_id: userID,
        role_codes: memberLifecycleRoles(member),
        managed_role_codes: [],
        effective_permission_codes: [],
      };

      state.accessCatalog = [];
      if (state.currentPermCanManageAccess) {
        const [catalogResp, fetchedAccessResp] = await Promise.all([
          request("GET", `/v2/projects/${state.projectID}/access/catalog`),
          request("GET", `/v2/projects/${state.projectID}/members/${userID}/access`),
        ]);
        state.accessCatalog = Array.isArray(catalogResp?.items) ? catalogResp.items : [];
        accessResp = fetchedAccessResp || accessResp;
      }

      renderPermissionsModalContent(accessResp, { member });
    } catch (err) {
      ui.permLoading.textContent = `Ошибка загрузки: ${err.message || String(err)}`;
    }
  }

  function renderPermissionsModalContent(access, options = {}) {
    access = access || {};
    const member = options.member || allMembers().find((m) => String(m.user_id) === String(state.currentPermUserID)) || null;
    ui.permLoading.hidden = true;
    ui.permContent.hidden = false;

    const SYSTEM_ROLE_NAMES = {
      TEAM_LEAD: "Тимлид",
      MEMBER: "Участник",
      INVITED_MEMBER: "Приглашён",
      PROJECT_PROFESSOR: "Преподаватель",
      APPLIED: "Заявка",
    };

    // System roles (read-only badges).
    const systemRoles = (access.role_codes || []).filter((c) => SYSTEM_LIFECYCLE_ROLES.has(c));
    ui.permSystemRoles.innerHTML = systemRoles.length
      ? systemRoles.map((c) => `<span class="perm-system-chip">${escapeHTML(SYSTEM_ROLE_NAMES[c] || c)}</span>`).join("")
      : '<span class="perm-empty">Нет базовых ролей</span>';

    if (!state.currentPermCanManageAccess) {
      ui.permAssignableRoles.innerHTML = '<div class="perm-empty-panel">Роли доступа может назначать только тимлид или участник с правом управления доступом.</div>';
      if (ui.permRoleDots) ui.permRoleDots.innerHTML = "";
      if (ui.permEffectiveCount) ui.permEffectiveCount.textContent = "0";
      ui.permEffectivePermissions.innerHTML = '<span class="perm-empty">Нет данных о правах доступа</span>';
      return;
    }

    // Access role radios. A project member can have only one delegated/custom role.
    const selectedRoleCode = String((access.managed_role_codes || [])[0] || "");
    const renderDots = () => {
      if (!ui.permRoleDots) return;
      const selected = String(ui.permAssignableRoles.querySelector("input[name='project_access_role']:checked")?.value || "");
      ui.permRoleDots.innerHTML = state.accessCatalog.map((item) => {
        const active = selected === item.code ? " active" : "";
        return `<span class="perm-role-dot${active}"></span>`;
      }).join("");
    };

    ui.permAssignableRoles.innerHTML = "";
    const renderRoleCard = (item) => {
      const label = document.createElement("label");
      const itemCode = String(item.code || "");
      label.className = "perm-role-card" + (selectedRoleCode === itemCode ? " active" : "");
      const checked = selectedRoleCode === itemCode ? "checked" : "";
      const asset = ROLE_ASSETS[item.code] || "/dev/static/assets/role-access.svg";
      const permissionCodes = Array.isArray(item.permission_codes) ? item.permission_codes : [];
      const rightsList = permissionCodes.length
        ? permissionCodes.map((p) => `<li><span>${escapeHTML(permissionLabel(p))}</span><code>${escapeHTML(p)}</code></li>`).join("")
        : "<li><span>Нет дополнительных прав</span></li>";
      label.innerHTML =
        `<input class="perm-role-input" type="radio" name="project_access_role" value="${escapeHTML(itemCode)}" data-role-code="${escapeHTML(itemCode)}" ${checked} aria-label="Назначить роль ${escapeHTML(item.name)}" />` +
        `<span class="perm-role-check" aria-hidden="true"></span>` +
        `<div class="perm-role-card-body">` +
          `<div class="perm-role-asset" aria-hidden="true">` +
            `<img src="${asset}" alt="" width="40" height="40" />` +
          `</div>` +
          `<strong>${escapeHTML(item.name)}</strong>` +
          `<span class="perm-role-code">${escapeHTML(item.display_code || item.code || "NO_ACCESS")}</span>` +
          `<span class="perm-role-desc">${escapeHTML(item.description)}</span>` +
          `<div class="perm-role-rights">` +
            `<span class="perm-role-rights-title">Права роли</span>` +
            `<ul>${rightsList}</ul>` +
          `</div>` +
        `</div>`;

      label.querySelector("input").addEventListener("change", () => {
        ui.permAssignableRoles.querySelectorAll(".perm-role-card").forEach((card) => {
          card.classList.toggle("active", Boolean(card.querySelector("input")?.checked));
        });
        renderDots();
      });

      ui.permAssignableRoles.appendChild(label);
    };

    renderRoleCard({
      code: "",
      display_code: "NO_ACCESS",
      name: "Без роли доступа",
      description: "Оставить участнику только базовый проектный доступ без дополнительных прав.",
      permission_codes: [],
    });

    if (!state.accessCatalog.length) {
      ui.permAssignableRoles.insertAdjacentHTML("beforeend", '<div class="perm-empty-panel">Нет дополнительных ролей доступа для назначения.</div>');
    }
    state.accessCatalog.forEach((item) => {
      renderRoleCard(item);
    });
    renderDots();

    // Effective permissions (read-only).
    const effectivePerms = access.effective_permission_codes || [];
    if (ui.permEffectiveCount) {
      ui.permEffectiveCount.textContent = String(effectivePerms.length);
    }
    ui.permEffectivePermissions.innerHTML = effectivePerms.length
      ? effectivePerms.map((c) => `<span class="perm-eff-chip" title="${escapeHTML(c)}">${escapeHTML(permissionLabel(c))}</span>`).join("")
      : '<span class="perm-empty">Нет разрешений</span>';
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
    const confirmed = await confirmAction({
      title: "Удалить проект",
      message: "Проект будет удален без возможности восстановления. Это действие затронет команду, задачи и материалы проекта.",
      confirmText: "Удалить проект",
      danger: true,
    });
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

  function syncAccessRolePermissionCount() {
    if (!ui.accessRolePermissionCount || !ui.accessRolePermissionsList) return;
    const count = ui.accessRolePermissionsList.querySelectorAll("input[type='checkbox']:checked").length;
    ui.accessRolePermissionCount.textContent = String(count);
  }

  function renderAccessRolePermissions() {
    if (!ui.accessRolePermissionsList) return;
    ui.accessRolePermissionsList.innerHTML = "";

    if (!Array.isArray(state.accessPermissionCatalog) || state.accessPermissionCatalog.length === 0) {
      ui.accessRolePermissionsList.innerHTML = '<div class="empty-state">Список прав пока недоступен.</div>';
      syncAccessRolePermissionCount();
      return;
    }

    state.accessPermissionCatalog.forEach((item) => {
      const label = document.createElement("label");
      label.className = "access-role-permission-item";
      label.innerHTML =
        `<input type="checkbox" value="${escapeHTML(item.code)}" />` +
        `<span class="access-role-permission-check" aria-hidden="true"></span>` +
        `<span class="access-role-permission-copy">` +
          `<strong>${escapeHTML(permissionLabel(item.code))}</strong>` +
          `<small>${escapeHTML(item.description || item.code)}</small>` +
          `<code>${escapeHTML(item.code)}</code>` +
        `</span>`;
      label.querySelector("input").addEventListener("change", () => {
        label.classList.toggle("checked", Boolean(label.querySelector("input")?.checked));
        syncAccessRolePermissionCount();
      });
      ui.accessRolePermissionsList.appendChild(label);
    });
    syncAccessRolePermissionCount();
  }

  async function loadAccessRolePermissionCatalog() {
    const resp = await request("GET", `/v2/projects/${state.projectID}/access/permissions`);
    state.accessPermissionCatalog = Array.isArray(resp?.items) ? resp.items : [];
  }

  async function openAccessRoleModal() {
    if (!canManageAccess()) {
      throw new Error("Создавать роли доступа может только участник с правом member.access.manage.");
    }

    if (ui.accessRoleNameInput) ui.accessRoleNameInput.value = "";
    if (ui.accessRoleCodeInput) ui.accessRoleCodeInput.value = "";
    if (ui.accessRoleDescriptionInput) ui.accessRoleDescriptionInput.value = "";
    if (ui.accessRolePermissionsList) ui.accessRolePermissionsList.innerHTML = "";
    if (ui.accessRoleLoading) {
      ui.accessRoleLoading.hidden = false;
      ui.accessRoleLoading.textContent = "Загрузка прав...";
    }
    if (ui.saveAccessRoleBtn) ui.saveAccessRoleBtn.disabled = true;
    syncAccessRolePermissionCount();
    openModal(ui.accessRoleModal);

    try {
      await loadAccessRolePermissionCatalog();
      if (ui.accessRoleLoading) ui.accessRoleLoading.hidden = true;
      renderAccessRolePermissions();
      if (ui.saveAccessRoleBtn) ui.saveAccessRoleBtn.disabled = false;
    } catch (err) {
      if (ui.accessRoleLoading) {
        ui.accessRoleLoading.hidden = false;
        ui.accessRoleLoading.textContent = `Ошибка загрузки прав: ${err.message || String(err)}`;
      }
      if (ui.saveAccessRoleBtn) ui.saveAccessRoleBtn.disabled = true;
    }
  }

  async function createAccessRoleFromModal() {
    const name = String(ui.accessRoleNameInput?.value || "").trim();
    const code = String(ui.accessRoleCodeInput?.value || "").trim();
    const description = String(ui.accessRoleDescriptionInput?.value || "").trim();
    const permissionCodes = Array.from(ui.accessRolePermissionsList?.querySelectorAll("input[type='checkbox']:checked") || [])
      .map((input) => String(input.value || "").trim())
      .filter(Boolean);

    if (!name || !code) {
      throw new Error("Заполните название и код роли.");
    }

    await request("POST", `/v2/projects/${state.projectID}/access/roles`, {
      name,
      code,
      description,
      permission_codes: permissionCodes,
    });

    closeModal(ui.accessRoleModal);
    setNotice("Роль доступа создана. Она появится в списке ролей участника.", false);
  }

  async function onMemberAction(actionBtn) {
    const row = actionBtn.closest("tr[data-user-id]");
    if (!row) return;

    const userID = row.getAttribute("data-user-id");
    const action = actionBtn.getAttribute("data-member-action");
    const status = String(row.getAttribute("data-member-status") || "").trim().toUpperCase();

    if (action === "role" || action === "permissions") {
      if (status !== "ACTIVE") {
        throw new Error("Роль можно менять только для активных участников.");
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
      await request("POST", `/v2/projects/${state.projectID}/members/${userID}/approve`, {});
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

    if (action === "remove") {
      const confirmed = await confirmAction({
        title: "Удалить участника",
        message: "Участник будет исключен из проекта. Если за ним были закреплены задачи, назначения будут сняты.",
        confirmText: "Удалить участника",
        danger: true,
      });
      if (!confirmed) {
        return;
      }
      await request("DELETE", `/v2/projects/${state.projectID}/members/${userID}`);
      setNotice(`Участник ${shortID(userID)} удален из проекта.`, false);
      await refreshData();
      return;
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

    if (action === "delete") {
      const titleNode = card.querySelector("h4");
      const taskTitle = String(titleNode ? titleNode.textContent : "").trim();
      const confirmed = await confirmAction({
        title: "Удалить задачу",
        message: taskTitle
          ? `Задача «${taskTitle}» будет удалена вместе с ее историей выполнения.`
          : "Задача будет удалена вместе с ее историей выполнения.",
        confirmText: "Удалить задачу",
        danger: true,
      });
      if (!confirmed) {
        return;
      }
      await request("DELETE", `/v2/projects/${state.projectID}/tasks/${taskID}`);
      if (state.taskMeta && state.taskMeta[taskID]) {
        delete state.taskMeta[taskID];
        saveJSON(taskMetaKey(), state.taskMeta);
      }
      setNotice(taskTitle ? `Задача «${taskTitle}» удалена.` : `Задача ${shortID(taskID)} удалена.`, false);
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
    setNotice("Проект переведен в активную фазу.", false);
    await refreshData();
  }

  async function onSubmitProjectForGrading() {
    const confirmed = await confirmAction({
      title: "Отправить проект на оценивание",
      message: "После подтверждения проект перейдет в статус GRADING и будет ждать финального решения преподавателя.",
      confirmText: "Отправить на оценивание",
    });
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

  async function savePermissions() {
    if (!state.currentPermUserID) return;

    if (!state.currentPermCanManageAccess) {
      setNotice("У вас нет права менять роли доступа.", true);
      return;
    }

    const selectedRole = String(ui.permAssignableRoles.querySelector("input[name='project_access_role']:checked")?.value || "");
    const selectedRoles = selectedRole ? [selectedRole] : [];

    try {
      const access = await request("PUT", `/v2/projects/${state.projectID}/members/${state.currentPermUserID}/access`, {
        managed_role_codes: selectedRoles,
      });

      await refreshData();
      closeModal(ui.permissionsModal);
      setNotice("Роль доступа участника обновлена.", false);
    } catch (err) {
      const friendly = roleSaveErrorMessage(err);
      setNotice(`Ошибка сохранения роли: ${friendly}`, true);
    }
  }

  async function refreshData() {
    state.project = await request("GET", `/v2/projects/${state.projectID}`);
    rememberUser(localStorage.getItem(LS_USER), localStorage.getItem(LS_STUDENT_NAME), localStorage.getItem(LS_STUDENT_EMAIL));
    rememberUser(state.project?.created_by, state.project?.created_by_name, state.project?.created_by_email);

    if (!canViewProjectDetails()) {
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

    const [stacks, positions, members, readiness, criteria, tasks, gradingResp, taskActivityResp, professorResp, myPermsResp] = await Promise.all([
      loadOptional("stacks", "GET", `/v2/projects/${state.projectID}/stacks`, []),
      loadOptional("positions", "GET", `/v2/projects/${state.projectID}/positions`, []),
      loadOptional("members", "GET", `/v2/projects/${state.projectID}/members`, []),
      loadOptional("readiness", "GET", `/v2/projects/${state.projectID}/readiness`, null),
      loadOptional("criteria", "GET", `/v2/projects/${state.projectID}/criteria`, []),
      loadOptional("tasks", "GET", `/v2/projects/${state.projectID}/tasks`, []),
      loadOptional("grading", "GET", `/v2/projects/${state.projectID}/grading`, { items: [] }),
      loadOptional("task activity", "GET", `/v2/projects/${state.projectID}/tasks/activity`, { items: [] }),
      loadOptional("assigned professor", "GET", `/v2/projects/${state.projectID}/professor`, { professor: null }),
      loadOptional("my permissions", "GET", `/v2/projects/${state.projectID}/my-permissions`, { permissions: [] }),
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
    state.myPermissions = myPermsResp && Array.isArray(myPermsResp.permissions) ? myPermsResp.permissions : [];

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

    managedModals().forEach((modal) => {
      if (!modal) return;
      modal.addEventListener("click", (e) => {
        if (e.target === modal) {
          closeModal(modal);
        }
      });
    });

    document.addEventListener("keydown", (e) => {
      if (e.key !== "Escape") return;
      managedModals().forEach((modal) => {
        if (!modal.hidden) closeModal(modal);
      });
    });
  }

  function wireEvents() {
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

    if (ui.previewFinalReportBtn) {
      ui.previewFinalReportBtn.addEventListener("click", () => {
        const previewURL = finalReportURL("inline");
        if (!previewURL || !ui.finalReportPreviewModal || !ui.finalReportPreviewFrame) return;
        ui.finalReportPreviewFrame.src = previewURL;
        openModal(ui.finalReportPreviewModal);
      });
    }

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

    if (ui.positionForm) {
      ui.positionForm.addEventListener("submit", async (e) => {
        try {
          await onCreatePosition(e);
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

    if (ui.openAccessRoleModalBtn) {
      ui.openAccessRoleModalBtn.addEventListener("click", async () => {
        try {
          await openAccessRoleModal();
        } catch (err) {
          setNotice(err.message || String(err), true);
        }
      });
    }

    if (ui.accessRoleNameInput && ui.accessRoleCodeInput) {
      ui.accessRoleNameInput.addEventListener("input", () => {
        if (String(ui.accessRoleCodeInput.value || "").trim()) return;
        ui.accessRoleCodeInput.value = String(ui.accessRoleNameInput.value || "")
          .trim()
          .toUpperCase()
          .replace(/[^A-Z0-9]+/g, "_")
          .replace(/^_+|_+$/g, "");
      });
    }

    if (ui.saveAccessRoleBtn) {
      ui.saveAccessRoleBtn.addEventListener("click", async () => {
        setButtonLoading(ui.saveAccessRoleBtn, true, "Создание...");
        try {
          await createAccessRoleFromModal();
        } catch (err) {
          setNotice(err.message || String(err), true);
        } finally {
          setButtonLoading(ui.saveAccessRoleBtn, false, "Создать роль");
        }
      });
    }

    ui.teamTableBody.addEventListener("click", async (e) => {
      const btn = e.target.closest("button[data-member-action]");
      if (!btn) return;
      try {
        await onMemberAction(btn);
      } catch (err) {
        setNotice(err.message || String(err), true);
      }
    });

    if (ui.readinessList) {
      ui.readinessList.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-stage-action]");
        if (!btn) return;
        runStageAction(btn.getAttribute("data-stage-action") || "");
      });
    }

    [ui.teamStageHint, ui.tasksStageHint, ui.criteriaStageHint].forEach((target) => {
      if (!target) return;
      target.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-stage-hints-visibility]");
        if (!btn) return;
        setStageHintsHidden(btn.getAttribute("data-stage-hints-visibility") === "hide");
      });
    });

    if (ui.hideStageHintsBtn) {
      ui.hideStageHintsBtn.addEventListener("click", () => {
        setStageHintsHidden(true);
      });
    }

    if (ui.showStageHintsBtn) {
      ui.showStageHintsBtn.addEventListener("click", () => {
        setStageHintsHidden(false);
      });
    }

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

    ui.taskListTabs.forEach((btn) => {
      btn.addEventListener("click", () => {
        state.taskListTab = btn.dataset.taskTab || "all";
        renderTaskList();
      });
    });

    if (ui.savePermissionsBtn) {
      ui.savePermissionsBtn.addEventListener("click", savePermissions);
    }

    ui.taskDisplayLimitButtons.forEach((btn) => {
      btn.addEventListener("click", () => {
        const limit = normalizeTaskDisplayLimit(btn.dataset.taskLimit);
        if (limit === state.taskDisplayLimit) return;
        setTaskDisplayLimit(limit);
        renderTasks();
      });
    });

    window.addEventListener("idsai:languagechange", () => {
      renderTaskDisplayLimitControl();
      renderTasks();
      renderReviewView();
    });
  }

  async function bootstrap() {
    const claims = await auth.ensureSession();
    if (!claims) return;
    if (claims.is_admin) {
      window.location.href = "/dev/admin";
      return;
    }

    bindProfile(claims);

    state.projectID = projectIDFromPath();
    if (!state.projectID) {
      window.location.href = claims.is_professor ? "/dev/professor" : "/dev/projects";
      return;
    }

    state.projectMeta = loadJSON(projectMetaKey(), {});
    state.taskMeta = loadJSON(taskMetaKey(), {});
    state.taskDisplayLimit = loadTaskDisplayLimit();
    state.favorite = Boolean(state.projectMeta.favorite);

    wireTabSwitching();
    wireModals();
    wireEvents();

    try {
      await refreshData();
      setView(initialViewFromURL());
      auth.setPageLoading(false);
    } catch (err) {
      if (err && err.status === 404 && auth && typeof auth.redirectToNotFound === "function") {
        auth.redirectToNotFound();
        return;
      }
      setNotice(`Не удалось загрузить проект: ${err.message || String(err)}`, true);
      auth.setPageLoading(false);
    }
  }

  bootstrap().catch((err) => {
    auth.setPageLoading(false);
    setNotice(`Не удалось загрузить проект: ${err.message || String(err)}`, true);
  });
})();
