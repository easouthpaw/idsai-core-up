(() => {
  const auth = window.IDSAIAuth;
  const roleSidebar = window.IDSAIRoleSidebar;

  const ui = {
    sidebarHost: document.querySelector("[data-role-sidebar]"),
    logoutBtn: null,

    departmentFilter: document.getElementById("departmentFilter"),
    searchInput: document.getElementById("searchInput"),
    refreshBtn: document.getElementById("refreshBtn"),
    pageStatus: document.getElementById("pageStatus"),
    treeRoot: document.getElementById("treeRoot"),

    adminRequestsSection: document.getElementById("adminRequestsSection"),
    requestStatusFilter: document.getElementById("requestStatusFilter"),
    refreshRequestsBtn: document.getElementById("refreshRequestsBtn"),
    requestsList: document.getElementById("requestsList"),
  };

  const state = {
    profile: null,
    departments: [],
    tree: [],
    requests: [],
    searchTimer: null,
    isAdmin: false,
    isProfessor: false,
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
    return e ? e.slice(0, 2).toUpperCase() : "GR";
  }

  function setStatus(message, isError) {
    ui.pageStatus.textContent = message || "";
    ui.pageStatus.classList.toggle("err", Boolean(isError));
  }

  async function requestJSON(url, options = {}) {
    const { resp, data } = await auth.requestJSON(url, options);
    if (!resp.ok) {
      const err = new Error(data && data.error ? data.error : `Request failed (${resp.status})`);
      err.status = resp.status;
      throw err;
    }
    return data;
  }

  function syncSidebar(profile) {
    const isAdmin = Boolean(profile && profile.is_admin);
    document.body.classList.toggle("role-admin", isAdmin);
    document.body.classList.toggle("role-teacher", !isAdmin);

    if (roleSidebar && typeof roleSidebar.renderSidebar === "function" && ui.sidebarHost) {
      roleSidebar.renderSidebar(ui.sidebarHost, {
        profile,
        role: isAdmin ? "admin" : "teacher",
        active: "groups",
        adminViewMode: "links",
      });
    }

    ui.logoutBtn = document.getElementById("logoutBtn");
  }

  function setDepartmentFilterOptions(items) {
    const list = Array.isArray(items) ? items : [];
    ui.departmentFilter.innerHTML = "";
    const first = document.createElement("option");
    first.value = "";
    first.textContent = "Все кафедры";
    ui.departmentFilter.appendChild(first);

    list.forEach((item) => {
      const code = String(item.code || "").toUpperCase();
      const name = String(item.name || "");
      const option = document.createElement("option");
      option.value = code;
      option.textContent = name ? `${code} — ${name}` : code;
      ui.departmentFilter.appendChild(option);
    });
  }

  function roleLabel(roleCode) {
    const code = String(roleCode || "").toUpperCase();
    if (code === "SUPER_ADMIN") return "Админ";
    if (code === "PROFESSOR") return "Преподаватель";
    if (code === "STUDENT") return "Студент";
    return code || "—";
  }

  function renderTree(items) {
    const departments = Array.isArray(items) ? items : [];
    ui.treeRoot.innerHTML = "";

    if (!departments.length) {
      ui.treeRoot.innerHTML = "<p>По выбранным фильтрам данных нет.</p>";
      return;
    }

    departments.forEach((department) => {
      const depDetails = document.createElement("details");
      depDetails.className = "dep-node";
      depDetails.open = true;

      const depSummary = document.createElement("summary");
      depSummary.textContent = String(department.name || department.code || "Кафедра");
      depDetails.appendChild(depSummary);

      const groupsList = document.createElement("div");
      groupsList.className = "groups-list";

      const groups = Array.isArray(department.groups) ? department.groups : [];
      if (!groups.length) {
        const empty = document.createElement("p");
        empty.textContent = "Группы не найдены";
        groupsList.appendChild(empty);
      }

      groups.forEach((group) => {
        const groupDetails = document.createElement("details");
        groupDetails.className = "group-node";

        const summary = document.createElement("summary");
        summary.innerHTML = `
          <strong>${escapeHTML(group.group_code || "—")}</strong>
          <span>Студентов: ${escapeHTML(group.total_students)}</span>
        `;
        groupDetails.appendChild(summary);

        const studentsList = document.createElement("ul");
        studentsList.className = "students-list";

        const students = Array.isArray(group.students) ? group.students : [];
        if (!students.length) {
          const empty = document.createElement("li");
          empty.textContent = "В этой группе пока нет студентов";
          studentsList.appendChild(empty);
        }

        students.forEach((student) => {
          const li = document.createElement("li");
          li.className = "student-row";

          const avatarURL = String(student.avatar_url || "").trim();
          const avatarFallback = initials(student.full_name, student.email);
          const avatar = avatarURL
            ? `<img src="${escapeHTML(avatarURL)}" alt="Avatar" width="44" height="44" loading="lazy" />`
            : escapeHTML(avatarFallback);

          li.innerHTML = `
            <div class="student-avatar">${avatar}</div>
            <div class="student-meta">
              <strong>${escapeHTML(student.full_name || "—")}</strong>
              <p>${escapeHTML(student.email || "—")}</p>
            </div>
            <div>${escapeHTML(roleLabel(student.role || student.role_code))}</div>
          `;
          studentsList.appendChild(li);
        });

        groupDetails.appendChild(studentsList);
        groupsList.appendChild(groupDetails);
      });

      depDetails.appendChild(groupsList);
      ui.treeRoot.appendChild(depDetails);
    });
  }

  function requestStatusLabel(status) {
    const s = String(status || "").toUpperCase();
    if (s === "PENDING") return "Ожидает";
    if (s === "APPROVED") return "Одобрено";
    if (s === "REJECTED") return "Отклонено";
    return s || "—";
  }

  function formatDateTime(value) {
    if (!value) return "—";
    const dt = new Date(value);
    if (Number.isNaN(dt.getTime())) return "—";
    return dt.toLocaleString("ru-RU");
  }

  function renderRequests(items) {
    if (!state.isAdmin) return;

    const list = Array.isArray(items) ? items : [];
    ui.requestsList.innerHTML = "";

    if (!list.length) {
      ui.requestsList.innerHTML = "<p>Заявок нет.</p>";
      return;
    }

    list.forEach((item) => {
      const card = document.createElement("article");
      card.className = "request-item";

      const id = String(item.id || "");
      const status = String(item.status || "").toUpperCase();
      const readonly = status !== "PENDING";

      card.innerHTML = `
        <strong>${escapeHTML(item.student_name || item.student_email || "Student")}</strong>
        <div>${escapeHTML(item.current_group_code || "—")} → ${escapeHTML(item.requested_group_code || "—")}</div>
        <div class="meta">Статус: ${escapeHTML(requestStatusLabel(status))}</div>
        <div class="meta">Создано: ${escapeHTML(formatDateTime(item.created_at))}</div>
        <div class="meta">Проверено: ${escapeHTML(formatDateTime(item.reviewed_at))}</div>
        <textarea data-comment="${escapeHTML(id)}" placeholder="Комментарий администратора" ${readonly ? "disabled" : ""}>${escapeHTML(item.admin_comment || "")}</textarea>
        <div class="actions">
          <button class="approve" data-action="approve" data-id="${escapeHTML(id)}" ${readonly ? "disabled" : ""}>Одобрить</button>
          <button class="reject" data-action="reject" data-id="${escapeHTML(id)}" ${readonly ? "disabled" : ""}>Отклонить</button>
        </div>
      `;

      ui.requestsList.appendChild(card);
    });
  }

  async function loadDepartments() {
    const data = await requestJSON("/v2/auth/departments", { method: "GET" });
    state.departments = Array.isArray(data.departments) ? data.departments : [];
    setDepartmentFilterOptions(state.departments);
  }

  async function loadTree() {
    const params = new URLSearchParams();
    const departmentCode = String(ui.departmentFilter.value || "").trim().toUpperCase();
    const search = String(ui.searchInput.value || "").trim();
    if (departmentCode) params.set("department_code", departmentCode);
    if (search) params.set("q", search);

    setStatus("Загрузка структуры групп...", false);
    const data = await requestJSON(`/v2/auth/groups/tree?${params.toString()}`, { method: "GET" });
    const tree = Array.isArray(data.departments) ? data.departments : [];
    state.tree = tree;
    renderTree(tree);
    setStatus(`Кафедр: ${tree.length}`, false);
  }

  async function loadAdminRequests() {
    if (!state.isAdmin) return;
    const params = new URLSearchParams();
    const status = String(ui.requestStatusFilter.value || "").trim().toUpperCase();
    if (status) params.set("status", status);

    const data = await requestJSON(`/v2/auth/admin/group-change-requests?${params.toString()}`, { method: "GET" });
    const requests = Array.isArray(data.requests) ? data.requests : [];
    state.requests = requests;
    renderRequests(requests);
  }

  async function reviewRequest(requestID, action) {
    const commentEl = Array.from(ui.requestsList.querySelectorAll("textarea[data-comment]"))
      .find((el) => String(el.dataset.comment || "") === requestID);
    const comment = commentEl ? String(commentEl.value || "").trim() : "";

    await requestJSON(`/v2/auth/admin/group-change-requests/${encodeURIComponent(requestID)}/review`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action, comment }),
    });

    await Promise.all([loadAdminRequests(), loadTree()]);
    setStatus("Заявка обновлена.", false);
  }

  function wireEvents() {
    if (ui.logoutBtn) {
      ui.logoutBtn.addEventListener("click", () => auth.logout());
    }

    ui.refreshBtn.addEventListener("click", () => {
      loadTree().catch((err) => setStatus(err.message || String(err), true));
    });

    ui.departmentFilter.addEventListener("change", () => {
      loadTree().catch((err) => setStatus(err.message || String(err), true));
    });

    ui.searchInput.addEventListener("input", () => {
      if (state.searchTimer) clearTimeout(state.searchTimer);
      state.searchTimer = setTimeout(() => {
        loadTree().catch((err) => setStatus(err.message || String(err), true));
      }, 250);
    });

    if (state.isAdmin) {
      ui.refreshRequestsBtn.addEventListener("click", () => {
        loadAdminRequests().catch((err) => setStatus(err.message || String(err), true));
      });

      ui.requestStatusFilter.addEventListener("change", () => {
        loadAdminRequests().catch((err) => setStatus(err.message || String(err), true));
      });

      ui.requestsList.addEventListener("click", (event) => {
        const button = event.target.closest("button[data-action][data-id]");
        if (!button) return;
        const requestID = String(button.dataset.id || "");
        const action = String(button.dataset.action || "").toLowerCase();
        if (!requestID || !action) return;
        reviewRequest(requestID, action).catch((err) => setStatus(err.message || String(err), true));
      });
    }
  }

  async function bootstrap() {
    const profile = await auth.ensureSession(undefined);
    if (!profile) return;

    state.profile = profile;
    state.isAdmin = Boolean(profile.is_admin);
    state.isProfessor = Boolean(profile.is_professor);

    if (!state.isAdmin && !state.isProfessor) {
      window.location.href = "/dev/projects";
      return;
    }

    syncSidebar(profile);

    if (state.isAdmin) {
      ui.adminRequestsSection.hidden = false;
    }

    wireEvents();

    const tasks = [loadDepartments(), loadTree()];
    if (state.isAdmin) {
      tasks.push(loadAdminRequests());
    }
    await Promise.all(tasks);
  }

  bootstrap().catch((err) => {
    setStatus(err.message || String(err), true);
  });
})();
