declare var snippet: {
    title: string;
    content: string;
}

declare var hljs: {
    highlightAll: () => void;
    initLineNumbersOnLoad: () => void;
}

declare var feather: {
    replace: () => void;
}

// Used inside header
import "./time-since";

const snippetElement = document.querySelector("#snippet-content");
snippetElement.textContent = snippet.content;

document.querySelector("#snippet-title").textContent = snippet.title;

hljs.highlightAll();
hljs.initLineNumbersOnLoad();

feather.replace();

document.addEventListener('keydown', function(event) {
    if ((event.ctrlKey || event.metaKey) && event.key === "a") { // Ctrl+A
        event.preventDefault();
    } else {
        return;
    }

    const range = document.createRange();
    range.selectNodeContents(snippetElement);

    const selection = window.getSelection();
    selection.removeAllRanges();
    selection.addRange(range);
});

document.querySelector("#copy-snippet-button").addEventListener("click", () => {
    navigator.clipboard.writeText(snippet.content).catch((error) => {
        alert("Failed to copy snippet to clipboard: " + error.message)
    });
});

document.querySelector("#download-snippet-button").addEventListener("click", () => {
    const element = document.createElement('a');
    element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(snippet.content));
    element.setAttribute('download', snippet.title);

    element.style.display = 'none';
    document.body.appendChild(element);

    element.click();

    document.body.removeChild(element);
});
