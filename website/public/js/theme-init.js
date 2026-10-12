(() => {
  const stored = localStorage.getItem("theme");
  const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  const isLight = stored === "light" || (!stored && !prefersDark);
  if (isLight) {
    document.documentElement.classList.add("light");
  }
})();
