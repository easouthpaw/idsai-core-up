(() => {
  "use strict";

  const auth = window.IDSAIAuth;
  const API = "/v2/kb";

  function notify(message, kind = "info") {
    if (auth && typeof auth.showAlert === "function") {
      auth.showAlert(message, kind);
      return;
    }
    window.alert(String(message || ""));
  }

  function confirmAction(options) {
    if (auth && typeof auth.showConfirmDialog === "function") {
      return auth.showConfirmDialog(options);
    }
    return Promise.resolve(window.confirm(String((options && options.message) || "")));
  }

  const state = {
    isEditor: false,
    article: null,
    isEditMode: false,
  };

  const ui = {
    breadcrumbs: document.getElementById("articleBreadcrumbs"),
    title: document.getElementById("articleTitle"),
    lead: document.getElementById("articleLead"),
    readingTime: document.getElementById("articleReadingTime"),
    meta: document.getElementById("articleMeta"),
    tags: document.getElementById("articleTags"),
    actions: document.getElementById("articleActions"),
    content: document.getElementById("articleContent"),
    editor: document.getElementById("articleEditor"),
    editTitle: document.getElementById("editTitle"),
    editTags: document.getElementById("editTags"),
    editContent: document.getElementById("editContent"),
    editPreview: document.getElementById("editPreview"),
    editStatus: document.getElementById("editStatus"),
    editBtn: document.getElementById("editArticleBtn"),
    deleteBtn: document.getElementById("deleteArticleBtn"),
    saveEditBtn: document.getElementById("saveEditBtn"),
    cancelEditBtn: document.getElementById("cancelEditBtn"),
    tocList: document.getElementById("tocList"),
    tocSidebar: document.getElementById("tocSidebar"),
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

  function initials(name) {
    const parts = String(name || "").trim().split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return (name || "??").slice(0, 2).toUpperCase();
  }

  function readingTime(content) {
    const words = String(content || "")
      .replace(/[`#>*_[\]\(\)\-]/g, " ")
      .trim()
      .split(/\s+/)
      .filter(Boolean).length;
    return Math.max(1, Math.round(words / 180));
  }

  function extractLead(content) {
    const lines = String(content || "")
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean);

    const lead = lines.find((line) => !/^#+\s/.test(line) && !/^[>*\-`]/.test(line));
    return lead ? lead.slice(0, 180) : "Короткая статья из базы знаний IDSAI с заметками, примерами и быстрым входом в тему.";
  }

  async function apiFetch(path, opts = {}) {
    if (auth && typeof auth.requestJSON === "function") {
      const { resp, data } = await auth.requestJSON(API + path, opts);
      if (!resp.ok) {
        throw new Error(data && typeof data === "object" && data.error ? data.error : `HTTP ${resp.status}`);
      }
      return resp.status === 204 ? null : data;
    }

    const token = auth?.getToken?.();
    const headers = { ...(opts.headers || {}) };
    const nextOpts = { ...opts };
    if (token) headers["Authorization"] = `Bearer ${token}`;
    if (nextOpts.body && typeof nextOpts.body === "object") {
      headers["Content-Type"] = "application/json";
      nextOpts.body = JSON.stringify(nextOpts.body);
    }
    const res = await fetch(API + path, { ...nextOpts, headers });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    if (res.status === 204) return null;
    return res.json();
  }

  // ── Markdown Rendering ────────────────────────────────────

  function renderMarkdown(md) {
    if (!window.marked) return escapeHTML(md);

    marked.setOptions({
      gfm: true,
      breaks: true,
      highlight: function (code, lang) {
        if (window.hljs && lang && hljs.getLanguage(lang)) {
          try {
            return hljs.highlight(code, { language: lang }).value;
          } catch {}
        }
        if (window.hljs) {
          try {
            return hljs.highlightAuto(code).value;
          } catch {}
        }
        return escapeHTML(code);
      },
    });

    let html = marked.parse(md || "");

    // Add language labels to code blocks
    html = html.replace(/<pre><code class="language-(\w+)">/g, (match, lang) => {
      return `<pre><span class="lang-label">${lang}</span><code class="language-${lang}">`;
    });

    // Convert GitHub-style alerts: > [!TIP], > [!WARNING], etc.
    html = html.replace(
      /<blockquote>\s*<p>\[!(NOTE|TIP|WARNING|CAUTION|IMPORTANT)\]\s*<br\s*\/?>\s*/gi,
      (_, type) => {
        const t = type.toLowerCase();
        const icons = { note: "info", tip: "lightbulb", warning: "warning", caution: "error", important: "priority_high" };
        const labels = { note: "Примечание", tip: "Совет", warning: "Внимание", caution: "Осторожно", important: "Важно" };
        const toneMap = { note: "info", tip: "tip", warning: "warning", caution: "danger", important: "info" };
        return `<div class="kb-callout kb-callout--${toneMap[t]}"><div class="kb-callout__icon"><span class="material-symbols-outlined">${icons[t]}</span></div><div class="kb-callout__body"><strong>${labels[t]}</strong>`;
      }
    );
    html = html.replace(/<\/p>\s*<\/blockquote>/g, (match) => {
      // This is approximate — proper handling would require a parser
      return `</div></div>`;
    });

    return html;
  }

  // ── TOC Generation ────────────────────────────────────────

  function generateTOC() {
    const headings = ui.content.querySelectorAll("h1, h2, h3");
    if (headings.length === 0) {
      ui.tocSidebar.style.display = "none";
      return;
    }

    let html = "";
    headings.forEach((h, i) => {
      const id = `heading-${i}`;
      h.id = id;
      const level = h.tagName.toLowerCase();
      const extraClass = level === "h3" ? "kb-toc-link--h3" : "";
      html += `<a class="kb-toc-link ${extraClass}" href="#${id}">${escapeHTML(h.textContent)}</a>`;
    });

    ui.tocList.innerHTML = html;
    ui.tocSidebar.style.display = "";

    // Intersection observer for active TOC item
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            ui.tocList.querySelectorAll(".kb-toc-link").forEach((link) => link.classList.remove("is-active"));
            const link = ui.tocList.querySelector(`[href="#${entry.target.id}"]`);
            if (link) link.classList.add("is-active");
          }
        });
      },
      { rootMargin: "-80px 0px -60% 0px", threshold: 0.1 }
    );

    headings.forEach((h) => observer.observe(h));
  }

  // ── Render Article ────────────────────────────────────────

  function render() {
    const a = state.article;
    if (!a) return;

    document.title = `${a.title} — IDSAI База знаний`;

    // Breadcrumbs
    ui.breadcrumbs.innerHTML = `
      <a href="/dev/kb">База знаний</a>
      <span aria-hidden="true">/</span>
      <strong>${escapeHTML(a.title)}</strong>
    `;

    // Title
    ui.title.textContent = a.title;
    if (ui.lead) {
      ui.lead.textContent = extractLead(a.content);
    }
    if (ui.readingTime) {
      ui.readingTime.textContent = `Чтение ${readingTime(a.content)} мин`;
    }

    // Meta
    const avatarContent = a.author_avatar
      ? `<img src="${escapeHTML(a.author_avatar)}" alt="" />`
      : initials(a.author_name);

    ui.meta.innerHTML = `
      <div class="kb-article-header__author">
        <div class="kb-article-header__avatar">${avatarContent}</div>
        <span>${escapeHTML(a.author_name)}</span>
      </div>
      <span>Обновлено ${relativeTime(a.updated_at)}</span>
      ${a.status === "DRAFT" ? `<span class="kb-article-card__status kb-article-card__status--draft">Черновик</span>` : ""}
    `;

    // Tags
    ui.tags.innerHTML = (a.tags || [])
      .map((t) => `<span class="kb-article-card__tag">${escapeHTML(t)}</span>`)
      .join("");

    // Content
    ui.content.innerHTML = renderMarkdown(a.content);
    generateTOC();

    // Editor actions
    if (state.isEditor) {
      ui.actions.style.display = "";
    }
  }

  // ── Edit Mode ─────────────────────────────────────────────

  function enterEditMode() {
    state.isEditMode = true;
    const a = state.article;
    ui.editTitle.value = a.title;
    ui.editTags.value = (a.tags || []).join(", ");
    ui.editContent.value = a.content;
    ui.editStatus.value = a.status;
    ui.content.style.display = "none";
    ui.editor.style.display = "";
    ui.actions.style.display = "none";
    ui.editContent.focus();
  }

  function exitEditMode() {
    state.isEditMode = false;
    ui.content.style.display = "";
    ui.editor.style.display = "none";
    if (state.isEditor) ui.actions.style.display = "";
    ui.editPreview.style.display = "none";
    ui.editContent.style.display = "";
    // Reset tabs
    ui.editor.querySelectorAll(".kb-editor-tab").forEach((t) => {
      t.classList.toggle("is-active", t.dataset.tab === "write");
    });
  }

  async function saveEdit() {
    const title = ui.editTitle.value.trim();
    const content = ui.editContent.value;
    const status = ui.editStatus.value;
    const tags = ui.editTags.value
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);

    if (!title) {
      notify("Заголовок обязателен", "warning");
      return;
    }

    try {
      state.article = await apiFetch(`/articles/${state.article.id}`, {
        method: "PATCH",
        body: { title, content, status, tags },
      });
      exitEditMode();
      render();
    } catch (err) {
      notify("Ошибка: " + err.message, "error");
    }
  }

  async function deleteArticle() {
    if (!await confirmAction({
      title: "Удалить статью",
      message: "Это действие необратимо. Статья будет удалена без возможности восстановления.",
      confirmText: "Удалить статью",
      danger: true,
    })) return;
    try {
      await apiFetch(`/articles/${state.article.id}`, { method: "DELETE" });
      window.location.href = "/dev/kb";
    } catch (err) {
      notify("Ошибка: " + err.message, "error");
    }
  }

  // ── Events ────────────────────────────────────────────────

  function setupEvents() {
    ui.editBtn.addEventListener("click", enterEditMode);
    ui.cancelEditBtn.addEventListener("click", exitEditMode);
    ui.saveEditBtn.addEventListener("click", saveEdit);
    ui.deleteBtn.addEventListener("click", deleteArticle);

    // Editor tabs
    ui.editor.querySelectorAll("[data-tab]").forEach((tab) => {
      tab.addEventListener("click", () => {
        const isWrite = tab.dataset.tab === "write";
        ui.editor.querySelectorAll(".kb-editor-tab").forEach((t) => t.classList.toggle("is-active", t === tab));
        ui.editContent.style.display = isWrite ? "" : "none";
        ui.editPreview.style.display = isWrite ? "none" : "";
        if (!isWrite) {
          ui.editPreview.innerHTML = renderMarkdown(ui.editContent.value);
        }
      });
    });

    // Smooth scroll for TOC
    ui.tocList.addEventListener("click", (e) => {
      const link = e.target.closest(".kb-toc-link");
      if (!link) return;
      e.preventDefault();
      const target = document.querySelector(link.getAttribute("href"));
      if (target) {
        target.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    });
  }

  // ── Init ──────────────────────────────────────────────────

  async function init() {
    const params = new URLSearchParams(window.location.search);
    const articleId = params.get("id");

    if (!articleId) {
      window.location.href = "/dev/kb";
      return;
    }

    if (auth && typeof auth.ensureSession === "function") {
      const profile = await auth.ensureSession(undefined);
      if (!profile) return;
      state.isEditor = profile?.is_admin || profile?.is_professor || false;
    }

    try {
      state.article = await apiFetch(`/articles/${articleId}`);
    } catch (err) {
      ui.title.textContent = "Статья не найдена";
      if (ui.lead) {
        ui.lead.textContent = "Похоже, материал был удален или ссылка на него устарела.";
      }
      ui.content.innerHTML = `<p style="color:var(--kb-muted);">Не удалось загрузить статью. ${escapeHTML(err.message)}</p>`;
      return;
    }

    setupEvents();
    render();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
