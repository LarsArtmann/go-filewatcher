if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
  for (const el of document.querySelectorAll("[data-animate]")) {
    el.classList.add("animate-fade-in");
  }
} else {
  const observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          entry.target.classList.add("animate-fade-in");
          observer.unobserve(entry.target);
        }
      }
    },
    { threshold: 0.1 },
  );
  for (const el of document.querySelectorAll("[data-animate]")) {
    observer.observe(el);
  }
}
