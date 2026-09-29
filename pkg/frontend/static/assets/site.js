// Examples use the same origin as the site, including local development ports.
document.querySelectorAll("[data-origin]").forEach(function (element) {
    element.textContent = window.location.origin;
});
