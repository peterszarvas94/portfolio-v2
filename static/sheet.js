// Spec sheet toggle. Loaded as an ES module, so it runs once the document is parsed.
// When the sheet opens below the visible screen (narrow layouts), scroll to it and move
// focus there; the stylesheet's scroll-behavior keeps the scroll instant for reduced motion.

const toggle = document.querySelector(".sheet-toggle");
const sheet = document.getElementById("typesheet");

toggle?.addEventListener("click", () => {
	sheet.hidden = !sheet.hidden;
	toggle.setAttribute("aria-expanded", String(!sheet.hidden));
	if (!sheet.hidden && sheet.getBoundingClientRect().top > innerHeight) {
		sheet.scrollIntoView({ block: "start" });
		sheet.focus({ preventScroll: true });
	}
});
