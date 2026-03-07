(() => {
  const pressables = Array.from(document.querySelectorAll("[data-pressable]"));
  const navTabs = Array.from(document.querySelectorAll(".site-nav .tab-link"));
  const processTabs = Array.from(document.querySelectorAll(".process-tab"));
  const codeTabs = Array.from(document.querySelectorAll(".code-tab"));

  function animatePress(el) {
    el.classList.remove("is-pressed");
    // Restart keyframe animation for repeated clicks.
    void el.offsetWidth;
    el.classList.add("is-pressed");
    window.setTimeout(() => el.classList.remove("is-pressed"), 260);
  }

  pressables.forEach((el) => {
    el.addEventListener("click", () => animatePress(el));
  });

  navTabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      navTabs.forEach((item) => item.classList.remove("active"));
      tab.classList.add("active");
    });
  });

  processTabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      processTabs.forEach((item) => item.classList.remove("active"));
      tab.classList.add("active");
    });
  });

  codeTabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      codeTabs.forEach((item) => item.classList.remove("active"));
      tab.classList.add("active");
    });
  });
})();
