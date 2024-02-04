const formatter = new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: true
});

window.customElements.define('date-time', class extends HTMLElement {
    connectedCallback() {
        const date = new Date(this.getAttribute('datetime'));
        this.textContent = formatter.format(date);
    }
})