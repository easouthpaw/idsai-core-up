(() => {
  const auth = window.IDSAIAuth;
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
  const DEFAULT_PROJECT_COVERS = [
    "https://images.pexels.com/photos/16129724/pexels-photo-16129724.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/17323801/pexels-photo-17323801.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/4508751/pexels-photo-4508751.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/5257576/pexels-photo-5257576.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/10499056/pexels-photo-10499056.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
    "https://images.pexels.com/photos/12899157/pexels-photo-12899157.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=900&w=1600",
  ];

  const createStatusEl = document.getElementById("createStatus");
  const myProjectsEl = document.getElementById("myProjects");
  const publicProjectsEl = document.getElementById("publicProjects");
  const groupDepartmentEl = document.getElementById("groupDepartment");
  const groupNumberEl = document.getElementById("groupNumber");
  const tabButtons = Array.from(document.querySelectorAll(".main-nav [data-tab]"));
  const paneMineEl = document.getElementById("paneMine");
  const paneCommunityEl = document.getElementById("paneCommunity");
  const toolbarMineEl = document.getElementById("toolbarMine");
  const toolbarCommunityEl = document.getElementById("toolbarCommunity");
  const tabTitleEl = document.getElementById("tabTitle");
  const tabSubtitleEl = document.getElementById("tabSubtitle");
  const tabCounterEl = document.getElementById("tabCounter");
  const crumbTabEl = document.getElementById("crumbTab");
  const searchInputEl = document.getElementById("searchInput");
  const communitySearchInputEl = document.getElementById("communitySearchInput");
  const communityTechFilterEl = document.getElementById("communityTechFilter");
  const communityDifficultyFilterEl = document.getElementById("communityDifficultyFilter");
  const createToggleBtnEl = document.getElementById("createToggleBtn");
  const createCancelBtnEl = document.getElementById("createCancelBtn");
  const createCloseBtnEl = document.getElementById("createCloseBtn");
  const createModalEl = document.getElementById("createModal");
  const privateGroupRowEl = document.getElementById("privateGroupRow");
  const consoleLogEl = document.getElementById("consoleLog");
  const filterButtons = Array.from(document.querySelectorAll(".filter-btn"));
  const viewButtons = Array.from(document.querySelectorAll(".view-btn"));
  const visibilityCards = Array.from(document.querySelectorAll(".visibility-card"));

  let groupOptions = [];
  let activeTab = "mine";
  let activeFilter = "all";
  let activeView = "grid";
  let selectedVisibility = "PUBLIC";
  let myProjects = [];
  let publicProjects = [];

  function logLine(text) {
    const now = new Date().toLocaleTimeString();
    const previous = (consoleLogEl.textContent || "").trim();
    const line = `[${now}] ${text}`;
    consoleLogEl.textContent = previous ? `${previous}\n${line}` : line;
    consoleLogEl.scrollTop = consoleLogEl.scrollHeight;
  }

  function setStatus(el, msg, ok) {
    if (el) {
      el.textContent = msg;
      el.className = `status ${ok ? "ok" : "err"}`;
    }
    logLine(msg);
  }

  function decodePayload(token) {
    const parts = token.split(".");
    if (parts.length < 2) throw new Error("invalid JWT");
    let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const mod = payload.length % 4;
    if (mod > 0) payload += "=".repeat(4 - mod);
    return JSON.parse(atob(payload));
  }

  function authHeaders(withJSON) {
    const headers = {};
    if (withJSON) headers["Content-Type"] = "application/json";
    return headers;
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

  function syncSidebar(profile, active = activeTab) {
    const host = document.querySelector("[data-role-sidebar]");
    if (!host || !roleSidebar || typeof roleSidebar.renderSidebar !== "function") {
      return;
    }

    host.dataset.sidebarActive = active;
    roleSidebar.renderSidebar(host, {
      role: "student",
      active,
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
    syncSidebar(profile || auth.getCachedProfile(), activeTab);
  }

  function escapeHTML(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function statusKind(status) {
    const s = String(status || "").toUpperCase();
    if (s === "ACTIVE" || s === "RECRUITMENT" || s === "DRAFT") return "inwork";
    if (s === "REVIEW" || s === "GRADING") return "review";
    if (s === "COMPLETED" || s === "ARCHIVE" || s === "DONE") return "done";
    return "default";
  }

  function projectStatusLabel(status) {
    const s = String(status || "").toUpperCase();
    if (s === "DRAFT" || s === "REVIEW") return "Подготовка";
    if (s === "RECRUITMENT") return "Набор";
    if (s === "ACTIVE") return "В работе";
    if (s === "GRADING") return "Оценивание";
    if (s === "COMPLETED" || s === "ARCHIVE" || s === "DONE") return "Завершен";
    return s || "-";
  }

  function visibilityLabel(value) {
    const v = String(value || "").toUpperCase();
    if (v === "PUBLIC") return "PUBLIC";
    if (v === "GROUP" || v === "FACULTY" || v === "PRIVATE") return "PRIVATE";
    return v || "-";
  }

  function formatDate(raw) {
    if (!raw) return "-";
    try {
      return new Date(raw).toLocaleString();
    } catch (_) {
      return String(raw);
    }
  }

  function openProject(project) {
    if (!project || !project.id) return;
    const navSection = activeTab === "community" ? "community" : "mine";
    localStorage.setItem(LS_STUDENT_SECTION, navSection);
    localStorage.setItem(LS_SELECTED_PROJECT, JSON.stringify({
      ...project,
      _nav_section: navSection,
    }));
    window.location.href = `/dev/projects/${project.id}?nav=${encodeURIComponent(navSection)}`;
  }

  function findProjectByID(id) {
    return myProjects.find((p) => String(p.id) === String(id)) ||
      publicProjects.find((p) => String(p.id) === String(id)) ||
      null;
  }

  function stableHash(input) {
    const s = String(input || "");
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

  function progressForProject(p) {
    const s = String(p.status || "").toUpperCase();
    if (s === "COMPLETED" || s === "ARCHIVE" || s === "DONE") return 100;
    if (s === "GRADING") return 90;
    if (s === "ACTIVE") return 70;
    if (s === "REVIEW") return 60;
    if (s === "RECRUITMENT") return 35;
    return 20;
  }

  function minePillLabel(p) {
    return projectStatusLabel(p.status);
  }

  function communityDifficulty(p) {
    const s = String(p.status || "").toUpperCase();
    if (s === "REVIEW" || s === "GRADING") return "ADVANCED";
    if (s === "ACTIVE") return "MEDIUM";
    if (s === "COMPLETED" || s === "ARCHIVE" || s === "DONE") return "ADVANCED";
    return "BEGINNER";
  }

  function communityDifficultyLabel(code) {
    if (code === "ADVANCED") return "Продвинутый";
    if (code === "MEDIUM") return "Средний";
    return "Новичок";
  }

  function communityDifficultyClass(code) {
    if (code === "ADVANCED") return "advanced";
    if (code === "MEDIUM") return "medium";
    return "beginner";
  }

  function inferTechTags(project) {
    const t = `${project.title || ""} ${project.description || ""}`.toLowerCase();
    const tags = [];
    const pushTag = (cond, label) => {
      if (cond && !tags.includes(label)) tags.push(label);
    };

    pushTag(t.includes("python"), "Python");
    pushTag(t.includes("go"), "Go");
    pushTag(t.includes("react"), "React");
    pushTag(t.includes("flutter"), "Flutter");
    pushTag(t.includes("node"), "Node.js");
    pushTag(t.includes("rust"), "Rust");
    pushTag(t.includes("js") || t.includes("javascript"), "JavaScript");
    pushTag(t.includes("ts") || t.includes("typescript"), "TypeScript");
    pushTag(t.includes("sql") || t.includes("postgres"), "PostgreSQL");

    if (tags.length === 0) tags.push("General");
    return tags.slice(0, 3);
  }

  function parseStacks(input) {
    const seen = new Set();
    return String(input || "")
      .split(",")
      .map((x) => x.trim().toUpperCase())
      .filter((x) => {
        if (!x) return false;
        if (seen.has(x)) return false;
        seen.add(x);
        return true;
      });
  }

  function isRecruiting(project) {
    const s = String(project.status || "").toUpperCase();
    return s === "RECRUITMENT" || s === "DRAFT";
  }

  function setVisibility(v) {
    selectedVisibility = v === "PRIVATE" ? "PRIVATE" : "PUBLIC";
    visibilityCards.forEach((card) => {
      card.classList.toggle("active", card.dataset.visibility === selectedVisibility);
    });

    const isPrivate = selectedVisibility === "PRIVATE";
    privateGroupRowEl.hidden = !isPrivate;
    groupDepartmentEl.disabled = !isPrivate;
    groupNumberEl.disabled = !isPrivate;
  }

  function openCreateModal() {
    if (createToggleBtnEl.hidden || createToggleBtnEl.disabled) {
      return;
    }
    createModalEl.hidden = false;
    document.body.classList.add("modal-open");
    setVisibility(selectedVisibility);
    createStatusEl.textContent = "";
  }

  function closeCreateModal() {
    createModalEl.hidden = true;
    document.body.classList.remove("modal-open");
    createStatusEl.textContent = "";
  }

  function applyMineFilters(items) {
    const search = (searchInputEl.value || "").trim().toLowerCase();
    return items.filter((p) => {
      const kind = statusKind(p.status);
      if (activeFilter !== "all" && kind !== activeFilter) return false;

      if (!search) return true;
      const hay = `${p.title || ""} ${p.description || ""}`.toLowerCase();
      return hay.includes(search);
    });
  }

  function applyCommunityFilters(items) {
    const search = (communitySearchInputEl.value || "").trim().toLowerCase();
    const tech = communityTechFilterEl.value || "ALL";
    const difficulty = communityDifficultyFilterEl.value || "ALL";

    return items.filter((p) => {
      const tags = inferTechTags(p);
      const diff = communityDifficulty(p);
      const hay = `${p.title || ""} ${p.description || ""} ${tags.join(" ")}`.toLowerCase();

      if (search && !hay.includes(search)) return false;
      if (tech !== "ALL" && !tags.some((tag) => tag.toUpperCase() === tech)) return false;
      if (difficulty !== "ALL" && diff !== difficulty) return false;
      return true;
    });
  }

  function updateCounts(items) {
    const stats = {
      all: items.length,
      inwork: 0,
      review: 0,
      done: 0,
    };

    items.forEach((p) => {
      const kind = statusKind(p.status);
      if (kind === "inwork") stats.inwork += 1;
      if (kind === "review") stats.review += 1;
      if (kind === "done") stats.done += 1;
    });

    document.getElementById("countAll").textContent = String(stats.all);
    document.getElementById("countInwork").textContent = String(stats.inwork);
    document.getElementById("countReview").textContent = String(stats.review);
    document.getElementById("countDone").textContent = String(stats.done);
  }

  function renderMine() {
    const filtered = applyMineFilters(myProjects);
    const canCreate = !createToggleBtnEl.hidden && !createToggleBtnEl.disabled;
    updateCounts(myProjects);

    myProjectsEl.classList.toggle("list-view", activeView === "list");
    myProjectsEl.innerHTML = "";

    if (filtered.length === 0) {
      myProjectsEl.innerHTML = canCreate
        ? '<article class="empty-card new-project" id="newProjectCard">Нет проектов под текущий фильтр.<br>Нажми, чтобы создать новый.</article>'
        : '<article class="empty-card">Нет проектов под текущий фильтр.</article>';
      const emptyCard = canCreate ? document.getElementById("newProjectCard") : null;
      if (emptyCard) {
        emptyCard.addEventListener("click", openCreateModal);
      }
      return;
    }

    filtered.forEach((p) => {
      const kind = statusKind(p.status);
      const progress = progressForProject(p);
      const owner = (localStorage.getItem(LS_STUDENT_NAME) || "ME").trim().slice(0, 2).toUpperCase();
      const article = document.createElement("article");
      article.className = "project-card mine-card";
      const pid = escapeHTML(p.id || "");

      article.innerHTML =
        `<div class="project-cover">` +
          `<img src="${escapeHTML(projectCoverURL(p))}" data-fallback-cover="${escapeHTML(defaultCoverURL(p))}" alt="Project cover" width="1600" height="900" loading="lazy" />` +
        `</div>` +
        `<div class="card-head">` +
          `<button class="card-title-link" data-open-id="${pid}" type="button">${escapeHTML(p.title || "-")}</button>` +
          `<span class="card-menu" aria-hidden="true">…</span>` +
        `</div>` +
        `<p class="card-desc">${escapeHTML(p.description || "Без описания")}</p>` +
        `<div class="mine-progress"><span style="width:${progress}%"></span></div>` +
        `<div class="card-meta">` +
          `<span>created: ${escapeHTML(formatDate(p.created_at))}</span>` +
          `<span>status: ${escapeHTML(projectStatusLabel(p.status))}</span>` +
        `</div>` +
        `<div class="mine-footer">` +
          `<span class="badge ${kind}">${escapeHTML(minePillLabel(p))}</span>` +
          `<span class="mine-owner">${escapeHTML(owner)}</span>` +
        `</div>` +
        `<div class="card-actions">` +
          `<span class="visibility-pill">${escapeHTML(visibilityLabel(p.visibility))}</span>` +
          `<button class="detail-btn" data-open-id="${pid}" type="button">Открыть</button>` +
        `</div>`;

      myProjectsEl.appendChild(article);
    });

    if (canCreate) {
      const createCard = document.createElement("button");
      createCard.type = "button";
      createCard.className = "empty-card new-project";
      createCard.innerHTML = "＋<br>Новый проект";
      createCard.addEventListener("click", openCreateModal);
      myProjectsEl.appendChild(createCard);
    }

    myProjectsEl.querySelectorAll("[data-open-id]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const id = btn.getAttribute("data-open-id");
        const project = findProjectByID(id);
        if (!project) return;
        openProject(project);
      });
    });
    bindCoverFallbacks(myProjectsEl);
  }

  function renderCommunity() {
    const activeCount = publicProjects.filter((p) => isRecruiting(p)).length;
    tabCounterEl.textContent = `${activeCount} активных`;

    const filtered = applyCommunityFilters(publicProjects);
    publicProjectsEl.classList.toggle("list-view", activeView === "list");
    publicProjectsEl.innerHTML = "";

    if (filtered.length === 0) {
      publicProjectsEl.innerHTML = '<article class="empty-card">Публичные проекты не найдены.</article>';
      return;
    }

    filtered.forEach((p) => {
      const diffCode = communityDifficulty(p);
      const diffClass = communityDifficultyClass(diffCode);
      const diffLabel = communityDifficultyLabel(diffCode);
      const tags = inferTechTags(p);
      const hash = stableHash(p.id || p.title || p.description || "");
      const membersCur = (hash % 4) + 1;
      const membersMax = membersCur + ((hash % 3) + 2);
      const recruiting = isRecruiting(p);
      const article = document.createElement("article");
      article.className = "project-card community-card";
      const pid = escapeHTML(p.id || "");
      article.innerHTML =
        `<div class="project-cover">` +
          `<img src="${escapeHTML(projectCoverURL(p))}" data-fallback-cover="${escapeHTML(defaultCoverURL(p))}" alt="Project cover" width="1600" height="900" loading="lazy" />` +
        `</div>` +
        `<div class="card-head">` +
          `<button class="card-title-link" data-open-id="${pid}" type="button">${escapeHTML(p.title || "-")}</button>` +
          `<span class="community-level ${diffClass}">${escapeHTML(diffLabel)}</span>` +
        `</div>` +
        `<p class="card-desc">${escapeHTML(p.description || "Без описания")}</p>` +
        `<div class="tech-tags">${tags.map((tag) => `<span class="tech-chip">${escapeHTML(tag)}</span>`).join("")}</div>` +
        `<div class="card-meta">` +
          `<span>author: ${escapeHTML(p.created_by_name || p.created_by || "-")}</span>` +
          `<span>created: ${escapeHTML(formatDate(p.created_at))}</span>` +
        `</div>` +
        `<div class="community-footer">` +
          `<span class="participants">👥 ${membersCur}/${membersMax} участников</span>` +
          `<button class="detail-btn" data-open-id="${pid}" type="button">${recruiting ? "Открыть проект" : "Набор закрыт"}</button>` +
        `</div>`;
      publicProjectsEl.appendChild(article);
    });

    publicProjectsEl.querySelectorAll("[data-open-id]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const id = btn.getAttribute("data-open-id");
        const project = findProjectByID(id);
        if (!project) return;
        openProject(project);
      });
    });
    bindCoverFallbacks(publicProjectsEl);
  }

  function bindCoverFallbacks(container) {
    if (!container) return;
    container.querySelectorAll("img[data-fallback-cover]").forEach((img) => {
      img.addEventListener("error", () => {
        const fallback = String(img.getAttribute("data-fallback-cover") || "").trim();
        if (fallback && img.src !== fallback) {
          img.src = fallback;
        }
      });
    });
  }

  function switchTab(tab) {
    activeTab = tab === "community" ? "community" : "mine";
    localStorage.setItem(LS_STUDENT_SECTION, activeTab);
    bindProfile(auth.getCachedProfile());
    const isMine = activeTab === "mine";
    const url = new URL(window.location.href);
    url.searchParams.set("tab", activeTab);
    window.history.replaceState({}, "", url.toString());

    tabButtons.forEach((btn) => {
      btn.classList.toggle("active", btn.dataset.tab === activeTab);
    });

    paneMineEl.classList.toggle("active", isMine);
    paneCommunityEl.classList.toggle("active", !isMine);
    toolbarMineEl.hidden = !isMine;
    toolbarCommunityEl.hidden = isMine;

    if (isMine) {
      tabTitleEl.textContent = "Мои проекты";
      tabSubtitleEl.textContent = "Управляйте репозиториями и отслеживайте прогресс команды.";
      crumbTabEl.textContent = "projects";
      tabCounterEl.hidden = true;
      renderMine();
    } else {
      const activeCount = publicProjects.filter((p) => isRecruiting(p)).length;
      tabTitleEl.textContent = "Обзор проектов";
      tabSubtitleEl.textContent = "Публичные проекты на стадии набора и активной разработки.";
      crumbTabEl.textContent = "public-overview";
      tabCounterEl.hidden = false;
      tabCounterEl.textContent = `${activeCount} активных`;
      renderCommunity();
    }
  }

  function initialTabFromURL() {
    const tab = String(new URLSearchParams(window.location.search).get("tab") || "").toLowerCase();
    return tab === "community" ? "community" : "mine";
  }

  function fillGroupDepartments() {
    const departments = Array.from(new Set(groupOptions
      .map((g) => (g.department || "").toUpperCase())
      .filter(Boolean))).sort();

    groupDepartmentEl.innerHTML = "";
    if (departments.length > 0) {
      const placeholder = document.createElement("option");
      placeholder.value = "";
      placeholder.textContent = "Выбери кафедру";
      groupDepartmentEl.appendChild(placeholder);
    }
    departments.forEach((dep) => {
      const option = document.createElement("option");
      option.value = dep;
      option.textContent = dep;
      groupDepartmentEl.appendChild(option);
    });

    if (departments.length === 0) {
      const option = document.createElement("option");
      option.value = "";
      option.textContent = "No groups";
      groupDepartmentEl.appendChild(option);
    }
  }

  function fillGroupNumbers() {
    const dep = (groupDepartmentEl.value || "").toUpperCase();
    const numbers = groupOptions
      .filter((g) => (g.department || "").toUpperCase() === dep)
      .map((g) => String(g.number || "").trim())
      .filter(Boolean)
      .sort((a, b) => {
        const na = parseInt(a, 10);
        const nb = parseInt(b, 10);
        if (!Number.isNaN(na) && !Number.isNaN(nb)) return na - nb;
        return a.localeCompare(b);
      });

    groupNumberEl.innerHTML = "";
    if (numbers.length > 0) {
      const placeholder = document.createElement("option");
      placeholder.value = "";
      placeholder.textContent = "Выбери номер группы";
      groupNumberEl.appendChild(placeholder);
    }
    numbers.forEach((num) => {
      const option = document.createElement("option");
      option.value = num;
      option.textContent = num;
      groupNumberEl.appendChild(option);
    });

    if (numbers.length === 0) {
      const option = document.createElement("option");
      option.value = "";
      option.textContent = "No numbers";
      groupNumberEl.appendChild(option);
    }
  }

  function normalizeGroup(raw) {
    const item = typeof raw === "object" && raw ? raw : {};
    const code = String(item.code || "").trim().toUpperCase();

    let department = String(item.department || item.department_code || "").trim().toUpperCase();
    let number = String(item.number || item.group_number || "").trim();

    if ((!department || !number) && code) {
      const parts = code.split("-", 2);
      if (!department && parts[0]) {
        department = parts[0].trim().toUpperCase();
      }
      if (!number && parts[1]) {
        number = parts[1].trim();
      }
    }

    // Fallback: extract numeric tail from code, e.g. "AI45" -> department "AI", number "45".
    if ((!department || !number) && code) {
      const match = code.match(/^([A-ZА-ЯЁ]+)[-_ ]?(\d+)$/i);
      if (match) {
        if (!department) department = match[1].toUpperCase();
        if (!number) number = match[2];
      }
    }

    return {
      id: String(item.id || "").trim(),
      code,
      name: String(item.name || "").trim(),
      department,
      number,
    };
  }

  async function loadGroups() {
    setVisibility(selectedVisibility);

    const { resp, data } = await auth.requestJSON("/v2/projects/groups", {
      method: "GET",
    });

    if (!resp.ok) {
      if (resp.status === 401) {
        throw new Error("Сессия истекла. Войди снова.");
      }
      throw new Error(typeof data === "object" && data && data.error ? data.error : "failed to load groups");
    }

    groupOptions = (Array.isArray(data) ? data : []).map(normalizeGroup);
    fillGroupDepartments();
    fillGroupNumbers();
    setVisibility(selectedVisibility);
  }

  async function loadMineProjects() {
    const started = performance.now();
    const { resp, data } = await auth.requestJSON("/v2/projects/my", {
      method: "GET",
    });

    const elapsed = Math.round(performance.now() - started);

    if (!resp.ok) {
      myProjects = [];
      setStatus(null, `List failed: ${resp.status} (${elapsed} ms)`, false);
      renderMine();
      return;
    }

    myProjects = Array.isArray(data) ? data : [];
    setStatus(null, `Projects loaded: ${myProjects.length} (${elapsed} ms)`, true);
    renderMine();
  }

  async function loadCommunityProjects() {
    const started = performance.now();
    const { resp, data } = await auth.requestJSON("/v2/projects/public", {
      method: "GET",
    });

    const elapsed = Math.round(performance.now() - started);

    if (!resp.ok) {
      publicProjects = [];
      setStatus(null, `Public list failed: ${resp.status} (${elapsed} ms)`, false);
      renderCommunity();
      return;
    }

    publicProjects = Array.isArray(data) ? data : [];
    setStatus(null, `Public projects loaded: ${publicProjects.length} (${elapsed} ms)`, true);
    renderCommunity();
  }

  async function createProject() {
    const title = document.getElementById("title").value.trim();
    const description = document.getElementById("description").value.trim();
    const stacks = parseStacks(document.getElementById("stacks").value);
    if (!title) {
      throw new Error("Название проекта обязательно");
    }

    const payload = {
      title,
      description,
      visibility: selectedVisibility,
    };

    if (selectedVisibility === "PRIVATE") {
      const dep = (groupDepartmentEl.value || "").toUpperCase().trim();
      const num = (groupNumberEl.value || "").trim();
      if (!dep || !num) {
        throw new Error("Выбери кафедру и номер группы для PRIVATE visibility");
      }
      payload.group_code = `${dep}-${num}`;
    }

    const started = performance.now();
    const { resp, data } = await auth.requestJSON("/v2/projects", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const elapsed = Math.round(performance.now() - started);

    if (!resp.ok) {
      setStatus(createStatusEl, `Create failed: ${resp.status} (${elapsed} ms)`, false);
      logLine(typeof data === "object" ? JSON.stringify(data) : String(data));
      return;
    }

    const createdID = typeof data === "object" && data ? String(data.project_id || "") : "";
    if (createdID && stacks.length > 0) {
      const { resp: stacksResp, data: stacksData } = await auth.requestJSON(`/v2/projects/${createdID}/stacks`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ stacks }),
      });
      if (!stacksResp.ok) {
        logLine(`stacks save failed for ${createdID}: ${stacksResp.status} ${JSON.stringify(stacksData)}`);
      }
    }

    setStatus(createStatusEl, `Project created (${elapsed} ms)`, true);
    await loadMineProjects();
    await loadCommunityProjects();
    switchTab("mine");
    closeCreateModal();
  }

  async function syncCapabilities() {
    if (!createToggleBtnEl || !auth || typeof auth.can !== "function") {
      return;
    }

    const canCreate = await auth.can("project.create");
    createToggleBtnEl.hidden = !canCreate;
    createToggleBtnEl.disabled = !canCreate;
    if (!canCreate && !createModalEl.hidden) {
      closeCreateModal();
    }
  }

  document.getElementById("createBtn").addEventListener("click", async () => {
    try {
      await createProject();
    } catch (e) {
      setStatus(createStatusEl, "Create request failed", false);
      logLine(e.message || String(e));
    }
  });

  document.getElementById("refreshActiveBtn").addEventListener("click", async () => {
    if (activeTab === "mine") {
      await loadMineProjects();
      return;
    }
    await loadCommunityProjects();
  });

  createToggleBtnEl.addEventListener("click", openCreateModal);
  createCancelBtnEl.addEventListener("click", closeCreateModal);
  createCloseBtnEl.addEventListener("click", closeCreateModal);
  createModalEl.addEventListener("click", (e) => {
    if (e.target === createModalEl) {
      closeCreateModal();
    }
  });

  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !createModalEl.hidden) {
      closeCreateModal();
    }
  });

  groupDepartmentEl.addEventListener("change", fillGroupNumbers);

  visibilityCards.forEach((card) => {
    card.addEventListener("click", () => {
      setVisibility(card.dataset.visibility || "PUBLIC");
    });
  });

  filterButtons.forEach((btn) => {
    btn.addEventListener("click", () => {
      activeFilter = btn.dataset.filter || "all";
      filterButtons.forEach((b) => b.classList.toggle("active", b === btn));
      if (activeTab === "mine") {
        renderMine();
      }
    });
  });

  viewButtons.forEach((btn) => {
    btn.addEventListener("click", () => {
      activeView = btn.dataset.view === "list" ? "list" : "grid";
      viewButtons.forEach((b) => b.classList.toggle("active", b === btn));
      if (activeTab === "mine") {
        renderMine();
      } else {
        renderCommunity();
      }
    });
  });

  tabButtons.forEach((btn) => {
    btn.addEventListener("click", async (event) => {
      if (btn.tagName === "A") {
        const isModifiedClick = event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button === 1;
        if (isModifiedClick) return;
        event.preventDefault();
      }
      switchTab(btn.dataset.tab || "mine");
      if (activeTab === "community" && publicProjects.length === 0) {
        await loadCommunityProjects();
      }
    });
  });

  searchInputEl.addEventListener("input", () => {
    if (activeTab === "mine") {
      renderMine();
    }
  });

  communitySearchInputEl.addEventListener("input", () => {
    if (activeTab === "community") {
      renderCommunity();
    }
  });

  communityTechFilterEl.addEventListener("change", () => {
    if (activeTab === "community") {
      renderCommunity();
    }
  });

  communityDifficultyFilterEl.addEventListener("change", () => {
    if (activeTab === "community") {
      renderCommunity();
    }
  });

  void (async () => {
    const claims = await auth.ensureSession("student");
    if (!claims) return;

    activeTab = initialTabFromURL();
    bindProfile(claims);
    await syncCapabilities();
    logLine("session initialized");
    loadGroups()
      .then(async () => {
        await loadMineProjects();
        await loadCommunityProjects();
        switchTab(activeTab);
        logLine("all systems operational");
      })
      .catch((e) => {
        setStatus(null, "Initial load failed", false);
        logLine(e.message || String(e));
      });
  })();
})();
