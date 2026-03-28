(() => {
  "use strict";

  const auth = window.IDSAIAuth;
  const API = "/v2/kb";
  const PAGE_SIZE = 20;

  const state = {
    isEditor: false,
    categories: [],
    activeCategoryId: null,
    articles: [],
    tags: [],
    activeTag: null,
    search: "",
    page: 0,
    total: 0,
    editingCategoryId: null,
  };

  const ui = {
    categoryTree: document.getElementById("categoryTree"),
    categoryActions: document.getElementById("categoryActions"),
    addCategoryBtn: document.getElementById("addCategoryBtn"),
    breadcrumbCategory: document.getElementById("breadcrumbCategory"),
    searchInput: document.getElementById("searchInput"),
    tagsBar: document.getElementById("tagsBar"),
    editorActions: document.getElementById("editorActions"),
    createArticleBtn: document.getElementById("createArticleBtn"),
    uploadMdFile: document.getElementById("uploadMdFile"),
    articleList: document.getElementById("articleList"),
    emptyState: document.getElementById("emptyState"),
    pagination: document.getElementById("pagination"),
    prevPage: document.getElementById("prevPage"),
    nextPage: document.getElementById("nextPage"),
    pageInfo: document.getElementById("pageInfo"),
    // Category modal
    categoryModal: document.getElementById("categoryModal"),
    categoryModalTitle: document.getElementById("categoryModalTitle"),
    catTitleInput: document.getElementById("catTitleInput"),
    catParentSelect: document.getElementById("catParentSelect"),
    cancelCatBtn: document.getElementById("cancelCatBtn"),
    saveCatBtn: document.getElementById("saveCatBtn"),
    // Article modal
    articleModal: document.getElementById("articleModal"),
    artTitleInput: document.getElementById("artTitleInput"),
    artCategorySelect: document.getElementById("artCategorySelect"),
    artTagsInput: document.getElementById("artTagsInput"),
    artStatusSelect: document.getElementById("artStatusSelect"),
    artContentInput: document.getElementById("artContentInput"),
    artPreview: document.getElementById("artPreview"),
    cancelArtBtn: document.getElementById("cancelArtBtn"),
    saveArtBtn: document.getElementById("saveArtBtn"),
  };

  // ── Helpers ───────────────────────────────────────────────

  function escapeHTML(str) {
    return String(str || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;");
  }

  function relativeTime(dateStr) {
    if (!dateStr) return "";
    const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000);
    if (diff < 60) return "только что";
    if (diff < 3600) return `${Math.floor(diff / 60)} мин. назад`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} ч. назад`;
    if (diff < 604800) return `${Math.floor(diff / 86400)} дн. назад`;
    return new Date(dateStr).toLocaleDateString("ru-RU", { day: "numeric", month: "short", year: "numeric" });
  }

  async function api(path, opts = {}) {
    const token = auth?.getToken?.();
    const headers = { ...(opts.headers || {}) };
    if (token) headers["Authorization"] = `Bearer ${token}`;
    if (opts.body && typeof opts.body === "object" && !(opts.body instanceof FormData)) {
      headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(opts.body);
    }
    const res = await fetch(API + path, { ...opts, headers });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    if (res.status === 204) return null;
    return res.json();
  }

  // ── Category Tree ─────────────────────────────────────────

  function buildTree(categories) {
    const map = new Map();
    const roots = [];
    categories.forEach((c) => {
      map.set(c.id, { ...c, children: [] });
    });
    categories.forEach((c) => {
      const node = map.get(c.id);
      if (c.parent_id && map.has(c.parent_id)) {
        map.get(c.parent_id).children.push(node);
      } else {
        roots.push(node);
      }
    });
    return roots;
  }

  function renderTreeNode(node, depth = 0) {
    const hasChildren = node.children && node.children.length > 0;
    const isActive = state.activeCategoryId === node.id;

    let html = `<div class="kb-tree-item ${isActive ? "is-active" : ""}" data-category-id="${node.id}" title="${escapeHTML(node.title)}">`;
    if (hasChildren) {
      html += `<button class="kb-tree-toggle is-expanded" data-toggle="${node.id}" type="button"><span class="material-symbols-outlined">chevron_right</span></button>`;
    } else {
      html += `<span class="material-symbols-outlined" style="font-size:16px;">description</span>`;
    }
    html += `<span style="flex:1;">${escapeHTML(node.title)}</span>`;
    html += `</div>`;

    if (hasChildren) {
      html += `<div class="kb-tree-children" data-children="${node.id}">`;
      node.children.forEach((child) => {
        html += renderTreeNode(child, depth + 1);
      });
      html += `</div>`;
    }

    return html;
  }

  function renderCategoryTree() {
    const tree = buildTree(state.categories);
    if (tree.length === 0) {
      ui.categoryTree.innerHTML = `<div style="color:var(--kb-muted);font-size:13px;padding:8px 0;">Нет категорий</div>`;
      return;
    }

    // "All" item
    let html = `<div class="kb-tree-item ${!state.activeCategoryId ? "is-active" : ""}" data-category-id="">
      <span class="material-symbols-outlined" style="font-size:16px;">home</span>
      <span style="flex:1;">Все статьи</span>
    </div>`;

    tree.forEach((node) => {
      html += renderTreeNode(node);
    });
    ui.categoryTree.innerHTML = html;
  }

  function populateCategorySelects() {
    const makeOptions = (selected) => {
      let html = `<option value="">— Корень —</option>`;
      state.categories.forEach((c) => {
        html += `<option value="${c.id}" ${c.id === selected ? "selected" : ""}>${escapeHTML(c.title)}</option>`;
      });
      return html;
    };
    ui.catParentSelect.innerHTML = makeOptions("");
    ui.artCategorySelect.innerHTML = state.categories.map((c) => `<option value="${c.id}">${escapeHTML(c.title)}</option>`).join("");
  }

  // ── Tags ──────────────────────────────────────────────────

  function renderTags() {
    if (!state.tags || state.tags.length === 0) {
      ui.tagsBar.innerHTML = "";
      return;
    }
    ui.tagsBar.innerHTML = state.tags
      .map((t) => `<button class="kb-tag ${state.activeTag === t.name ? "is-active" : ""}" data-tag="${escapeHTML(t.name)}" type="button">${escapeHTML(t.name)}</button>`)
      .join("");
  }

  // ── Articles ──────────────────────────────────────────────

  function renderArticles() {
    if (!state.articles || state.articles.length === 0) {
      ui.articleList.innerHTML = "";
      ui.emptyState.style.display = "";
      ui.pagination.style.display = "none";
      return;
    }
    ui.emptyState.style.display = "none";

    ui.articleList.innerHTML = state.articles
      .map((a) => {
        const isDraft = a.status === "DRAFT";
        const tagsHTML = (a.tags || []).map((t) => `<span class="kb-article-card__tag">${escapeHTML(t)}</span>`).join("");
        const statusHTML = isDraft
          ? `<span class="kb-article-card__status kb-article-card__status--draft">Черновик</span>`
          : `<span class="kb-article-card__status kb-article-card__status--published">Опубликовано</span>`;

        return `
          <a class="kb-article-card ${isDraft ? "kb-article-card--draft" : ""}" href="/dev/kb/article?id=${a.id}" data-article-id="${a.id}">
            <div style="display:flex;align-items:center;gap:10px;">
              <h3 class="kb-article-card__title" style="flex:1;">${escapeHTML(a.title)}</h3>
              ${statusHTML}
            </div>
            <div class="kb-article-card__meta">
              <span class="material-symbols-outlined">person</span>
              ${escapeHTML(a.author_name)}
              <span>·</span>
              <span>${relativeTime(a.updated_at)}</span>
            </div>
            ${tagsHTML ? `<div class="kb-article-card__tags">${tagsHTML}</div>` : ""}
          </a>`;
      })
      .join("");

    // Pagination
    const totalPages = Math.ceil(state.total / PAGE_SIZE);
    if (totalPages > 1) {
      ui.pagination.style.display = "";
      ui.pageInfo.textContent = `${state.page + 1} / ${totalPages}`;
      ui.prevPage.disabled = state.page === 0;
      ui.nextPage.disabled = state.page >= totalPages - 1;
    } else {
      ui.pagination.style.display = "none";
    }
  }

  // ── Data Loading ──────────────────────────────────────────

  async function loadCategories() {
    try {
      state.categories = await api("/categories");
      if (!Array.isArray(state.categories)) state.categories = [];
    } catch {
      state.categories = [];
    }
    renderCategoryTree();
    populateCategorySelects();
  }

  async function loadTags() {
    try {
      state.tags = await api("/tags");
      if (!Array.isArray(state.tags)) state.tags = [];
    } catch {
      state.tags = [];
    }
    renderTags();
  }

  async function loadArticles() {
    try {
      const params = new URLSearchParams();
      if (state.activeCategoryId) params.set("category_id", state.activeCategoryId);
      if (state.search) params.set("search", state.search);
      if (state.activeTag) params.set("tag", state.activeTag);
      params.set("limit", PAGE_SIZE);
      params.set("offset", state.page * PAGE_SIZE);

      const data = await api("/articles?" + params.toString());
      state.articles = data.items || [];
      state.total = data.total || 0;
    } catch {
      state.articles = [];
      state.total = 0;
    }
    renderArticles();
  }

  // ── Category CRUD ─────────────────────────────────────────

  function openCategoryModal(editCat = null) {
    state.editingCategoryId = editCat?.id || null;
    ui.categoryModalTitle.textContent = editCat ? "Редактировать категорию" : "Новая категория";
    ui.catTitleInput.value = editCat?.title || "";
    populateCategorySelects();
    if (editCat?.parent_id) {
      ui.catParentSelect.value = editCat.parent_id;
    } else {
      ui.catParentSelect.value = "";
    }
    ui.categoryModal.style.display = "";
    ui.catTitleInput.focus();
  }

  function closeCategoryModal() {
    ui.categoryModal.style.display = "none";
    state.editingCategoryId = null;
  }

  async function saveCategory() {
    const title = ui.catTitleInput.value.trim();
    if (!title) return;
    const parentId = ui.catParentSelect.value || null;

    try {
      if (state.editingCategoryId) {
        await api(`/categories/${state.editingCategoryId}`, {
          method: "PATCH",
          body: { title, sort_order: 0 },
        });
      } else {
        await api("/categories", {
          method: "POST",
          body: { title, parent_id: parentId, sort_order: 0 },
        });
      }
      closeCategoryModal();
      await loadCategories();
    } catch (err) {
      alert("Ошибка: " + err.message);
    }
  }

  // ── Article Create ────────────────────────────────────────

  function openArticleModal() {
    ui.artTitleInput.value = "";
    ui.artContentInput.value = "";
    ui.artTagsInput.value = "";
    ui.artStatusSelect.value = "DRAFT";
    populateCategorySelects();
    if (state.activeCategoryId) {
      ui.artCategorySelect.value = state.activeCategoryId;
    }
    ui.articleModal.style.display = "";
    ui.artTitleInput.focus();
  }

  function closeArticleModal() {
    ui.articleModal.style.display = "none";
  }

  async function saveArticle() {
    const title = ui.artTitleInput.value.trim();
    const content = ui.artContentInput.value;
    const categoryId = ui.artCategorySelect.value;
    const status = ui.artStatusSelect.value;
    const tags = ui.artTagsInput.value
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);

    if (!title || !categoryId) {
      alert("Заполните заголовок и выберите категорию");
      return;
    }

    try {
      const article = await api("/articles", {
        method: "POST",
        body: { title, content, category_id: categoryId, status, tags },
      });
      closeArticleModal();
      window.location.href = `/dev/kb/article?id=${article.id}`;
    } catch (err) {
      alert("Ошибка: " + err.message);
    }
  }

  // ── Upload .md ────────────────────────────────────────────

  async function handleMdUpload(file) {
    if (!state.activeCategoryId && state.categories.length > 0) {
      alert("Выберите категорию для загрузки");
      return;
    }
    const categoryId = state.activeCategoryId || (state.categories[0]?.id || "");
    if (!categoryId) {
      alert("Создайте категорию перед загрузкой");
      return;
    }

    const form = new FormData();
    form.append("file", file);
    form.append("category_id", categoryId);
    form.append("status", "DRAFT");

    try {
      const article = await api("/articles/upload", { method: "POST", body: form });
      window.location.href = `/dev/kb/article?id=${article.id}`;
    } catch (err) {
      alert("Ошибка загрузки: " + err.message);
    }
  }

  // ── Event Handlers ────────────────────────────────────────

  let searchDebounce = null;

  function setupEvents() {
    // Category tree clicks
    ui.categoryTree.addEventListener("click", (e) => {
      const toggle = e.target.closest("[data-toggle]");
      if (toggle) {
        toggle.classList.toggle("is-expanded");
        const children = ui.categoryTree.querySelector(`[data-children="${toggle.dataset.toggle}"]`);
        if (children) children.style.display = children.style.display === "none" ? "" : "none";
        return;
      }
      const item = e.target.closest("[data-category-id]");
      if (item) {
        state.activeCategoryId = item.dataset.categoryId || null;
        state.page = 0;
        ui.breadcrumbCategory.textContent = item.dataset.categoryId
          ? state.categories.find((c) => c.id === item.dataset.categoryId)?.title || "Категория"
          : "Все статьи";
        renderCategoryTree();
        loadArticles();
      }
    });

    // Tag clicks
    ui.tagsBar.addEventListener("click", (e) => {
      const tag = e.target.closest("[data-tag]");
      if (!tag) return;
      const name = tag.dataset.tag;
      state.activeTag = state.activeTag === name ? null : name;
      state.page = 0;
      renderTags();
      loadArticles();
    });

    // Search
    ui.searchInput.addEventListener("input", () => {
      clearTimeout(searchDebounce);
      searchDebounce = setTimeout(() => {
        state.search = ui.searchInput.value.trim();
        state.page = 0;
        loadArticles();
      }, 300);
    });

    // Ctrl+K shortcut
    document.addEventListener("keydown", (e) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "k") {
        e.preventDefault();
        ui.searchInput.focus();
      }
    });

    // Pagination
    ui.prevPage.addEventListener("click", () => {
      if (state.page > 0) {
        state.page--;
        loadArticles();
      }
    });
    ui.nextPage.addEventListener("click", () => {
      const totalPages = Math.ceil(state.total / PAGE_SIZE);
      if (state.page < totalPages - 1) {
        state.page++;
        loadArticles();
      }
    });

    // Category modal
    ui.addCategoryBtn.addEventListener("click", () => openCategoryModal());
    ui.cancelCatBtn.addEventListener("click", closeCategoryModal);
    ui.saveCatBtn.addEventListener("click", saveCategory);
    ui.categoryModal.addEventListener("click", (e) => {
      if (e.target === ui.categoryModal) closeCategoryModal();
    });

    // Article modal
    ui.createArticleBtn.addEventListener("click", openArticleModal);
    ui.cancelArtBtn.addEventListener("click", closeArticleModal);
    ui.saveArtBtn.addEventListener("click", saveArticle);
    ui.articleModal.addEventListener("click", (e) => {
      if (e.target === ui.articleModal) closeArticleModal();
    });

    // Editor tabs
    document.querySelectorAll("[data-tab]").forEach((tab) => {
      tab.addEventListener("click", () => {
        const isWrite = tab.dataset.tab === "write";
        tab.closest(".kb-editor-tabs").querySelectorAll(".kb-editor-tab").forEach((t) => t.classList.toggle("is-active", t === tab));
        ui.artContentInput.style.display = isWrite ? "" : "none";
        ui.artPreview.style.display = isWrite ? "none" : "";
        if (!isWrite && window.marked) {
          ui.artPreview.innerHTML = marked.parse(ui.artContentInput.value || "");
        }
      });
    });

    // Upload .md file
    ui.uploadMdFile.addEventListener("change", (e) => {
      const file = e.target.files?.[0];
      if (file) handleMdUpload(file);
      e.target.value = "";
    });
  }

  // ── Init ──────────────────────────────────────────────────

  async function init() {
    // Determine editor status
    if (auth) {
      const profile = await auth.fetchCurrentProfile?.();
      state.isEditor = profile?.is_admin || profile?.is_professor || false;
    }

    // Show/hide editor controls
    if (state.isEditor) {
      ui.categoryActions.style.display = "";
      ui.editorActions.style.display = "flex";
    }

    setupEvents();
    await Promise.all([loadCategories(), loadTags()]);
    await loadArticles();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
