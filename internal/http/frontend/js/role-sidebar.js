(() => {
  const auth = window.IDSAIAuth;
  const LS_STUDENT_NAME = "idsai_student_name";
  const LS_STUDENT_EMAIL = "idsai_student_email";
  const LS_AVATAR_URL = "idsai_avatar_url";
  const LS_IS_ADMIN = "idsai_is_admin";
  const LS_IS_PROFESSOR = "idsai_is_professor";

  const ROLE_ADMIN = "admin";
  const ROLE_TEACHER = "teacher";
  const ROLE_STUDENT = "student";
  const ROLE_PENDING = "pending";

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

  function hasKnownProfile(profile) {
    if (!profile || typeof profile !== "object") {
      return false;
    }
    return Boolean(
      String(profile.full_name || "").trim() ||
      String(profile.email || "").trim() ||
      String(profile.avatar_url || "").trim() ||
      profile.is_admin ||
      profile.is_professor
    );
  }

  function resolveRole(roleHint, profile) {
    const hint = String(roleHint || "").trim().toLowerCase();
    if (hint === ROLE_ADMIN) return ROLE_ADMIN;
    if (hint === ROLE_TEACHER || hint === "professor") return ROLE_TEACHER;
    if (hint === ROLE_STUDENT) return ROLE_STUDENT;
    if (!hasKnownProfile(profile)) return ROLE_PENDING;

    if (profile && profile.is_admin) return ROLE_ADMIN;
    if (profile && profile.is_professor) return ROLE_TEACHER;
    return ROLE_STUDENT;
  }

  function avatarHTML(avatarURL, fallbackText, size) {
    const url = String(avatarURL || "").trim();
    if (!url) return escapeHTML(fallbackText);
    return `<img src="${escapeHTML(url)}" alt="Avatar" width="${size}" height="${size}" loading="lazy" />`;
  }

  function navItemClass(activeKey, itemKey, extraClass = "") {
    const parts = ["role-sidebar__link"];
    if (activeKey === itemKey) parts.push("active");
    if (extraClass) parts.push(extraClass);
    return parts.join(" ");
  }

  function navIcon(icon) {
    if (icon === "dashboard") {
      return `
        <span class="role-sidebar__icon" aria-hidden="true">
          <svg class="role-sidebar__icon-svg" viewBox="0 0 24 24" fill="none" focusable="false">
            <rect x="3.5" y="3.5" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.7"></rect>
            <rect x="13.5" y="3.5" width="7" height="11" rx="1.5" stroke="currentColor" stroke-width="1.7"></rect>
            <rect x="3.5" y="13.5" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.7"></rect>
            <rect x="13.5" y="17.5" width="7" height="3" rx="1.5" stroke="currentColor" stroke-width="1.7"></rect>
          </svg>
        </span>
      `;
    }

    return `
      <span class="role-sidebar__icon" aria-hidden="true">
        <span class="material-symbols-outlined">${icon}</span>
      </span>
    `;
  }

  function actionLink(item, activeKey) {
    return `
      <a class="${navItemClass(activeKey, item.key)}" href="${item.href}" data-sidebar-item="${item.key}">
        ${navIcon(item.icon)}
        <span class="role-sidebar__label">${item.label}</span>
      </a>
    `;
  }

  function actionButton(item, activeKey) {
    return `
      <button
        class="${navItemClass(activeKey, item.key)}"
        data-view="${item.key}"
        data-sidebar-view="${item.key}"
        data-sidebar-item="${item.key}"
        type="button"
      >
        ${navIcon(item.icon)}
        <span class="role-sidebar__label">${item.label}</span>
      </button>
    `;
  }

  function allowedNavItems(navItems, scope, skipPermissionFilter = false) {
    if (skipPermissionFilter || !auth || typeof auth.canCached !== "function") {
      return navItems;
    }
    return navItems.filter((item) => !item.permission || auth.canCached(item.permission, scope));
  }

  function bindSidebarLogout(container) {
    if (!(container instanceof HTMLElement)) {
      return;
    }

    const logoutBtn = container.querySelector("#logoutBtn");
    if (!(logoutBtn instanceof HTMLElement) || logoutBtn.dataset.bound === "1") {
      return;
    }

    logoutBtn.dataset.bound = "1";
    logoutBtn.addEventListener("click", (event) => {
      event.preventDefault();
      if (auth && typeof auth.logout === "function") {
        void auth.logout();
        return;
      }
      window.location.href = "/dev/login";
    });
  }

  function sidebarFrame(role, activeKey, profile, navItems, inlineViews, scope, skipPermissionFilter) {
    const isAdmin = role === ROLE_ADMIN;
    const isTeacher = role === ROLE_TEACHER;
    const name = profile.full_name || profile.email || (isAdmin ? "Администратор" : isTeacher ? "Преподаватель" : "Студент");
    const email = profile.email || (isAdmin ? "admin@idsai.local" : isTeacher ? "professor@idsai.dev" : "student@idsai.dev");
    const iv = initials(name, email, isAdmin ? "AD" : isTeacher ? "PR" : "ST");
    const navHTML = allowedNavItems(navItems, scope, skipPermissionFilter)
      .map((item) => (inlineViews && isAdmin && item.inline !== false ? actionButton(item, activeKey) : actionLink(item, activeKey)))
      .join("");

    const settingsClass = navItemClass(activeKey, "settings");

    return `
      <a class="role-sidebar__brand" href="${isAdmin ? "/dev/admin" : isTeacher ? "/dev/professor" : "/dev/projects"}" aria-label="IDSAI Corp.">
        <span class="role-sidebar__brand-mark">
          <img src="/dev/static/assets/idsai-corp-logo.png" alt="IDSAI Corp. logo" width="42" height="42" />
        </span>
        <span class="role-sidebar__brand-text">
          <strong>IDSAI Corp.</strong>
          <small>${isAdmin ? "Admin Console" : isTeacher ? "Professor Workspace" : "Student Workspace"}</small>
        </span>
      </a>

      <nav class="role-sidebar__nav" aria-label="${isAdmin ? "Навигация администратора" : isTeacher ? "Навигация преподавателя" : "Навигация студента"}">
        ${navHTML}
      </nav>

      <div class="role-sidebar__bottom">
        <div class="role-sidebar__actions">
          <a class="${settingsClass}" href="/dev/settings">
            ${navIcon("settings")}
            <span class="role-sidebar__label">Настройки</span>
          </a>
          <button id="logoutBtn" class="role-sidebar__link role-sidebar__logout" type="button">
            ${navIcon("logout")}
            <span class="role-sidebar__label">Выйти</span>
          </button>
        </div>

        <div class="role-sidebar__account">
          <div id="sidebarInitials" class="role-sidebar__avatar">${avatarHTML(profile.avatar_url, iv, 40)}</div>
          <div class="role-sidebar__account-text">
            <strong id="sidebarName">${escapeHTML(name)}</strong>
            <small>${isAdmin ? "SUPER_ADMIN" : isTeacher ? "ПРЕПОДАВАТЕЛЬ" : "СТУДЕНТ"}</small>
          </div>
        </div>
      </div>
    `;
  }

  function buildAdminSidebar(activeKey, profile, inlineViews, scope, skipPermissionFilter) {
    return sidebarFrame(
      ROLE_ADMIN,
      activeKey,
      profile,
      [
        { key: "profile", label: "Профиль", icon: "person", href: "/dev/profile", inline: false },
        { key: "dashboard", label: "Дашборд", icon: "dashboard", href: "/dev/admin?view=dashboard", permission: "admin.manage_rbac" },
        { key: "projects", label: "Проекты", icon: "folder_data", href: "/dev/admin?view=projects", permission: "admin.manage_rbac" },
        { key: "users", label: "Пользователи", icon: "group", href: "/dev/admin?view=users", permission: "admin.manage_rbac" },
        { key: "groups", label: "Группы", icon: "account_tree", href: "/dev/groups", inline: false, permission: "admin.manage_rbac" },
        { key: "kb", label: "База знаний", icon: "menu_book", href: "/dev/kb", inline: false },
      ],
      inlineViews,
      scope,
      skipPermissionFilter
    );
  }

  function buildTeacherSidebar(activeKey, profile, scope, skipPermissionFilter) {
    return sidebarFrame(
      ROLE_TEACHER,
      activeKey,
      profile,
      [
        { key: "profile", label: "Профиль", icon: "person", href: "/dev/profile" },
        { key: "dashboard", label: "Дашборд", icon: "dashboard", href: "/dev/professor" },
        { key: "reviews", label: "Заявки на ревью", icon: "fact_check", href: "/dev/professor/reviews" },
        { key: "criteria", label: "Критерии", icon: "checklist_rtl", href: "/dev/professor/criteria" },
        { key: "grading", label: "Оценивание", icon: "grading", href: "/dev/professor/grading" },
        { key: "groups", label: "Группы", icon: "account_tree", href: "/dev/groups" },
        { key: "projects", label: "Проекты", icon: "folder_open", href: "/dev/professor#projects" },
        { key: "kb", label: "База знаний", icon: "menu_book", href: "/dev/kb" },
      ],
      false,
      scope,
      skipPermissionFilter
    );
  }

  function buildStudentSidebar(activeKey, profile, scope, skipPermissionFilter) {
    return sidebarFrame(
      ROLE_STUDENT,
      activeKey,
      profile,
      [
        { key: "profile", label: "Профиль", icon: "person", href: "/dev/profile" },
        { key: "mine", label: "Мои проекты", icon: "folder", href: "/dev/projects?tab=mine" },
        { key: "community", label: "Сообщество", icon: "public", href: "/dev/projects?tab=community" },
        { key: "invites", label: "Заявки", icon: "mark_email_unread", href: "/dev/invites" },
        { key: "kb", label: "База знаний", icon: "menu_book", href: "/dev/kb" },
      ],
      false,
      scope,
      skipPermissionFilter
    );
  }

  function buildPendingSidebar() {
    return `
      <div class="role-sidebar__brand role-sidebar__brand--static" aria-hidden="true">
        <span class="role-sidebar__brand-mark">
          <img src="/dev/static/assets/idsai-corp-logo.png" alt="" width="42" height="42" />
        </span>
        <span class="role-sidebar__brand-text">
          <strong>IDSAI Corp.</strong>
          <small>Workspace</small>
        </span>
      </div>

      <div class="role-sidebar__skeleton-block" aria-hidden="true">
        <span class="role-sidebar__skeleton-line role-sidebar__skeleton-line--lg"></span>
        <span class="role-sidebar__skeleton-line"></span>
        <span class="role-sidebar__skeleton-line"></span>
        <span class="role-sidebar__skeleton-line"></span>
      </div>

      <div class="role-sidebar__bottom" aria-hidden="true">
        <div class="role-sidebar__skeleton-block role-sidebar__skeleton-block--compact">
          <span class="role-sidebar__skeleton-line"></span>
          <span class="role-sidebar__skeleton-line"></span>
        </div>
        <div class="role-sidebar__account role-sidebar__account--pending">
          <span class="role-sidebar__skeleton-avatar"></span>
          <div class="role-sidebar__account-text">
            <span class="role-sidebar__skeleton-line role-sidebar__skeleton-line--sm"></span>
            <span class="role-sidebar__skeleton-line role-sidebar__skeleton-line--xs"></span>
          </div>
        </div>
      </div>
    `;
  }

  function renderSidebar(container, opts = {}) {
    if (!(container instanceof HTMLElement)) return { role: "", active: "" };

    let active = String(opts.active || container.dataset.sidebarActive || "").trim().toLowerCase();
    const roleHint = opts.role || container.dataset.sidebarRole;
    const profile = normalizeProfile(opts.profile);
    const skipPermissionFilter = Boolean(opts.skipPermissionFilter);
    const role = resolveRole(roleHint, profile);
    const scope = opts.scope || (auth && typeof auth.getDefaultScope === "function" ? auth.getDefaultScope() : null);
    const inlineViews = opts.adminViewMode
      ? String(opts.adminViewMode).toLowerCase() === "inline"
      : String(container.dataset.adminViewMode || "").toLowerCase() === "inline";

    if (role === ROLE_TEACHER) {
      const path = String(window.location.pathname || "").toLowerCase();
      const hash = String(window.location.hash || "").toLowerCase();
      if (path === "/dev/professor" && hash === "#projects") {
        active = "projects";
      }
    }

    document.body.classList.toggle("role-admin", role === ROLE_ADMIN);
    document.body.classList.toggle("role-teacher", role === ROLE_TEACHER);
    document.body.classList.toggle("role-student", role === ROLE_STUDENT);

    if (role === ROLE_PENDING) {
      container.className = "role-sidebar role-sidebar--pending";
      container.innerHTML = buildPendingSidebar();
      return { role, active: "" };
    }

    if (role === ROLE_ADMIN) {
      container.className = "role-sidebar role-sidebar--admin";
      container.innerHTML = buildAdminSidebar(active || "dashboard", profile, inlineViews, scope, skipPermissionFilter);
      bindSidebarLogout(container);
      return { role, active: active || "dashboard" };
    }

    if (role === ROLE_TEACHER) {
      container.className = "role-sidebar role-sidebar--teacher";
      container.innerHTML = buildTeacherSidebar(active || "dashboard", profile, scope, skipPermissionFilter);
      bindSidebarLogout(container);
      return { role, active: active || "dashboard" };
    }

    container.className = "role-sidebar role-sidebar--student";
    container.innerHTML = buildStudentSidebar(active || "profile", profile, scope, skipPermissionFilter);
    bindSidebarLogout(container);

    return { role, active: active || "profile" };
  }

  function mountFromDOM() {
    const hosts = Array.from(document.querySelectorAll("[data-role-sidebar]"));
    if (!hosts.length) {
      return;
    }

    const cached = localProfile();
    hosts.forEach((host) => {
      renderSidebar(host, { profile: cached, skipPermissionFilter: true });
    });
    syncTeacherHashState();

    if (!auth || typeof auth.fetchCurrentProfile !== "function") {
      return;
    }

    void (async () => {
      try {
        const profile = await auth.fetchCurrentProfile();
        if (profile && typeof auth.loadCapabilities === "function") {
          await auth.loadCapabilities();
        }
        hosts.forEach((host) => {
          renderSidebar(host, { profile: profile || cached, skipPermissionFilter: !profile });
        });
      } catch (_) {
        hosts.forEach((host) => {
          renderSidebar(host, { profile: cached, skipPermissionFilter: true });
        });
      }
      syncTeacherHashState();
    })();
  }

  function syncTeacherHashState() {
    const path = String(window.location.pathname || "").toLowerCase();
    const hash = String(window.location.hash || "").toLowerCase();
    const nextActive = path === "/dev/professor" && hash === "#projects" ? "projects" : "dashboard";

    const hosts = Array.from(document.querySelectorAll(".role-sidebar--teacher[data-role-sidebar]"));
    hosts.forEach((host) => {
      const baseActive = String(host.dataset.sidebarActive || "").trim().toLowerCase();
      if (path !== "/dev/professor" || (baseActive !== "dashboard" && baseActive !== "projects")) return;

      const items = Array.from(host.querySelectorAll(".role-sidebar__link[data-sidebar-item]"));
      items.forEach((item) => {
        item.classList.toggle("active", item.dataset.sidebarItem === nextActive);
      });
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mountFromDOM);
  } else {
    mountFromDOM();
  }

  window.addEventListener("hashchange", syncTeacherHashState);

  window.IDSAIRoleSidebar = {
    renderSidebar,
  };
})();
