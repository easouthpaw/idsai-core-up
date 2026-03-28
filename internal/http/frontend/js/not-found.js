(() => {
  const auth = window.IDSAIAuth;

  function workspaceTarget() {
    if (localStorage.getItem("idsai_is_admin") === "1") {
      return { href: "/dev/admin", label: "В админку" };
    }
    if (localStorage.getItem("idsai_is_professor") === "1") {
      return { href: "/dev/professor", label: "В кабинет преподавателя" };
    }
    return { href: "/dev/projects", label: "К проектам" };
  }

  function requestedPath() {
    const params = new URLSearchParams(window.location.search || "");
    const fromQuery = String(params.get("from") || "").trim();
    if (fromQuery) {
      return fromQuery;
    }
    const path = `${window.location.pathname || ""}${window.location.search || ""}`;
    return path || "/";
  }

  function init() {
    const pathEl = document.getElementById("missingPath");
    if (pathEl) {
      pathEl.textContent = requestedPath();
    }

    const workspaceLink = document.getElementById("workspaceLink");
    if (workspaceLink) {
      const target = workspaceTarget();
      workspaceLink.href = target.href;
      workspaceLink.textContent = target.label;
    }

    if (auth && typeof auth.fetchCurrentProfile === "function") {
      void auth.fetchCurrentProfile().then((profile) => {
        if (!profile || !workspaceLink) {
          return;
        }
        workspaceLink.href = auth.targetByProfile(profile);
        workspaceLink.textContent = profile.is_admin
          ? "В админку"
          : profile.is_professor
            ? "В кабинет преподавателя"
            : "К проектам";
      }).catch(() => {});
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
