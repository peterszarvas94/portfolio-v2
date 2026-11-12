// Spec sheet toggle. Loaded as an ES module, so it runs once the document is parsed.
// There are two toggles: one in the top bar (wide screens, where the sheet opens beside
// the page) and one under the page (narrow screens, where the sheet opens below it).
// The stylesheet shows one at a time; both stay in sync. If the sheet still opens below
// the visible screen, scroll to it and move focus there; the stylesheet's scroll-behavior
// keeps the scroll instant for reduced motion.

const toggles = document.querySelectorAll(".sheet-toggle");
const sheet = document.getElementById("typesheet");

for (const toggle of toggles) {
	toggle.addEventListener("click", () => {
		sheet.hidden = !sheet.hidden;
		for (const t of toggles) t.setAttribute("aria-expanded", String(!sheet.hidden));
		if (!sheet.hidden && sheet.getBoundingClientRect().top > innerHeight) {
			sheet.scrollIntoView({ block: "start" });
			sheet.focus({ preventScroll: true });
		}
	});
}
