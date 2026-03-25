(() => {
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

  const ROLE_ADMIN = "admin";
  const ROLE_TEACHER = "teacher";

  const ADMIN_THEME_CSS = "/dev/static/css/admin.css";
  const TEACHER_THEME_CSS = "/dev/static/css/professor.css";

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function initials(name, email, fallback) {
    const text = String(name || "").trim();
    if (text) {
      const parts = text.split(/\s+/).filter(Boolean);
      if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
      return text.slice(0, 2).toUpperCase();
    }
    const mail = String(email || "").trim();
    if (mail) return mail.slice(0, 2).toUpperCase();
    return fallback || "ID";
  }

  function localProfile() {
    return {
      full_name: localStorage.getItem(LS_STUDENT_NAME) || "",
      email: localStorage.getItem(LS_STUDENT_EMAIL) || "",
      avatar_url: localStorage.getItem(LS_AVATAR_URL) || "",
      is_admin: localStorage.getItem(LS_IS_ADMIN) === "1",
      is_professor: localStorage.getItem(LS_IS_PROFESSOR) === "1",
    };
  }

  function normalizeProfile(profile) {
    const source = profile && typeof profile === "object" ? profile : localProfile();
    return {
      full_name: String(source.full_name || source.name || ""),
      email: String(source.email || ""),
      avatar_url: String(source.avatar_url || source.avatarURL || ""),
      is_admin: Boolean(source.is_admin),
      is_professor: Boolean(source.is_professor),
    };
  }

  function resolveRole(roleHint, profile) {
    const hint = String(roleHint || "").trim().toLowerCase();
    if (hint === ROLE_ADMIN) return ROLE_ADMIN;
    if (hint === ROLE_TEACHER || hint === "professor") return ROLE_TEACHER;

    if (profile && profile.is_admin) return ROLE_ADMIN;
    if (profile && profile.is_professor) return ROLE_TEACHER;
    return ROLE_TEACHER;
  }

  function maybeSetThemeLink(themeLinkID, role) {
    if (!themeLinkID) return;
    const link = document.getElementById(themeLinkID);
    if (!(link instanceof HTMLLinkElement)) return;

    const next = role === ROLE_ADMIN ? ADMIN_THEME_CSS : TEACHER_THEME_CSS;
    if (link.getAttribute("href") !== next) {
      link.setAttribute("href", next);
    }
  }

  function avatarHTML(avatarURL, fallbackText, size) {
    const url = String(avatarURL || "").trim();
    if (!url) return escapeHTML(fallbackText);
    return `<img src="${escapeHTML(url)}" alt="Avatar" width="${size}" height="${size}" loading="lazy" />`;
  }

  function navItemClass(activeKey, itemKey) {
    return activeKey === itemKey ? "side-link active" : "side-link";
  }

  function adminViewItem(item, activeKey, inlineViews) {
    const className = navItemClass(activeKey, item.key);
    const icon = `<span class="icon-box"><span class="material-symbols-outlined">${item.icon}</span></span>`;
    const label = `<span class="nav-label">${item.label}</span>`;

    if (inlineViews) {
      return `<button class="${className}" data-view="${item.key}" type="button">${icon}${label}</button>`;
    }

    return `<a class="${className}" href="/dev/admin?view=${encodeURIComponent(item.key)}">${icon}${label}</a>`;
  }

  function buildAdminSidebar(activeKey, profile, inlineViews) {
    const name = profile.full_name || profile.email || "Администратор";
    const email = profile.email || "admin@idsai.local";
    const iv = initials(name, email, "AD");

    const viewItems = [
      { key: "dashboard", label: "Дашборд", icon: "dashboard" },
      { key: "projects", label: "Проекты", icon: "folder_data" },
      { key: "users", label: "Пользователи", icon: "group" },
    ];

    const adminNav = viewItems.map((item) => adminViewItem(item, activeKey, inlineViews)).join("");
    const groupsClass = activeKey === "groups" ? "side-link active" : "side-link";
    const settingsClass = activeKey === "settings" ? "side-link ghost active" : "side-link ghost";

    return `
      <div class="brand">
        <div class="brand-icon">
          <img src="/dev/static/assets/idsai-corp-logo.png" alt="IDSAI Corp. logo" width="160" height="40" />
        </div>
        <div class="logo-text">
          <strong>IDSAI Corp.</strong>
          <small>Admin Console</small>
        </div>
      </div>

      <nav class="side-nav" aria-label="Навигация админки">
        ${adminNav}
        <a class="${groupsClass}" href="/dev/groups">
          <span class="icon-box"><span class="material-symbols-outlined">account_tree</span></span>
          <span class="nav-label">Группы</span>
        </a>
        <button class="side-link ghost" type="button" disabled>
          <span class="icon-box"><span class="material-symbols-outlined">schedule</span></span>
          <span class="nav-label">Расписание</span>
        </button>
        <a class="${settingsClass}" href="/dev/settings">
          <span class="icon-box"><span class="material-symbols-outlined">settings</span></span>
          <span class="nav-label">Настройки</span>
        </a>
      </nav>

      <p class="workspace-title">Рабочие пространства</p>
      <div class="workspace-list">
        <span class="workspace-item"><i class="dot violet"></i><span class="user-info-text">Кафедра ИВТ</span></span>
        <span class="workspace-item"><i class="dot amber"></i><span class="user-info-text">Студсовет</span></span>
      </div>

      <div class="sidebar-bottom">
        <div class="admin-chip">
          <div id="sidebarInitials" class="chip-avatar">${avatarHTML(profile.avatar_url, iv, 34)}</div>
          <div class="user-info-text">
            <p id="sidebarName">${escapeHTML(name)}</p>
            <small>SUPER_ADMIN</small>
          </div>
        </div>
        <button id="logoutBtn" class="logout-btn" type="button">
          <span class="icon-box"><span class="material-symbols-outlined">logout</span></span>
          <span class="nav-label">Выйти</span>
        </button>
      </div>
    `;
  }

  function teacherNavItem(label, href, active, key) {
    const className = active === key ? "active" : "";
    return `<a class="${className}" href="${href}">${label}</a>`;
  }

  function buildTeacherSidebar(activeKey, profile) {
    const name = profile.full_name || profile.email || "Преподаватель";
    const email = profile.email || "professor@idsai.dev";
    const iv = initials(name, email, "PR");

    return `
      <div class="prof-brand">
        <img src="/dev/static/assets/idsai-corp-logo.png" alt="IDSAI Corp. logo" width="160" height="40" />
        <div>
          <strong>IDSAI Corp.</strong>
          <p>Преподаватель</p>
        </div>
      </div>

      <div class="prof-user">
        <div id="profAvatar" class="avatar">${avatarHTML(profile.avatar_url, iv, 38)}</div>
        <div>
          <strong id="profName">${escapeHTML(name)}</strong>
          <p id="profEmail">${escapeHTML(email)}</p>
        </div>
      </div>

      <nav class="prof-nav">
        ${teacherNavItem("Дашборд", "/dev/professor", activeKey, "dashboard")}
        ${teacherNavItem("Заявки на ревью", "/dev/professor/reviews", activeKey, "reviews")}
        ${teacherNavItem("Критерии", "/dev/professor/criteria", activeKey, "criteria")}
        ${teacherNavItem("Оценивание", "/dev/professor/grading", activeKey, "grading")}
        ${teacherNavItem("Группы", "/dev/groups", activeKey, "groups")}
        ${teacherNavItem("Проекты", "/dev/projects", activeKey, "projects")}
        ${teacherNavItem("Настройки", "/dev/settings", activeKey, "settings")}
      </nav>

      <footer class="prof-sidebar-foot">
        <button id="logoutBtn" type="button">Выйти</button>
      </footer>
    `;
  }

  function renderSidebar(container, opts = {}) {
    if (!(container instanceof HTMLElement)) return { role: "", active: "" };

    const active = String(opts.active || container.dataset.sidebarActive || "").trim().toLowerCase();
    const roleHint = opts.role || container.dataset.sidebarRole;
    const profile = normalizeProfile(opts.profile);
    const role = resolveRole(roleHint, profile);
    const inlineViews = opts.adminViewMode
      ? String(opts.adminViewMode).toLowerCase() === "inline"
      : String(container.dataset.adminViewMode || "").toLowerCase() === "inline";

    maybeSetThemeLink(container.dataset.sidebarThemeLink, role);
    document.body.classList.toggle("role-admin", role === ROLE_ADMIN);
    document.body.classList.toggle("role-teacher", role !== ROLE_ADMIN);

    if (role === ROLE_ADMIN) {
      container.className = "admin-sidebar";
      container.innerHTML = buildAdminSidebar(active || "dashboard", profile, inlineViews);
      return { role, active: active || "dashboard" };
    }

    container.className = "prof-sidebar";
    container.innerHTML = buildTeacherSidebar(active || "dashboard", profile);
    return { role: ROLE_TEACHER, active: active || "dashboard" };
  }

  function mountFromDOM() {
    const hosts = Array.from(document.querySelectorAll("[data-role-sidebar]"));
    hosts.forEach((host) => {
      renderSidebar(host);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mountFromDOM);
  } else {
    mountFromDOM();
  }

  window.IDSAIRoleSidebar = {
    renderSidebar,
  };
})();
